package retry

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
)

// NewHTTPClient returns a new http.Client with the http.DefaultTransport
// wrapped with retrialable round tripper.
func NewHTTPClient(policy Policy) *http.Client {
	return WithPolicy(http.DefaultClient, policy)
}

// WithPolicy returns a copy of http.Client with retrialable transport.
// If Check func in not specified in the given policy, default HTTPCheck
// policy will be used.
func WithPolicy(client *http.Client, policy Policy) *http.Client {
	if policy.Check == nil {
		policy.Check = HTTPCheck
	}

	cp := *client
	if cp.Transport == nil {
		cp.Transport = http.DefaultTransport
	}

	cp.Transport = transport{
		Policy:       policy,
		RoundTripper: cp.Transport,
	}

	return &cp
}

// HTTPCheck is the policy callback, which will retry on connection
// errors and certain server responses.
func HTTPCheck(err error, res interface{}) (bool, error) {
	if err != nil {
		// Untyped error returned by net/http when the number of redirects is exhausted.
		if ok, _ := regexp.MatchString(`stopped after \d+ redirects\z`, err.Error()); ok {
			return false, nil
		}

		// Untyped error returned by net/http when the scheme specified in the URL is invalid.
		if ok, _ := regexp.MatchString(`unsupported protocol scheme`, err.Error()); ok {
			return false, nil
		}

		// Error was due to TLS cert verification failure.
		if errors.Is(err, x509.UnknownAuthorityError{}) {
			return false, nil
		}

		if err == context.Canceled || err == context.DeadlineExceeded {
			return false, nil
		}

		return true, nil
	}

	// Verify the server response.
	if resp, ok := res.(*http.Response); ok {
		if resp.StatusCode >= 500 && resp.StatusCode != 501 {
			// Maybe a temporary outage.
			return true, nil
		}

		switch resp.StatusCode {
		case 0, http.StatusRequestTimeout, http.StatusTooManyRequests:
			return true, nil
		}
	}

	// No need for another attempt.
	return false, nil
}

// transport is a wrapper for HTTP RoundTripper that is responsible for retrying.
type transport struct {
	RoundTripper http.RoundTripper
	Policy       Policy
}

// RoundTrip wraps calling an HTTP Transport with retries.
func (t transport) RoundTrip(req *http.Request) (resp *http.Response, reqErr error) {
	defer t.CloseIdleConnections()

	count, doErr := t.Policy.Do(req.Context(), func(retrying bool) (interface{}, error) {
		if retrying {
			if reqErr == nil && resp != nil {
				// Read the response body to reuse keep-alive connection.
				// Read more https://gist.github.com/mholt/eba0f2cc96658be0f717
				io.Copy(ioutil.Discard, io.LimitReader(resp.Body, 4096)) // nolint
				resp.Body.Close()
			}

			if req.ContentLength > 0 {
				// GetBody should be set to enable the request body rewind,
				// http.NewRequest automatically sets it for common types,
				// otherwise it is unable to rewind.
				if req.GetBody == nil {
					return nil, errors.New("request.GetBody is nil")
				}

				// RoundTripper shouldn't modify request, except for reading and closing the body.
				if req.Body, reqErr = req.GetBody(); reqErr != nil {
					return nil, fmt.Errorf("rewinding body: %w", reqErr)
				}
			}
		}

		resp, reqErr = t.RoundTripper.RoundTrip(req) // nolint:bodyclose // should be closed by client
		return resp, reqErr
	})

	if doErr != nil {
		return nil, &requestError{
			count: count + 1,
			err:   doErr,
		}
	}

	return resp, nil
}

// CloseIdleConnections closes all idle connections if internal transport
// supports it.
func (t transport) CloseIdleConnections() {
	if tr, ok := t.RoundTripper.(interface{ CloseIdleConnections() }); ok {
		tr.CloseIdleConnections()
	}
}

type requestError struct {
	count int
	err   error
}

func (e *requestError) Error() string {
	return fmt.Sprintf("request attempt %d: %s", e.count, e.err.Error())
}

func (e *requestError) Is(target error) bool {
	return e.err == target
}
