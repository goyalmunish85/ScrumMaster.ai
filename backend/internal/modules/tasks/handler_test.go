package tasks

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aios/backend/internal/db"
	"github.com/aios/backend/internal/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) {
	var err error
	db.DB, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	db.DB.Exec("DELETE FROM tasks")
	db.DB.Exec("DELETE FROM operational_event_records")

	err = db.DB.AutoMigrate(&models.Task{}, &models.OperationalEventRecord{})
	if err != nil {
		t.Fatalf("Failed to migrate database schemas: %v", err)
	}

	// Reset Cache
	cacheMutex.Lock()
	tasksCache = nil
	cacheExpiration = time.Time{}
	cacheMutex.Unlock()
}

func TestGetTasksHandler_Success(t *testing.T) {
	setupTestDB(t)

	// Create a task with PII fields
	task := models.Task{
		ID:       "test-id",
		Title:    "Test Task",
		Status:   models.StatusDraft,
		Assignee: "Sensitive Assignee",
		Reporter: "Sensitive Reporter",
		Client:   "Sensitive Client",
	}
	db.DB.Create(&task)

	req, err := http.NewRequest("GET", "/api/tasks", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetTasksHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var responseTasks []OptimizedTask
	err = json.Unmarshal(rr.Body.Bytes(), &responseTasks)
	if err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}

	if len(responseTasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(responseTasks))
	}

	if responseTasks[0].Title != "Test Task" {
		t.Errorf("Expected task title 'Test Task', got '%s'", responseTasks[0].Title)
	}

	// Verify that GDPR fields are not in the OptimizedTask payload
	// Since OptimizedTask doesn't have these fields, we check the raw JSON output as well
	var rawMap []map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &rawMap)

	if _, exists := rawMap[0]["assignee"]; exists {
		t.Errorf("Assignee should be omitted from payload")
	}
	if _, exists := rawMap[0]["reporter"]; exists {
		t.Errorf("Reporter should be omitted from payload")
	}
	if _, exists := rawMap[0]["client"]; exists {
		t.Errorf("Client should be omitted from payload")
	}
}

func TestGetTasksHandler_Cache(t *testing.T) {
	setupTestDB(t)

	// Create first task
	task1 := models.Task{
		ID:    "test-id-1",
		Title: "Test Task 1",
	}
	db.DB.Create(&task1)

	// First Request - Should cache the result
	req, _ := http.NewRequest("GET", "/api/tasks", nil)
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetTasksHandler)
	handler.ServeHTTP(rr, req)

	var responseTasks1 []OptimizedTask
	json.Unmarshal(rr.Body.Bytes(), &responseTasks1)
	if len(responseTasks1) != 1 {
		t.Fatalf("Expected 1 task in first request, got %d", len(responseTasks1))
	}

	// Create second task
	task2 := models.Task{
		ID:    "test-id-2",
		Title: "Test Task 2",
	}
	db.DB.Create(&task2)

	// Second Request - Should return cached result (only 1 task)
	req2, _ := http.NewRequest("GET", "/api/tasks", nil)
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	var responseTasks2 []OptimizedTask
	json.Unmarshal(rr2.Body.Bytes(), &responseTasks2)
	if len(responseTasks2) != 1 {
		t.Fatalf("Expected 1 task in cached second request, got %d", len(responseTasks2))
	}

	// Verify Cache Expiration (Mocking)
	cacheMutex.Lock()
	cacheExpiration = time.Now().Add(-1 * time.Second) // Expire the cache
	cacheMutex.Unlock()

	// Third Request - Cache expired, should fetch both tasks
	req3, _ := http.NewRequest("GET", "/api/tasks", nil)
	rr3 := httptest.NewRecorder()
	handler.ServeHTTP(rr3, req3)

	var responseTasks3 []OptimizedTask
	json.Unmarshal(rr3.Body.Bytes(), &responseTasks3)
	if len(responseTasks3) != 2 {
		t.Fatalf("Expected 2 tasks after cache expired, got %d", len(responseTasks3))
	}
}

func TestGetTaskHandler_Success(t *testing.T) {
	setupTestDB(t)

	taskID := "detail-test-id"
	task := models.Task{
		ID:          taskID,
		Title:       "Detail Task",
		Description: "Detailed description",
		Assignee:    "Sensitive Assignee",
	}
	db.DB.Create(&task)

	eventID := "event-id-1"
	eventPayload := `{"status": "DONE", "assignee": "Alice", "reporter": "Bob", "nested": {"email": "test@test.com", "other": "info"}}`
	event := models.OperationalEventRecord{
		ID:        eventID,
		TaskID:    &taskID,
		EventType: "TASK_UPDATED",
		Payload:   eventPayload,
		CreatedAt: time.Now(),
	}
	db.DB.Create(&event)

	req, err := http.NewRequest("GET", "/api/tasks/"+taskID, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/tasks/{id}", GetTaskHandler)
	mux.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response TaskDetailResponse
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}

	if response.ID != taskID {
		t.Errorf("Expected task ID %s, got %s", taskID, response.ID)
	}
	if response.Description != "Detailed description" {
		t.Errorf("Expected Description 'Detailed description', got '%s'", response.Description)
	}

	// Verify GDPR on main task (raw JSON check)
	var rawMap map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &rawMap)
	if _, exists := rawMap["assignee"]; exists {
		t.Errorf("Assignee should be omitted from root payload")
	}

	// Verify GDPR on linked activities
	if len(response.Activities) != 1 {
		t.Fatalf("Expected 1 activity, got %d", len(response.Activities))
	}

	actPayload, ok := response.Activities[0].Payload.(map[string]interface{})
	if !ok {
		t.Fatalf("Activity payload is not a map")
	}
	if _, exists := actPayload["assignee"]; exists {
		t.Errorf("Assignee should be omitted from activity payload")
	}
	if _, exists := actPayload["reporter"]; exists {
		t.Errorf("Reporter should be omitted from activity payload")
	}
	nestedMap, ok := actPayload["nested"].(map[string]interface{})
	if ok {
		if _, exists := nestedMap["email"]; exists {
			t.Errorf("Email should be omitted from nested activity payload")
		}
		if nestedMap["other"] != "info" {
			t.Errorf("Other field should be preserved in nested payload")
		}
	} else {
		t.Errorf("Nested payload should be a map")
	}
}

func TestGetTaskHandler_NotFound(t *testing.T) {
	setupTestDB(t)

	req, _ := http.NewRequest("GET", "/api/tasks/missing-id", nil)
	rr := httptest.NewRecorder()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/tasks/{id}", GetTaskHandler)
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected 404 Not Found, got %v", rr.Code)
	}
}

func TestGetTasksHandler_Gzip(t *testing.T) {
	setupTestDB(t)

	task := models.Task{
		ID:    "gzip-test",
		Title: "Gzip Task",
	}
	db.DB.Create(&task)

	req, _ := http.NewRequest("GET", "/api/tasks", nil)
	req.Header.Set("Accept-Encoding", "gzip")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetTasksHandler)
	handler.ServeHTTP(rr, req)

	if rr.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("Expected Content-Encoding to be gzip")
	}

	gzReader, err := gzip.NewReader(rr.Body)
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer gzReader.Close()

	bodyBytes, _ := io.ReadAll(gzReader)

	var responseTasks []OptimizedTask
	json.Unmarshal(bodyBytes, &responseTasks)
	if len(responseTasks) != 1 || responseTasks[0].Title != "Gzip Task" {
		t.Errorf("Failed to read correct data from gzipped response")
	}
}
