package settings

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/aios/backend/internal/db"
	"github.com/aios/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SlackConfigRequest struct {
	SlackChannels []string `json:"slack_channels"`
}

func UpdateSlackConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType != "" && contentType != "application/json" {
		http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
		return
	}

	var req SlackConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if req.SlackChannels == nil {
		http.Error(w, "slack_channels is required", http.StatusBadRequest)
		return
	}

	for _, ch := range req.SlackChannels {
		if ch == "" {
			http.Error(w, "slack_channels cannot contain empty strings", http.StatusBadRequest)
			return
		}
	}

	if db.DB == nil {
		http.Error(w, "Database not initialized", http.StatusInternalServerError)
		return
	}

	err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("platform = ?", "slack").Delete(&models.IntegrationTarget{}).Error; err != nil {
			return err
		}

		if len(req.SlackChannels) > 0 {
			now := time.Now()
			targets := make([]models.IntegrationTarget, 0, len(req.SlackChannels))
			for _, ch := range req.SlackChannels {
				targets = append(targets, models.IntegrationTarget{
					ID:        uuid.New().String(),
					Platform:  "slack",
					TargetID:  ch,
					CreatedAt: now,
				})
			}
			if err := tx.Create(&targets).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		http.Error(w, "Failed to update slack configuration in database", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Slack configuration updated successfully",
		"config": map[string]interface{}{
			"slack_channels": req.SlackChannels,
		},
	})
}
