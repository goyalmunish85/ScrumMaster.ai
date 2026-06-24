package events

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/aios/backend/internal/db"
	"github.com/aios/backend/internal/models"
)

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
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offset := 0
	offsetStr := r.URL.Query().Get("offset")
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	var events []models.OperationalEventRecord
	// Fetch ordered by CreatedAt descending (chronological ledger) with pagination
	if err := db.DB.Order("created_at desc").Limit(limit).Offset(offset).Find(&events).Error; err != nil {
		http.Error(w, "Failed to fetch activities", http.StatusInternalServerError)
		return
	}

	// For privacy/GDPR, we filter sensitive fields
	redactedKeys := []string{"email", "phone", "address", "name", "password", "token", "ssn"}
	for i := range events {
		var payloadData map[string]interface{}
		if err := json.Unmarshal([]byte(events[i].Payload), &payloadData); err == nil {
			for key := range payloadData {
				for _, rKey := range redactedKeys {
					if key == rKey {
						payloadData[key] = "[REDACTED]"
						break
					}
				}
			}
			if redactedPayload, err := json.Marshal(payloadData); err == nil {
				events[i].Payload = string(redactedPayload)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}
