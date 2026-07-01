package tasks

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/aios/backend/internal/db"
	"github.com/aios/backend/internal/models"
	"github.com/aios/backend/internal/modules/memory"
	"github.com/aios/backend/internal/utils/gdpr"
)

// SearchTasksHandler handles semantic task search via GET /api/search
func SearchTasksHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "query parameter is required", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 10 // default
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			if parsedLimit > 100 { // Bound limit
				limit = 100
			} else {
				limit = parsedLimit
			}
		}
	}

	// 1. Try Qdrant semantic search
	var tasks []models.Task
	qdrantTasks, err := memory.SearchTasksSemantic(query, uint64(limit))
	if err == nil && len(qdrantTasks) > 0 {
		tasks = qdrantTasks
	} else {
		// 2. Fallback to SQL LIKE if Qdrant fails or returns 0
		db.DB.Where("LOWER(title) LIKE ?", "%"+strings.ToLower(query)+"%").Limit(limit).Find(&tasks)
	}

	// 3. GDPR Sanitization (Redact PII)
	var sanitizedTasks []interface{}
	for _, task := range tasks {
		taskJSON, _ := json.Marshal(task)
		var taskMap map[string]interface{}
		_ = json.Unmarshal(taskJSON, &taskMap)

		sanitizedTaskMap := gdpr.SanitizePayload(taskMap)
		sanitizedTasks = append(sanitizedTasks, sanitizedTaskMap)
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if len(sanitizedTasks) == 0 {
		// return empty array instead of null
		_, _ = w.Write([]byte("[]"))
		return
	}
	_ = json.NewEncoder(w).Encode(sanitizedTasks)
}
