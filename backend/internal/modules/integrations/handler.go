package integrations

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/aios/backend/internal/db"
	"github.com/aios/backend/internal/models"
	"github.com/google/uuid"
)

type Config struct {
	SlackChannels  []string `json:"slack_channels"`
	JiraProjects   []string `json:"jira_projects"`
	GitlabProjects []string `json:"gitlab_projects"`
	GoogleSheets   []string `json:"google_sheets"`
}

// SyncIntegrationsHandler triggers a pull from external systems
func SyncIntegrationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Platform string `json:"platform"`
		TargetID string `json:"target_id"`
		FullSync bool   `json:"full_sync"`
	}
	if r.Body != nil {
		json.NewDecoder(r.Body).Decode(&req)
	}

	var targets []models.IntegrationTarget
	if db.DB != nil {
		db.DB.Find(&targets)
	}

	// Return immediately so the UI doesn't hang
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Integrations synchronization started in the background",
	})

	// Run everything in the background sequentially to avoid hitting LLM 429 Rate Limits
	go func() {
		recordSync := func(platform, targetID string, syncFunc func(string) (string, error)) {
			log.Printf("[SYNC] Pulling %s: %s", platform, targetID)
			msg, err := syncFunc(targetID)

			syncLog := models.SyncLog{
				ID:        uuid.New().String(),
				Platform:  platform,
				TargetID:  targetID,
				CreatedAt: time.Now(),
			}
			if err != nil {
				log.Printf("[ERROR] Failed to sync %s %s: %v", platform, targetID, err)
				syncLog.Status = "ERROR"
				syncLog.Message = err.Error()
			} else {
				syncLog.Status = "SUCCESS"
				syncLog.Message = msg
			}
			if db.DB != nil {
				db.DB.Create(&syncLog)
			}

			// Sleep to allow LLM rate limit buckets (Groq TPM) to reset
			time.Sleep(5 * time.Second)
		}

		// Process targets dynamically
		for _, target := range targets {
			if req.Platform != "" && req.Platform != target.Platform {
				continue
			}
			if req.TargetID != "" && req.TargetID != target.TargetID {
				continue
			}

			switch target.Platform {
			case "sheets":
				recordSync("sheets", target.TargetID, SyncGoogleSheet)
			case "slack":
				recordSync("slack", target.TargetID, SyncSlackChannel)
			case "gitlab":
				recordSync("gitlab", target.TargetID, SyncGitlabProject)
			case "jira":
				log.Printf("[SYNC] Pulling jira: %s", target.TargetID)
				msg, err := SyncJiraProject(target.TargetID, req.FullSync)
				syncLog := models.SyncLog{
					ID:        uuid.New().String(),
					Platform:  "jira",
					TargetID:  target.TargetID,
					CreatedAt: time.Now(),
				}
				if err != nil {
					log.Printf("[ERROR] Failed to sync jira %s: %v", target.TargetID, err)
					syncLog.Status = "ERROR"
					syncLog.Message = err.Error()
				} else {
					syncLog.Status = "SUCCESS"
					syncLog.Message = msg
				}
				if db.DB != nil {
					db.DB.Create(&syncLog)
				}
			}
		}
	}()
}

// GetSyncLogsHandler retrieves recent sync history
func GetSyncLogsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var logs []models.SyncLog
	if db.DB != nil {
		db.DB.Order("created_at desc").Limit(100).Find(&logs)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

// GetTargetsHandler retrieves all active integration targets
func GetTargetsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var targets []models.IntegrationTarget
	if db.DB != nil {
		db.DB.Order("created_at desc").Find(&targets)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(targets)
}

// AddTargetHandler adds a new integration target
func AddTargetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Platform string `json:"platform"`
		TargetID string `json:"target_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	if req.Platform == "" || req.TargetID == "" {
		http.Error(w, "Platform and target_id required", http.StatusBadRequest)
		return
	}

	target := models.IntegrationTarget{
		ID:        uuid.New().String(),
		Platform:  req.Platform,
		TargetID:  req.TargetID,
		CreatedAt: time.Now(),
	}

	if db.DB != nil {
		if err := db.DB.Create(&target).Error; err != nil {
			http.Error(w, "Failed to save target", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(target)
}

// DeleteTargetHandler removes an integration target
func DeleteTargetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Target ID is required", http.StatusBadRequest)
		return
	}

	if db.DB != nil {
		if err := db.DB.Where("id = ?", id).Delete(&models.IntegrationTarget{}).Error; err != nil {
			http.Error(w, "Failed to delete target", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "deleted"})
}
