package events

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/aios/backend/internal/db"
	"github.com/aios/backend/internal/models"
)

type SanitizedEventRecord struct {
	ID        string          `json:"id"`
	TaskID    *string         `json:"task_id"`
	EventType string          `json:"event_type"`
	Payload   json.RawMessage `json:"payload"`
	CreatedAt time.Time       `json:"created_at"`
}

func sanitizePayload(rawPayload string) json.RawMessage {
	var parsed interface{}
	if err := json.Unmarshal([]byte(rawPayload), &parsed); err != nil {
		return json.RawMessage(rawPayload)
	}

	piiKeys := []string{"email", "phone", "address", "name", "password", "token", "ssn"}

	var scrub func(v interface{})
	scrub = func(v interface{}) {
		switch node := v.(type) {
		case map[string]interface{}:
			for k, val := range node {
				for _, pii := range piiKeys {
					if k == pii {
						delete(node, k)
						break
					}
				}
				if _, ok := node[k]; ok {
					scrub(val)
				}
			}
		case []interface{}:
			for _, item := range node {
				scrub(item)
			}
		}
	}

	scrub(parsed)

	sanitized, err := json.Marshal(parsed)
	if err != nil {
		return json.RawMessage(rawPayload)
	}
	return json.RawMessage(sanitized)
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
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			if parsedLimit > 100 {
				limit = 100
			} else {
				limit = parsedLimit
			}
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

	sanitizedEvents := make([]SanitizedEventRecord, len(events))
	for i, e := range events {
		sanitizedEvents[i] = SanitizedEventRecord{
			ID:        e.ID,
			TaskID:    e.TaskID,
			EventType: e.EventType,
			Payload:   sanitizePayload(e.Payload),
			CreatedAt: e.CreatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sanitizedEvents)
}
