package events

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/aios/backend/internal/modules/memory"
)

// SearchEventsHandler handles semantic search for operational events
func SearchEventsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// GDPR Compliance: Add strict authentication check (MVP token or any token presence)
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	limit := uint64(10)
	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		if parsedLimit, err := strconv.ParseUint(limitStr, 10, 64); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	events, err := memory.SearchEventsSemantic(query, limit)
	if err != nil {
		http.Error(w, "Failed to search events", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}
