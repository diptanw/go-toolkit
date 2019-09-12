package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithAuthenticator(t *testing.T) {
	var auth fakeAuthenticator = func(r *http.Request) error {
		return nil
	}

	client := http.DefaultClient
	authClient := WithAuthenticator(client, auth)

	require.NotEqual(t, client, authClient)
	require.NotNil(t, authClient)
	require.NotNil(t, authClient.Transport)
	require.IsType(t, transport{}, authClient.Transport)
}

func TestTransport_RoundTrip(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.Header["Authorization"], "test")
	}))

	client := srv.Client()

	var auth fakeAuthenticator = func(r *http.Request) error {
		r.Header.Set("Authorization", "test")
		return nil
	}

	authClient := WithAuthenticator(client, auth)
	resp, err := authClient.Get(srv.URL)

	require.NoError(t, err)
	require.Contains(t, resp.Request.Header.Get("Authorization"), "test")
}

func TestTransport_RoundTrip_error(t *testing.T) {
	client := http.DefaultClient

	var auth fakeAuthenticator = func(r *http.Request) error {
		return assert.AnError
	}

	authClient := WithAuthenticator(client, auth)
	_, err := authClient.Get("http://localhost:8080")

	require.Error(t, err)
	require.True(t, errors.Is(err, assert.AnError))
}

func TestWithDummy(t *testing.T) {
	client := http.DefaultClient
	authClient := WithDummy(client)

	require.NotEqual(t, client, authClient)
	require.NotNil(t, authClient)
	require.NotNil(t, authClient.Transport)

	req := httptest.NewRequest("GET", "http://localhost", nil)

	require.NoError(t, dummyBearer{}.Authenticate(req))
	require.Equal(t, "Bearer dummy", req.Header.Get("Authorization"))
}

type fakeAuthenticator func(*http.Request) error

func (a fakeAuthenticator) Authenticate(r *http.Request) error {
	return a(r)
}
