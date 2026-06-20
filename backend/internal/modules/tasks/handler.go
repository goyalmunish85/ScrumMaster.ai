package tasks

import (
	"encoding/csv"
	"encoding/json"
	"net/http"

	"github.com/aios/backend/internal/db"
	"github.com/aios/backend/internal/models"
)

// GetTasksHandler fetches all active tasks from the database.
func GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var allTasks []models.Task
	// Order by most recently updated so the dashboard feels lively
	result := db.DB.Order("updated_at desc").Find(&allTasks)

	if result.Error != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allTasks)
}

// ExportTasksHandler fetches all active tasks and returns them as a CSV file.
func ExportTasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var allTasks []models.Task
	result := db.DB.Order("created_at desc").Find(&allTasks)
	if result.Error != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=\"tasks_export.csv\"")

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{"Jira Key", "Title", "Status", "Assignee", "Client", "Team", "Task Type", "Sprint", "Source Name", "Due Date", "Updated At"}
	if err := writer.Write(header); err != nil {
		http.Error(w, "Failed to write CSV header", http.StatusInternalServerError)
		return
	}

	// Write rows
	for _, t := range allTasks {
		dueDateStr := ""
		if t.DueDate != nil {
			dueDateStr = t.DueDate.Format("2006-01-02")
		}

		row := []string{
			t.JiraKey,
			t.Title,
			string(t.Status),
			t.Assignee,
			t.Client,
			t.Team,
			t.TaskType,
			t.Sprint,
			t.SourceName,
			dueDateStr,
			t.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		if err := writer.Write(row); err != nil {
			http.Error(w, "Failed to write CSV row", http.StatusInternalServerError)
			return
		}
	}
}
