package events

import (
	"encoding/json"
	"net/http"

	"github.com/aios/backend/internal/db"
	"github.com/aios/backend/internal/models"
)

// GetEventsHandler fetches the recent operational events
func GetEventsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var recentEvents []models.OperationalEventRecord
	// Order by most recently created
	result := db.DB.Order("created_at desc").Limit(50).Find(&recentEvents)

	if result.Error != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(recentEvents)
}
