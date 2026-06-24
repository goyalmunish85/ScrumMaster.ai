package events

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/aios/backend/internal/db"
	"github.com/aios/backend/internal/models"
)

type ActivityResponse struct {
	ID        string                 `json:"id"`
	TaskID    *string                `json:"task_id"`
	EventType string                 `json:"event_type"`
	Payload   map[string]interface{} `json:"payload"`
	CreatedAt time.Time              `json:"created_at"`
}

func GetActivitiesHandler(w http.ResponseWriter, r *http.Request) {
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

	limit := 50
	limitStr := r.URL.Query().Get("limit")
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
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

	offset := 0
	offsetStr := r.URL.Query().Get("offset")
	if offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err != nil {
			http.Error(w, "Invalid offset parameter", http.StatusBadRequest)
			return
		}
		if parsedOffset < 0 {
			http.Error(w, "Offset cannot be negative", http.StatusBadRequest)
			return
		}
		offset = parsedOffset
	}

	var events []models.OperationalEventRecord
	// Fetch ordered by CreatedAt descending (chronological ledger) with pagination
	if err := db.DB.Order("created_at desc").Limit(limit).Offset(offset).Find(&events).Error; err != nil {
		http.Error(w, "Failed to fetch activities", http.StatusInternalServerError)
		return
	}

	response := []ActivityResponse{}
	for _, event := range events {
		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(event.Payload), &payload); err != nil {
			// If payload is invalid JSON, skip it or return empty payload
			payload = make(map[string]interface{})
		}

		// GDPR Compliance: Scrub PII fields
		delete(payload, "assignee")
		delete(payload, "reporter")

		response = append(response, ActivityResponse{
			ID:        event.ID,
			TaskID:    event.TaskID,
			EventType: event.EventType,
			Payload:   payload,
			CreatedAt: event.CreatedAt,
		})
	}

	// For privacy/GDPR, we could filter sensitive fields if required, but payload is just task info here
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
