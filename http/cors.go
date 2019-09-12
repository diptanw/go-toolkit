package http

import "net/http"

// CORS is a middleware func that allows cross-origin requests.
func CORS(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Expose-Headers", "Location")

		if r.Method == "OPTIONS" {
			return
		}

		h.ServeHTTP(w, r)
	}
}
