package settings

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/aios/backend/internal/db"
	"github.com/aios/backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SheetsConfigRequest represents the incoming request payload
type SheetsConfigRequest struct {
	GoogleSheets []string `json:"google_sheets"`
}

// UpdateSheetsConfigHandler handles POST /api/settings/sheets
func UpdateSheetsConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate content type
	contentType := r.Header.Get("Content-Type")
	if contentType != "" && contentType != "application/json" {
		http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
		return
	}

	var req SheetsConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Input validation
	if req.GoogleSheets == nil {
		http.Error(w, "google_sheets is required", http.StatusBadRequest)
		return
	}

	for _, sheet := range req.GoogleSheets {
		if sheet == "" {
			http.Error(w, "google_sheets cannot contain empty strings", http.StatusBadRequest)
			return
		}
	}

	// Update Database
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		// Hard-delete existing sheets
		if err := tx.Unscoped().Where("platform = ?", "sheets").Delete(&models.IntegrationTarget{}).Error; err != nil {
			return err
		}

		// Insert new sheets
		for _, sheet := range req.GoogleSheets {
			target := models.IntegrationTarget{
				ID:       uuid.NewString(),
				Platform: "sheets",
				TargetID: sheet,
			}
			if err := tx.Create(&target).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		http.Error(w, "Failed to update sheets configuration in database", http.StatusInternalServerError)
		return
	}

	// Update environment variable
	os.Setenv("GOOGLE_SHEETS", strings.Join(req.GoogleSheets, ","))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Sheets configuration updated successfully",
		"config": map[string]interface{}{
			"google_sheets": req.GoogleSheets,
		},
	})
}
