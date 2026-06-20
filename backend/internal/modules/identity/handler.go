package identity

import (
	"encoding/json"
	"net/http"
	"time"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"` // MVP only, will replace with proper auth later
}

type LoginResponse struct {
	Token     string `json:"token"`
	Workspace string `json:"workspace"`
}

// LoginHandler handles basic JWT-style login (mocked for Phase 1/2)
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Mock Authentication for MVP
	// In production, verify against DB and sign a real JWT
	resp := LoginResponse{
		Token:     "mock-jwt-token-for-" + req.Email + "-" + time.Now().String(),
		Workspace: "default-workspace-id",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
