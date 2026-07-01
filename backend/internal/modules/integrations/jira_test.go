package integrations

import (
	"encoding/json"
	"testing"
)

func TestParseJiraSearchResponse(t *testing.T) {
	// Mock JSON response reflecting real Jira response with changelog
	mockJSON := []byte(`{
		"startAt": 0,
		"maxResults": 50,
		"total": 1,
		"issues": [
			{
				"key": "TEST-1",
				"fields": {
					"summary": "Test Ticket",
					"status": {
						"name": "Done"
					}
				},
				"changelog": {
					"histories": [
						{
							"created": "2024-02-18T10:30:00.000+0000",
							"items": [
								{
									"field": "status",
									"fromString": "To Do",
									"toString": "In Progress"
								}
							]
						}
					]
				}
			}
		]
	}`)

	var resp JiraSearchResponse
	err := json.Unmarshal(mockJSON, &resp)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if len(resp.Issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(resp.Issues))
	}

	issue := resp.Issues[0]
	if issue.Key != "TEST-1" {
		t.Errorf("Expected issue key 'TEST-1', got '%s'", issue.Key)
	}

	if issue.Changelog == nil {
		t.Fatal("Expected changelog to not be nil")
	}

	if len(issue.Changelog.Histories) != 1 {
		t.Fatalf("Expected 1 history, got %d", len(issue.Changelog.Histories))
	}

	history := issue.Changelog.Histories[0]
	if history.Created != "2024-02-18T10:30:00.000+0000" {
		t.Errorf("Expected created date '2024-02-18T10:30:00.000+0000', got '%s'", history.Created)
	}

	if len(history.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(history.Items))
	}

	item := history.Items[0]
	if item.Field != "status" {
		t.Errorf("Expected item field 'status', got '%s'", item.Field)
	}
	if item.ToString != "In Progress" {
		t.Errorf("Expected item toString 'In Progress', got '%s'", item.ToString)
	}
}
