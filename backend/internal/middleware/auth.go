package middleware

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strings"
)

// BasicAuth middleware protects routes using HTTP Basic Authentication
func BasicAuth(next http.Handler) http.Handler {
	creds := os.Getenv("ADMIN_CREDENTIALS")

	var expectedUser, expectedPass string
	if creds != "" {
		parts := strings.SplitN(creds, ":", 2)
		if len(parts) == 2 {
			expectedUser = parts[0]
			expectedPass = parts[1]
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Bypass authentication for login route
		if r.URL.Path == "/api/v1/auth/login" {
			next.ServeHTTP(w, r)
			return
		}

		// Only protect /api/ routes
		if !strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}

		user, pass, ok := r.BasicAuth()
		if !ok {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Deny all access if ADMIN_CREDENTIALS is not configured
		if creds == "" || expectedUser == "" {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized: Authentication Configuration Missing", http.StatusUnauthorized)
			return
		}

		if subtle.ConstantTimeCompare([]byte(user), []byte(expectedUser)) == 1 &&
			subtle.ConstantTimeCompare([]byte(pass), []byte(expectedPass)) == 1 {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}
