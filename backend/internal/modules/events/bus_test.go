package events

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTaskStatusChangedPayload(t *testing.T) {
	// Test payload mimicking what's published by Jira sync
	payloadBytes := []byte(`{"task_name":"Test Task","status":"IN_PROGRESS","timestamp":"2024-02-18T10:30:00.000+0000","jira_key":"TEST-1"}`)

	var payload struct {
		TaskName  string `json:"task_name"`
		Status    string `json:"status"`
		Timestamp string `json:"timestamp,omitempty"`
		JiraKey   string `json:"jira_key,omitempty"`
	}

	err := json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if payload.TaskName != "Test Task" {
		t.Errorf("Expected TaskName 'Test Task', got '%s'", payload.TaskName)
	}
	if payload.Status != "IN_PROGRESS" {
		t.Errorf("Expected Status 'IN_PROGRESS', got '%s'", payload.Status)
	}
	if payload.Timestamp != "2024-02-18T10:30:00.000+0000" {
		t.Errorf("Expected Timestamp '2024-02-18T10:30:00.000+0000', got '%s'", payload.Timestamp)
	}
	if payload.JiraKey != "TEST-1" {
		t.Errorf("Expected JiraKey 'TEST-1', got '%s'", payload.JiraKey)
	}

	// Test timestamp parsing logic from bus.go
	parsedTime, err := time.Parse("2006-01-02T15:04:05.999-0700", payload.Timestamp)
	if err != nil {
		parsedTime, err = time.Parse(time.RFC3339, payload.Timestamp)
	}

	if err != nil {
		t.Fatalf("Failed to parse timestamp: %v", err)
	}

	expectedYear := 2024
	if parsedTime.Year() != expectedYear {
		t.Errorf("Expected year %d, got %d", expectedYear, parsedTime.Year())
	}
}
