package oauth2

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"

	httpkit "github.com/diptanw/go-toolkit/http"
)

func ExampleAuthenticator() {
	authSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body != nil {
			body, _ := ioutil.ReadAll(r.Body)
			if strings.Contains(string(body), "client_id=test-id&client_secret=test-secret") {
				fmt.Println("auth: credentials granted")
				io.WriteString(w, "access_token=test")
			}
		}
	}))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "Bearer test" {
			fmt.Println("server: request authorized")
		}
	}))

	config := Config{
		ClientID:     "test-id",
		ClientSecret: "test-secret",
		TokenURL:     authSrv.URL,
	}

	authenticator := NewAuthenticator(context.Background(), config, "resources")
	authClient := httpkit.WithAuthenticator(srv.Client(), authenticator)

	authClient.Get(srv.URL + "/resources")

	// Output:
	// auth: credentials granted
	// server: request authorized
}
