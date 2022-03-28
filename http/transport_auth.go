package http

import (
	"fmt"
	"net/http"
)

// Authenticator defines the request authentication function.
type Authenticator interface {
	Authenticate(*http.Request) error
}

// WithAuthenticator returns a copy of http.Client with authentication transport.
// Should be set in the foremost end of the round-tripper chain.
func WithAuthenticator(client *http.Client, auth Authenticator) *http.Client {
	cp := *client
	if cp.Transport == nil {
		cp.Transport = http.DefaultTransport
	}

	cp.Transport = transport{
		auth: auth,
		next: cp.Transport,
	}

	return &cp
}

type transport struct {
	auth Authenticator
	next http.RoundTripper
}

func (t transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := t.auth.Authenticate(req); err != nil {
		return nil, fmt.Errorf("authenticate request: %w", err)
	}

	return t.next.RoundTrip(req)
}

// WithDummy returns a copy of http.Client with dummy authentication transport.
func WithDummy(client *http.Client) *http.Client {
	return WithAuthenticator(client, dummyBearer{})
}

type dummyBearer struct{}

func (d dummyBearer) Authenticate(r *http.Request) error {
	r.Header.Set("Authorization", "Bearer dummy")
	return nil
}
