package search

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/aios/backend/internal/modules/events"
	"github.com/aios/backend/internal/modules/memory"
	"github.com/aios/backend/internal/utils/gdpr"
)

// SearchEventsHandler searches events semantically based on a query parameter 'q'
func SearchEventsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Missing 'q' query parameter", http.StatusBadRequest)
		return
	}

	limit := uint64(50)
	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		parsedLimit, err := strconv.ParseUint(limitStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
			return
		}
		if parsedLimit < 1 {
			http.Error(w, "Limit must be greater than 0", http.StatusBadRequest)
			return
		}
		if parsedLimit > 100 {
			limit = 100
		} else {
			limit = parsedLimit
		}
	}

	matchedEvents, err := memory.SearchEventsSemantic(query, limit)
	if err != nil {
		http.Error(w, "Failed to search events", http.StatusInternalServerError)
		return
	}

	response := []events.ActivityResponse{}
	for _, event := range matchedEvents {
		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(event.Payload), &payload); err != nil {
			// If payload is invalid JSON, skip it or return empty payload
			payload = make(map[string]interface{})
		}

		// GDPR Compliance: Scrub PII fields using the shared sanitize utility
		sanitizedPayloadObj := gdpr.SanitizePayload(payload)
		sanitizedPayload, ok := sanitizedPayloadObj.(map[string]interface{})
		if !ok {
			sanitizedPayload = make(map[string]interface{})
		}

		response = append(response, events.ActivityResponse{
			ID:        event.ID,
			TaskID:    event.TaskID,
			EventType: event.EventType,
			Payload:   sanitizedPayload,
			CreatedAt: event.CreatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
