package events

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/aios/backend/internal/models"
	"github.com/google/uuid"
)

func TestParseTimestamp(t *testing.T) {
	// Simulate the backdating logic in bus.go
	payloadStr := `{"task_name":"Test Task","status":"IN_PROGRESS","timestamp":"2024-06-20T10:00:00.000+0000"}`

	eventRecord := models.OperationalEventRecord{
		ID:        uuid.New().String(),
		EventType: "TASK_STATUS_CHANGED",
		Payload:   payloadStr,
	}

	var timestampPayload struct {
		Timestamp string `json:"timestamp"`
	}
	if err := json.Unmarshal([]byte(payloadStr), &timestampPayload); err == nil && timestampPayload.Timestamp != "" {
		if parsedTime, err := time.Parse(time.RFC3339, timestampPayload.Timestamp); err == nil {
			eventRecord.CreatedAt = parsedTime
		} else if parsedTime, err := time.Parse("2006-01-02T15:04:05.999-0700", timestampPayload.Timestamp); err == nil {
			eventRecord.CreatedAt = parsedTime
		} else {
			t.Errorf("Failed to parse timestamp: %v", err)
		}
	} else {
		t.Errorf("Failed to unmarshal timestamp payload: %v", err)
	}

	expectedTime, _ := time.Parse("2006-01-02T15:04:05.999-0700", "2024-06-20T10:00:00.000+0000")
	if !eventRecord.CreatedAt.Equal(expectedTime) {
		t.Errorf("Expected CreatedAt to be %v, got %v", expectedTime, eventRecord.CreatedAt)
	}
}

func TestParseTimestampRFC3339(t *testing.T) {
	// Simulate the backdating logic in bus.go
	payloadStr := `{"task_name":"Test Task","status":"IN_PROGRESS","timestamp":"2024-06-20T10:00:00Z"}`

	eventRecord := models.OperationalEventRecord{
		ID:        uuid.New().String(),
		EventType: "TASK_STATUS_CHANGED",
		Payload:   payloadStr,
	}

	var timestampPayload struct {
		Timestamp string `json:"timestamp"`
	}
	if err := json.Unmarshal([]byte(payloadStr), &timestampPayload); err == nil && timestampPayload.Timestamp != "" {
		if parsedTime, err := time.Parse(time.RFC3339, timestampPayload.Timestamp); err == nil {
			eventRecord.CreatedAt = parsedTime
		} else if parsedTime, err := time.Parse("2006-01-02T15:04:05.999-0700", timestampPayload.Timestamp); err == nil {
			eventRecord.CreatedAt = parsedTime
		} else {
			t.Errorf("Failed to parse timestamp: %v", err)
		}
	} else {
		t.Errorf("Failed to unmarshal timestamp payload: %v", err)
	}

	expectedTime, _ := time.Parse(time.RFC3339, "2024-06-20T10:00:00Z")
	if !eventRecord.CreatedAt.Equal(expectedTime) {
		t.Errorf("Expected CreatedAt to be %v, got %v", expectedTime, eventRecord.CreatedAt)
	}
}
