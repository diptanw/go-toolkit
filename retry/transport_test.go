package retry

import (
	"bytes"
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ExampleNewHTTPClient() {
	var reqNum int

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code := http.StatusRequestTimeout
		if reqNum >= 2 {
			code = http.StatusOK
		}

		reqNum++
		w.WriteHeader(code)
		fmt.Printf("Server response %d: %d\n", reqNum, code)
	}))

	defer ts.Close()

	client := NewHTTPClient(Policy{RetryMax: 3})

	res, err := client.Get(ts.URL)
	if err != nil {
		log.Fatal(err) // nolint:gocritic
	}

	defer res.Body.Close()

	fmt.Printf("Status: %s", res.Status)

	// Output:
	// Server response 1: 408
	// Server response 2: 408
	// Server response 3: 200
	// Status: 200 OK
}

func TestNewHTTPClient(t *testing.T) {
	client := NewHTTPClient(Policy{})

	require.NotNil(t, client)
	require.NotNil(t, client.Transport)
	require.IsType(t, transport{}, client.Transport)
}

func TestWithPolicy(t *testing.T) {
	client := http.Client{}
	retryClient := WithPolicy(&client, Policy{})

	require.NotNil(t, retryClient)
	require.NotEqual(t, client, retryClient)

	tr := retryClient.Transport.(transport)

	require.NotNil(t, tr.Policy)
	require.NotNil(t, tr.Policy.Check)
}

func TestHTTPCheck(t *testing.T) {
	tests := map[string]struct {
		giveResponse interface{}
		giveErr      error
		wantRetry    bool
	}{
		"retry if status TooManyRequests": {
			&http.Response{StatusCode: http.StatusTooManyRequests},
			nil,
			true,
		},
		"retry if status RequestTimedout": {
			&http.Response{StatusCode: http.StatusRequestTimeout},
			nil,
			true,
		},
		"retry if status greater or equal 500": {
			&http.Response{StatusCode: http.StatusServiceUnavailable},
			nil,
			true,
		},
		"retry if error": {
			nil,
			assert.AnError,
			true,
		},
		"do not retry if status OK": {
			&http.Response{StatusCode: http.StatusOK},
			nil,
			false,
		},
		"do not retry if protocol scheme error": {
			nil,
			errors.New("unsupported protocol scheme"),
			false,
		},
		"do not retry if too many redirects error": {
			nil,
			errors.New("stopped after 10 redirects"),
			false,
		},
		"do not retry if unknown authority error": {
			nil,
			x509.UnknownAuthorityError{},
			false,
		},
		"do not retry if context canceled error": {
			nil,
			context.Canceled,
			false,
		},
		"do not retry if deadline exceeded error": {
			nil,
			context.DeadlineExceeded,
			false,
		},
		"do not retry if status NotImplemented": {
			&http.Response{StatusCode: http.StatusNotImplemented},
			nil,
			false,
		},
		"do not retry if no error and no response": {
			nil,
			nil,
			false,
		},
	}

	for name, test := range tests {
		tc := test

		t.Run(name, func(t *testing.T) {
			retry, _ := HTTPCheck(tc.giveErr, tc.giveResponse)
			assert.Equal(t, tc.wantRetry, retry)
		})
	}
}

func TestTransport_RoundTrip_CheckPolicy(t *testing.T) {
	tests := map[string]struct {
		givePolicy  Policy
		giveInnerRT *fakeRoundTripper
		giveRequest *http.Request
		wantErr     error
	}{
		"success": {
			Policy{},
			&fakeRoundTripper{retResp: &http.Response{}},
			&http.Request{},
			nil,
		},
		"do to retry on unrecoverable error": {
			Policy{
				RetryMax: 2,
				Check: func(err error, res interface{}) (bool, error) {
					return err != assert.AnError, nil
				},
			},
			&fakeRoundTripper{retErrs: []error{assert.AnError}},
			&http.Request{},
			&requestError{1, assert.AnError},
		},
		"retry on recoverable error": {
			Policy{
				RetryMax: 2,
				Check: func(err error, res interface{}) (bool, error) {
					return err == assert.AnError, nil
				},
			},
			&fakeRoundTripper{retErrs: []error{assert.AnError}},
			&http.Request{},
			&requestError{3, assert.AnError},
		},
		"recovered with success": {
			Policy{
				RetryMax: 2,
				Check: func(err error, res interface{}) (bool, error) {
					return err == assert.AnError, nil
				},
			},
			&fakeRoundTripper{retErrs: []error{assert.AnError, nil}},
			&http.Request{},
			nil,
		},
		"recovered with another error": {
			Policy{
				RetryMax: 2,
				Check: func(err error, res interface{}) (bool, error) {
					return err == assert.AnError, nil
				},
			},
			&fakeRoundTripper{retErrs: []error{assert.AnError, errors.New("error")}},
			&http.Request{},
			&requestError{2, errors.New("error")},
		},
	}

	for name, test := range tests {
		tc := test

		t.Run(name, func(t *testing.T) {
			tr := transport{
				Policy:       tc.givePolicy,
				RoundTripper: tc.giveInnerRT,
			}

			resp, err := tr.RoundTrip(tc.giveRequest)
			if resp != nil && resp.Body != nil {
				resp.Body.Close()
			}

			assert.Equal(t, tc.wantErr, err)
			assert.True(t, tc.giveInnerRT.isClosed)
		})
	}
}

