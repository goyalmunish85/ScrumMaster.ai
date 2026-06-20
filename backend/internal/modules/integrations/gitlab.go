package integrations

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/aios/backend/internal/modules/events"
)

// SyncGitlabProject fetches issues from a GitLab project and pushes them to the event bus
func SyncGitlabProject(projectID string) (string, error) {
	domain := os.Getenv("GITLAB_DOMAIN")
	token := os.Getenv("GITLAB_API_TOKEN")

	if domain == "" {
		domain = "gitlab.com"
	}
	if token == "" {
		return "", fmt.Errorf("GITLAB_API_TOKEN missing in .env")
	}

	// Fetch issues from the GitLab project
	// URL Encode the project ID (can be numeric or namespace/project)
	encodedProjectID := strings.ReplaceAll(projectID, "/", "%2F")
	apiURL := fmt.Sprintf("https://%s/api/v4/projects/%s/issues?state=opened&per_page=100", domain, encodedProjectID)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("PRIVATE-TOKEN", token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch gitlab project %s, status: %d", projectID, resp.StatusCode)
	}

	var issues []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return "", err
	}

	if len(issues) == 0 {
		return fmt.Sprintf("Synced gitlab project %s: 0 issues found", projectID), nil
	}

	type ExtractedTask struct {
		JiraKey     string `json:"jira_key"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Status      string `json:"status"`
		DueDate     string `json:"due_date"`
		Assignee    string `json:"assignee"`
		Project     string `json:"project"`
		SourceName  string `json:"source_name"`
	}

	var tasks []ExtractedTask
	for _, issue := range issues {
		iid := ""
		if val, ok := issue["iid"].(float64); ok {
			iid = fmt.Sprintf("GL-%.0f", val)
		}

		title := ""
		if val, ok := issue["title"].(string); ok {
			title = val
		}

		description := ""
		if val, ok := issue["description"].(string); ok {
			description = val
		}

		assigneeName := "Unassigned"
		if assignee, ok := issue["assignee"].(map[string]interface{}); ok && assignee != nil {
			if name, ok := assignee["name"].(string); ok {
				assigneeName = name
			}
		}

		dueDate := ""
		if val, ok := issue["due_date"].(string); ok && val != "" {
			dueDate = val
		}

		status := "IN_PROGRESS"

		tasks = append(tasks, ExtractedTask{
			JiraKey:     iid, // Using JiraKey field for universal external ID
			Name:        title,
			Description: description,
			Status:      status,
			DueDate:     dueDate,
			Assignee:    assigneeName,
			Project:     projectID,
			SourceName:  "GitLab: " + projectID,
		})
	}

	// 2. Fetch Merge Requests
	mrURL := fmt.Sprintf("https://%s/api/v4/projects/%s/merge_requests?state=opened&per_page=100", domain, encodedProjectID)
	reqMR, _ := http.NewRequest("GET", mrURL, nil)
	reqMR.Header.Set("PRIVATE-TOKEN", token)
	respMR, err := client.Do(reqMR)
	if err == nil && respMR.StatusCode == http.StatusOK {
		var mrs []map[string]interface{}
		json.NewDecoder(respMR.Body).Decode(&mrs)
		respMR.Body.Close()

		for _, mr := range mrs {
			iid := ""
			if val, ok := mr["iid"].(float64); ok {
				iid = fmt.Sprintf("MR-%.0f", val)
			}

			title := ""
			if val, ok := mr["title"].(string); ok {
				title = val
			}

			description := ""
			if val, ok := mr["description"].(string); ok {
				description = val
			}

			assigneeName := "Unassigned"
			if assignee, ok := mr["assignee"].(map[string]interface{}); ok && assignee != nil {
				if name, ok := assignee["name"].(string); ok {
					assigneeName = name
				}
			}

			tasks = append(tasks, ExtractedTask{
				JiraKey:     iid,
				Name:        "[MR] " + title,
				Description: description,
				Status:      "IN_PROGRESS",
				Assignee:    assigneeName,
				Project:     projectID,
				SourceName:  "GitLab: " + projectID,
			})
		}
	}

	payloadBytes, _ := json.Marshal(map[string]interface{}{
		"tasks": tasks,
	})

	events.Publish(events.OperationalEvent{Type: events.BulkTasks, Payload: payloadBytes})

	return fmt.Sprintf("Extracted %d issues from GitLab %s", len(tasks), projectID), nil
}
