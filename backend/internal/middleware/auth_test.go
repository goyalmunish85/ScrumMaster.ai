package middleware

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestBasicAuth(t *testing.T) {
	// Setup mock handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	tests := []struct {
		name           string
		method         string
		path           string
		authHeader     string
		envCreds       string
		expectedStatus int
	}{
		{
			name:           "Login route is excluded",
			method:         "POST",
			path:           "/api/v1/auth/login",
			authHeader:     "",
			envCreds:       "admin:admin",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Non-API route is excluded",
			method:         "GET",
			path:           "/healthz",
			authHeader:     "",
			envCreds:       "admin:admin",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "API route without auth fails",
			method:         "GET",
			path:           "/api/v1/tasks",
			authHeader:     "",
			envCreds:       "admin:admin",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "API route with wrong auth fails",
			method:         "GET",
			path:           "/api/v1/tasks",
			authHeader:     "Basic " + base64.StdEncoding.EncodeToString([]byte("wrong:wrong")),
			envCreds:       "admin:admin",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "API route fails if credentials missing",
			method:         "GET",
			path:           "/api/v1/tasks",
			authHeader:     "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:admin")),
			envCreds:       "", // Not configured
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "API route fails if credentials malformed",
			method:         "GET",
			path:           "/api/v1/tasks",
			authHeader:     "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:admin")),
			envCreds:       "admin", // Malformed
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "API route with correct custom auth succeeds",
			method:         "GET",
			path:           "/api/v1/tasks",
			authHeader:     "Basic " + base64.StdEncoding.EncodeToString([]byte("custom:pass")),
			envCreds:       "custom:pass",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envCreds != "" {
				os.Setenv("ADMIN_CREDENTIALS", tt.envCreds)
				defer os.Unsetenv("ADMIN_CREDENTIALS")
			} else {
				os.Unsetenv("ADMIN_CREDENTIALS")
			}

			// Initialize middleware AFTER setting the env var
			handlerToTest := BasicAuth(nextHandler)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rr := httptest.NewRecorder()
			handlerToTest.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}
		})
	}
}
