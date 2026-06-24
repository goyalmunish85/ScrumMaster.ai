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
	var returnedEvents []events.ActivityResponse
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

	// Test invalid limit
	reqInvalidLimit, _ := http.NewRequest(http.MethodGet, "/api/activities?limit=abc", nil)
	reqInvalidLimit.Header.Set("Authorization", "Bearer mock-token")
	rrInvalidLimit := httptest.NewRecorder()
	handler.ServeHTTP(rrInvalidLimit, reqInvalidLimit)
	if rrInvalidLimit.Code != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code for invalid limit: got %v want %v",
			rrInvalidLimit.Code, http.StatusBadRequest)
	}

	// Test negative limit
	reqNegativeLimit, _ := http.NewRequest(http.MethodGet, "/api/activities?limit=-5", nil)
	reqNegativeLimit.Header.Set("Authorization", "Bearer mock-token")
	rrNegativeLimit := httptest.NewRecorder()
	handler.ServeHTTP(rrNegativeLimit, reqNegativeLimit)
	if rrNegativeLimit.Code != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code for negative limit: got %v want %v",
			rrNegativeLimit.Code, http.StatusBadRequest)
	}

	// Test invalid offset
	reqInvalidOffset, _ := http.NewRequest(http.MethodGet, "/api/activities?offset=abc", nil)
	reqInvalidOffset.Header.Set("Authorization", "Bearer mock-token")
	rrInvalidOffset := httptest.NewRecorder()
	handler.ServeHTTP(rrInvalidOffset, reqInvalidOffset)
	if rrInvalidOffset.Code != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code for invalid offset: got %v want %v",
			rrInvalidOffset.Code, http.StatusBadRequest)
	}

	// Test negative offset
	reqNegativeOffset, _ := http.NewRequest(http.MethodGet, "/api/activities?offset=-5", nil)
	reqNegativeOffset.Header.Set("Authorization", "Bearer mock-token")
	rrNegativeOffset := httptest.NewRecorder()
	handler.ServeHTTP(rrNegativeOffset, reqNegativeOffset)
	if rrNegativeOffset.Code != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code for negative offset: got %v want %v",
			rrNegativeOffset.Code, http.StatusBadRequest)
	}

	// Test PII Scrubbing
	testEvent3 := models.OperationalEventRecord{
		ID:        "3",
		EventType: "TASK_ASSIGNED",
		Payload:   `{"task_name":"Test Task 3", "assignee":"john.doe@example.com", "reporter":"jane.doe@example.com"}`,
		CreatedAt: now.Add(time.Hour),
	}
	db.DB.Create(&testEvent3)

	reqPII, _ := http.NewRequest(http.MethodGet, "/api/activities?limit=1", nil)
	reqPII.Header.Set("Authorization", "Bearer mock-token")
	rrPII := httptest.NewRecorder()
	handler.ServeHTTP(rrPII, reqPII)

	if rrPII.Code != http.StatusOK {
		t.Errorf("handler returned wrong status code for PII check: got %v want %v",
			rrPII.Code, http.StatusOK)
	}

	var piiEvents []events.ActivityResponse
	err = json.NewDecoder(rrPII.Body).Decode(&piiEvents)
	if err != nil {
		t.Fatalf("failed to decode response for PII check: %v", err)
	}

	if len(piiEvents) != 1 {
		t.Fatalf("expected 1 event for PII check, got %d", len(piiEvents))
	}

	if piiEvents[0].ID != "3" {
		t.Errorf("expected event ID 3, got %v", piiEvents[0].ID)
	}

	if _, ok := piiEvents[0].Payload["assignee"]; ok {
		t.Errorf("expected assignee to be scrubbed, but it was found in payload")
	}

	if _, ok := piiEvents[0].Payload["reporter"]; ok {
		t.Errorf("expected reporter to be scrubbed, but it was found in payload")
	}
}
