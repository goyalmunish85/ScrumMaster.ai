package tasks

import (
	"compress/gzip"
	"context"
	"encoding/csv"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/aios/backend/internal/db"
	"github.com/aios/backend/internal/models"
	"github.com/aios/backend/internal/utils/gdpr"
)

var (
	tasksCache      []OptimizedTask
	cacheMutex      sync.RWMutex
	cacheExpiration time.Time
)

const cacheDuration = 30 * time.Second

type OptimizedTask struct {
	ID         string            `json:"id"`
	Title      string            `json:"title"`
	Status     models.TaskStatus `json:"status"`
	Priority   string            `json:"priority"`
	Labels     string            `json:"labels"`
	Project    string            `json:"project"`
	JiraKey    string            `json:"jira_key"`
	Team       string            `json:"team"`
	TaskType   string            `json:"task_type"`
	Sprint     string            `json:"sprint"`
	ParentKey  string            `json:"parent_key"`
	SourceName string            `json:"source_name"`
	DueDate    *time.Time        `json:"due_date"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

type ActivityResponse struct {
	ID        string      `json:"id"`
	TaskID    *string     `json:"task_id"`
	EventType string      `json:"event_type"`
	Payload   interface{} `json:"payload"`
	CreatedAt time.Time   `json:"created_at"`
}

type TaskDetailResponse struct {
	OptimizedTask
	Description string             `json:"description"`
	Activities  []ActivityResponse `json:"activities"`
}

// GetTaskHandler fetches a single task and its linked activity logs by ID.
func GetTaskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Task ID is required", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var task models.Task
	result := db.DB.WithContext(ctx).First(&task, "id = ?", id)
	if result.Error != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	var events []models.OperationalEventRecord
	db.DB.WithContext(ctx).Order("created_at desc").Find(&events, "task_id = ?", id)

	activities := []ActivityResponse{}
	for _, event := range events {
		var payload interface{}
		if err := json.Unmarshal([]byte(event.Payload), &payload); err != nil {
			payload = make(map[string]interface{})
		}

		sanitizedPayload := gdpr.SanitizePayload(payload)

		activities = append(activities, ActivityResponse{
			ID:        event.ID,
			TaskID:    event.TaskID,
			EventType: event.EventType,
			Payload:   sanitizedPayload,
			CreatedAt: event.CreatedAt,
		})
	}

	response := TaskDetailResponse{
		OptimizedTask: OptimizedTask{
			ID:         task.ID,
			Title:      task.Title,
			Status:     task.Status,
			Priority:   task.Priority,
			Labels:     task.Labels,
			Project:    task.Project,
			JiraKey:    task.JiraKey,
			Team:       task.Team,
			TaskType:   task.TaskType,
			Sprint:     task.Sprint,
			ParentKey:  task.ParentKey,
			SourceName: task.SourceName,
			DueDate:    task.DueDate,
			CreatedAt:  task.CreatedAt,
			UpdatedAt:  task.UpdatedAt,
		},
		Description: task.Description,
		Activities:  activities,
	}

	w.Header().Set("Content-Type", "application/json")
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		json.NewEncoder(gz).Encode(response)
	} else {
		json.NewEncoder(w).Encode(response)
	}
}

// GetTasksHandler fetches all active tasks from the database.
func GetTasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cacheMutex.RLock()
	if time.Now().Before(cacheExpiration) && tasksCache != nil {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			w.Header().Set("Content-Encoding", "gzip")
			gz := gzip.NewWriter(w)
			defer gz.Close()
			json.NewEncoder(gz).Encode(tasksCache)
		} else {
			json.NewEncoder(w).Encode(tasksCache)
		}
		cacheMutex.RUnlock()
		return
	}
	cacheMutex.RUnlock()

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var allTasks []models.Task
	// Order by most recently updated so the dashboard feels lively
	result := db.DB.WithContext(ctx).Order("updated_at desc").Find(&allTasks)

	if result.Error != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	optimizedTasks := make([]OptimizedTask, len(allTasks))
	for i, t := range allTasks {
		optimizedTasks[i] = OptimizedTask{
			ID:         t.ID,
			Title:      t.Title,
			Status:     t.Status,
			Priority:   t.Priority,
			Labels:     t.Labels,
			Project:    t.Project,
			JiraKey:    t.JiraKey,
			Team:       t.Team,
			TaskType:   t.TaskType,
			Sprint:     t.Sprint,
			ParentKey:  t.ParentKey,
			SourceName: t.SourceName,
			DueDate:    t.DueDate,
			CreatedAt:  t.CreatedAt,
			UpdatedAt:  t.UpdatedAt,
		}
	}

	cacheMutex.Lock()
	tasksCache = optimizedTasks
	cacheExpiration = time.Now().Add(cacheDuration)
	cacheMutex.Unlock()

	w.Header().Set("Content-Type", "application/json")

	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		json.NewEncoder(gz).Encode(optimizedTasks)
	} else {
		json.NewEncoder(w).Encode(optimizedTasks)
	}
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
