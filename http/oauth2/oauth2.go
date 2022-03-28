package oauth2

import (
	"context"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// Authenticator is a type that handles the http authentication toward oauth2 oidc provider.
type Authenticator struct {
	src oauth2.TokenSource
}

// Config is a struct with authentication options used for token generation.
type Config struct {
	ClientID     string
	ClientSecret string
	TokenURL     string
	Audience     string
}

// NewAuthenticator returns a new instance of oauth2 http request authenticator.
func NewAuthenticator(ctx context.Context, cfg Config, scopes ...string) *Authenticator {
	c := clientcredentials.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		TokenURL:     cfg.TokenURL,
	}

	if cfg.Audience != "" {
		c.EndpointParams = url.Values{}
		c.EndpointParams.Set("audience", cfg.Audience)
	}

	c.Scopes = append(c.Scopes, scopes...)

	return &Authenticator{
		src: c.TokenSource(ctx),
	}
}

// Authenticate sets the token into authorization header for the given request.
func (a *Authenticator) Authenticate(r *http.Request) error {
	t, err := a.src.Token()
	if err != nil {
		return err
	}

	t.SetAuthHeader(r)

	return nil
}
