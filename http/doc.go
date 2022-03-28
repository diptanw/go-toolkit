/*
Package http provides types for HTTP authentication. Supported oauth2
authenticator for authorizing HTTP requests towards the oidc providers.

Basic Usage

	import (
		...
		"github.com/diptanw/go-toolkit/auth"
		"github.com/diptanw/go-toolkit/auth/oauth2"
		"github.com/coreos/go-oidc"
	)

	...

	provider, err := oidc.NewProvider(context.Background(), "http://auth.remote.resource")
	if err != nil {
		return err
	}

	config := oauth2.Config{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		TokenURL:     provider.Endpoint().TokenURL,
	}

	authenticator := oauth2.NewAuthenticator(context.Background(), config, "profiles")
	authClient := auth.WithAuthenticator(http.DefaultClient, authenticator)

	resp, err := authClient.Get("http://remote.resource/profiles")
*/
package http
