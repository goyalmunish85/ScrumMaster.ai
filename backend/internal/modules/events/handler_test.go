package events_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aios/backend/internal/db"
	"github.com/aios/backend/internal/models"
	"github.com/aios/backend/internal/modules/events"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestGetActivitiesHandler(t *testing.T) {
	// Setup in-memory sqlite db
	testDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}
	db.DB = testDB
	db.DB.AutoMigrate(&models.OperationalEventRecord{})

	// Insert test data
	now := time.Now()
	testEvent1 := models.OperationalEventRecord{
		ID:        "1",
		EventType: "TASK_CREATED",
		Payload:   `{"task_name":"Test Task 1"}`,
		CreatedAt: now.Add(-time.Hour),
	}
	testEvent2 := models.OperationalEventRecord{
		ID:        "2",
		EventType: "TASK_COMPLETED",
		Payload:   `{"task_name":"Test Task 2"}`,
		CreatedAt: now,
	}
	db.DB.Create(&testEvent1)
	db.DB.Create(&testEvent2)

	// Create request
	req, err := http.NewRequest(http.MethodGet, "/api/activities?limit=10&offset=0", nil)
	if err != nil {
		t.Fatal(err)
	}
	// Add auth header to bypass 401
	req.Header.Set("Authorization", "Bearer mock-token")

	// Create response recorder
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(events.GetActivitiesHandler)

	// Call handler
	handler.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check response body
	var returnedEvents []models.OperationalEventRecord
	err = json.NewDecoder(rr.Body).Decode(&returnedEvents)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(returnedEvents) != 2 {
		t.Errorf("handler returned unexpected number of events: got %v want %v",
			len(returnedEvents), 2)
	}

	// Check order (should be descending by created_at)
	if returnedEvents[0].ID != "2" || returnedEvents[1].ID != "1" {
		t.Errorf("handler returned events in wrong order: got %v want %v",
			returnedEvents[0].ID, "2")
	}

	// Test unauthorized request
	reqUnauth, _ := http.NewRequest(http.MethodGet, "/api/activities", nil)
	rrUnauth := httptest.NewRecorder()
	handler.ServeHTTP(rrUnauth, reqUnauth)
	if rrUnauth.Code != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code for unauth: got %v want %v",
			rrUnauth.Code, http.StatusUnauthorized)
	}
}