func TestTransport_RoundTrip_ResponseDrained(t *testing.T) {
	rc := fakeReadCloser{Buffer: bytes.NewBufferString("test")}
	tr := transport{
		Policy: Policy{
			RetryMax: 1,
			Check: func(err error, res interface{}) (bool, error) {
				return true, nil
			},
		},
		RoundTripper: &fakeRoundTripper{
			retResp: &http.Response{
				Body: &rc,
			},
		},
	}

	resp, _ := tr.RoundTrip(&http.Request{})
	resp.Body.Close()

	assert.Equal(t, 4, rc.readBytes)
}

func TestTransport_RoundTrip_Retrying(t *testing.T) {
	tests := map[string]struct {
		giveRequest *http.Request
		wantErr     error
	}{
		"request GetBody is nil error": {
			&http.Request{
				ContentLength: 1,
			},
			&requestError{2, errors.New("request.GetBody is nil")},
		},
		"rewind body error": {
			&http.Request{
				ContentLength: 1,
				GetBody: func() (io.ReadCloser, error) {
					return nil, errors.New("error")
				},
			},
			&requestError{2, fmt.Errorf("rewinding body: %w", errors.New("error"))},
		},
		"rewind body OK": {
			&http.Request{
				ContentLength: 1,
				GetBody: func() (io.ReadCloser, error) {
					return nil, nil
				},
			},
			&requestError{2, assert.AnError},
		},
	}

	for name, test := range tests {
		tc := test

		t.Run(name, func(t *testing.T) {
			tr := transport{
				Policy: Policy{
					RetryMax: 1,
					Check: func(err error, res interface{}) (bool, error) {
						return true, nil
					},
				},
				RoundTripper: &fakeRoundTripper{
					retErrs: []error{assert.AnError},
				},
			}

			resp, err := tr.RoundTrip(tc.giveRequest)
			if resp != nil && resp.Body != nil {
				resp.Body.Close()
			}

			assert.Equal(t, tc.wantErr, err)
		})
	}
}

func TestTransport_CloseIdleConnections(t *testing.T) {
	mockRt := fakeRoundTripper{}
	tr := transport{
		RoundTripper: &mockRt,
	}

	tr.CloseIdleConnections()

	assert.True(t, mockRt.isClosed)
}

func TestRequestError_Error(t *testing.T) {
	err := &requestError{
		count: 10,
		err:   errors.New("error"),
	}

	assert.Error(t, err)
	assert.EqualError(t, err, "request attempt 10: error")
}

func TestRequestError_Is(t *testing.T) {
	err := &requestError{
		err: assert.AnError,
	}

	assert.True(t, errors.Is(err, assert.AnError))
}

type fakeRoundTripper struct {
	isClosed bool
	trips    int
	retResp  *http.Response
	retErrs  []error
}

func (rt *fakeRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	rt.trips++

	if l := len(rt.retErrs); l > 0 {
		return rt.retResp, rt.retErrs[(rt.trips-1)%l]
	}

	return rt.retResp, nil
}

func (rt *fakeRoundTripper) CloseIdleConnections() {
	rt.isClosed = true
}

type fakeReadCloser struct {
	*bytes.Buffer
	readBytes int
}

func (r *fakeReadCloser) Read(p []byte) (int, error) {
	n, err := r.Buffer.Read(p)
	r.readBytes += n

	return n, err
}

func (r *fakeReadCloser) Close() error {
	return nil
}
