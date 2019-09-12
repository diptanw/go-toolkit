package http

import (
	"net/http"
	"regexp"
	"strings"
)

// MiddlewareFunc is a func type for middleware http handler.
type MiddlewareFunc func(http.HandlerFunc) http.HandlerFunc

// Mux is a simple routes multiplexer.
type Mux struct {
	routes     []route
	middleware []MiddlewareFunc
}

type route struct {
	method  string
	regex   *regexp.Regexp
	params  map[int]string
	handler http.HandlerFunc
}

// WithMiddleware adds a middleware wrapper for the root handler.
func (m *Mux) WithMiddleware(mw ...MiddlewareFunc) {
	m.middleware = append(m.middleware, mw...)
}

// AddRoute adds a new route to the Handler.
func (m *Mux) AddRoute(method, pattern string, handler http.HandlerFunc) {
	parts := strings.Split(pattern, "/")

	// find parts that starts with ":" and replace with regex
	j, params := 0, make(map[int]string)

	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			expr := "([^/]+)"
			params[j] = part
			parts[i] = expr
			j++
		}
	}

	pattern = strings.Join(parts, "/")
	regex := regexp.MustCompile(pattern)

	m.routes = append(m.routes, route{
		method:  method,
		regex:   regex,
		handler: handler,
		params:  params,
	})
}

// ServeHTTP handles all page routing.
func (m *Mux) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	f := func(rw http.ResponseWriter, r *http.Request) {
		if ok, route := m.find(r.Method, r.URL.Path); ok {
			route.handler(rw, r)
			return
		}

		http.NotFound(rw, r)
	}

	wrap(f, m.middleware)(rw, r)
}

func (m *Mux) find(method, path string) (bool, route) {
	for _, route := range m.routes {
		if method != route.method {
			continue
		}

		requestPath := strings.TrimPrefix(path, "/")
		if !route.regex.MatchString(requestPath) {
			continue
		}

		matches := route.regex.FindStringSubmatch(requestPath)

		// check that the Route matches the URL pattern.
		if len(matches[0]) != len(requestPath) {
			continue
		}

		return true, route
	}

	return false, route{}
}

// GetParams returns a map of params and it's values.
func (m *Mux) GetParams(method, path string) map[string]string {
	if ok, route := m.find(method, path); ok {
		if len(route.params) > 0 {
			matches := route.regex.FindStringSubmatch(path)
			values := make(map[string]string)

			for i, match := range matches[1:] {
				param := route.params[i][1:]
				values[param] = match
			}

			return values
		}
	}

	return nil
}

func wrap(h http.HandlerFunc, mw []MiddlewareFunc) http.HandlerFunc {
	for _, m := range mw {
		h = m(h)
	}

	return h
}
