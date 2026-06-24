package integrations

import (
	"encoding/json"
	"testing"
)

func TestParseJiraChangelog(t *testing.T) {
	mockResponseJSON := `{
		"issues": [
			{
				"key": "TEST-1",
				"fields": {
					"summary": "Test Issue",
					"status": {
						"name": "Done"
					}
				},
				"changelog": {
					"histories": [
						{
							"id": "10001",
							"created": "2024-06-20T10:00:00.000+0000",
							"items": [
								{
									"field": "status",
									"fromString": "To Do",
									"toString": "In Progress"
								}
							]
						},
						{
							"id": "10002",
							"created": "2024-06-21T15:30:00.000+0000",
							"items": [
								{
									"field": "status",
									"fromString": "In Progress",
									"toString": "Done"
								}
							]
						}
					]
				}
			}
		]
	}`

	var response JiraSearchResponse
	if err := json.Unmarshal([]byte(mockResponseJSON), &response); err != nil {
		t.Fatalf("Failed to parse mock JSON: %v", err)
	}

	if len(response.Issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(response.Issues))
	}

	issue := response.Issues[0]
	if issue.Key != "TEST-1" {
		t.Errorf("Expected key TEST-1, got %s", issue.Key)
	}

	histories := issue.Changelog.Histories
	if len(histories) != 2 {
		t.Fatalf("Expected 2 histories, got %d", len(histories))
	}

	h1 := histories[0]
	if h1.Created != "2024-06-20T10:00:00.000+0000" {
		t.Errorf("Expected first history created date 2024-06-20T10:00:00.000+0000, got %s", h1.Created)
	}
	if len(h1.Items) != 1 {
		t.Fatalf("Expected 1 item in first history, got %d", len(h1.Items))
	}
	item1 := h1.Items[0]
	if item1.Field != "status" || item1.ToString != "In Progress" {
		t.Errorf("Expected status change to In Progress, got %v", item1)
	}

	h2 := histories[1]
	if h2.Created != "2024-06-21T15:30:00.000+0000" {
		t.Errorf("Expected second history created date 2024-06-21T15:30:00.000+0000, got %s", h2.Created)
	}
	if len(h2.Items) != 1 {
		t.Fatalf("Expected 1 item in second history, got %d", len(h2.Items))
	}
	item2 := h2.Items[0]
	if item2.Field != "status" || item2.ToString != "Done" {
		t.Errorf("Expected status change to Done, got %v", item2)
	}
}
