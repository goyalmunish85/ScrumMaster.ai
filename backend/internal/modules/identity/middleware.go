package identity

import (
	"encoding/base64"
	"net/http"
	"os"
	"strings"
)

// AuthMiddleware intercepts incoming HTTP requests to ensure they contain valid authorization.
// It skips authentication for the /api/v1/auth/login route.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Bypass authentication for the login route
		if r.URL.Path == "/api/v1/auth/login" {
			next.ServeHTTP(w, r)
			return
		}

		// Require Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized: Missing Authorization header", http.StatusUnauthorized)
			return
		}

		// Support Basic Auth or Bearer Token (JWT mock)
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 {
			http.Error(w, "Unauthorized: Invalid Authorization header format", http.StatusUnauthorized)
			return
		}

		authType := parts[0]
		credentials := parts[1]

		if authType == "Basic" {
			expectedCredentials := os.Getenv("ADMIN_CREDENTIALS")
			if expectedCredentials == "" {
				expectedCredentials = "admin:admin"
			}

			decoded, err := base64.StdEncoding.DecodeString(credentials)
			if err != nil || string(decoded) != expectedCredentials {
				http.Error(w, "Unauthorized: Invalid Basic credentials", http.StatusUnauthorized)
				return
			}
		} else {
			http.Error(w, "Unauthorized: Unsupported Authorization type", http.StatusUnauthorized)
			return
		}

		// Proceed to the actual handler
		next.ServeHTTP(w, r)
	})
}
