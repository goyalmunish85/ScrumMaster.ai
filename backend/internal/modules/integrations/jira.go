package integrations

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/aios/backend/internal/modules/events"
)

type JiraChangelog struct {
	Histories []struct {
		Created string `json:"created"`
		Items   []struct {
			Field      string `json:"field"`
			FromString string `json:"fromString"`
			ToString   string `json:"toString"`
		} `json:"items"`
	} `json:"histories"`
}

type JiraSearchResponse struct {
	Names      map[string]string `json:"names"`
	StartAt    int               `json:"startAt"`
	MaxResults int               `json:"maxResults"`
	Total      int               `json:"total"`
	Issues     []struct {
		Key       string                 `json:"key"`
		Fields    map[string]interface{} `json:"fields"`
		Changelog *JiraChangelog         `json:"changelog"`
	} `json:"issues"`
}

// SyncJiraProject fetches tickets for a project and extracts them as tasks
func SyncJiraProject(projectKey string, fullSync bool) (string, error) {
	domain := os.Getenv("JIRA_DOMAIN")
	email := os.Getenv("JIRA_EMAIL")
	token := os.Getenv("JIRA_API_TOKEN")

	if domain == "" || email == "" || token == "" {
		return "", fmt.Errorf("JIRA credentials are missing in .env")
	}

	jql := fmt.Sprintf("project=%s AND updated >= -3d", projectKey)
	if fullSync {
		jql = fmt.Sprintf("project=%s", projectKey)
	}
	encodedJQL := url.QueryEscape(jql)

	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", email, token)))

	var allIssues []map[string]interface{}
	var fieldNames map[string]string

	startAt := 0
	maxResults := 50

	for {
		apiURL := fmt.Sprintf("https://%s/rest/api/3/search/jql?jql=%s&expand=names,changelog&fields=*all&maxResults=%d&startAt=%d", domain, encodedJQL, maxResults, startAt)

		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			return "", err
		}

		req.Header.Set("Authorization", "Basic "+auth)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return "", fmt.Errorf("failed to fetch jira api, status: %d", resp.StatusCode)
		}

		var jiraResp JiraSearchResponse
		if err := json.NewDecoder(resp.Body).Decode(&jiraResp); err != nil {
			resp.Body.Close()
			return "", err
		}
		resp.Body.Close()

		if fieldNames == nil && jiraResp.Names != nil {
			fieldNames = jiraResp.Names
		}

		for _, issue := range jiraResp.Issues {
			issueMap := map[string]interface{}{
				"key":       issue.Key,
				"fields":    issue.Fields,
				"changelog": issue.Changelog,
			}
			allIssues = append(allIssues, issueMap)
		}

		if len(jiraResp.Issues) == 0 {
			break
		}

		startAt += len(jiraResp.Issues)
		if startAt >= jiraResp.Total {
			break
		}
	}

	log.Printf("[SYNC] Fetched %d total tickets from Jira project %s (Active in last 3 days)", len(allIssues), projectKey)

	// Create reverse mapping: "Customer Name" -> "customfield_10014"
	nameToCustomField := make(map[string]string)
	for cfID, name := range fieldNames {
		nameToCustomField[name] = cfID
	}

	type ExtractedTask struct {
		JiraKey     string `json:"jira_key"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Status      string `json:"status"`
		DueDate     string `json:"due_date,omitempty"`
		Priority    string `json:"priority"`
		Labels      string `json:"labels"`
		Assignee    string `json:"assignee"`
		Reporter    string `json:"reporter"`
		Project     string `json:"project"`

		// Rich metadata
		Client     string `json:"client"`
		Team       string `json:"team"`
		TaskType   string `json:"task_type"`
		Sprint     string `json:"sprint"`
		ParentKey  string `json:"parent_key"`
		SourceName string `json:"source_name"`
	}
	var tasks []ExtractedTask

	getStringField := func(fields map[string]interface{}, key string) string {
		if val, ok := fields[key].(string); ok {
			return val
		}
		return ""
	}

	getNestedStringField := func(fields map[string]interface{}, parent, child string) string {
		if obj, ok := fields[parent].(map[string]interface{}); ok {
			if val, ok := obj[child].(string); ok {
				return val
			}
		}
		return ""
	}

	getCustomFieldString := func(fields map[string]interface{}, name string) string {
		cfID, ok := nameToCustomField[name]
		if !ok {
			return ""
		}
		val := fields[cfID]
		if val == nil {
			return ""
		}
		// It could be a string
		if s, ok := val.(string); ok {
			return s
		}
		// It could be an object with "value" or "name"
		if obj, ok := val.(map[string]interface{}); ok {
			if v, ok := obj["value"].(string); ok {
				return v
			}
			if n, ok := obj["name"].(string); ok {
				return n
			}
		}
		// It could be an array of objects (like sprint)
		if arr, ok := val.([]interface{}); ok && len(arr) > 0 {
			if obj, ok := arr[0].(map[string]interface{}); ok {
				if n, ok := obj["name"].(string); ok {
					return n
				}
				if v, ok := obj["value"].(string); ok {
					return v
				}
			}
		}
		return ""
	}

	for _, issueData := range allIssues {
		key := issueData["key"].(string)
		fields := issueData["fields"].(map[string]interface{})

		jiraStatus := getNestedStringField(fields, "status", "name")
		status := "DRAFT"
		if jiraStatus == "Done" || jiraStatus == "Closed" || jiraStatus == "Resolved" {
			status = "DONE"
		} else if jiraStatus == "In Progress" || jiraStatus == "In Review" {
			status = "IN_PROGRESS"
		}

		name := getStringField(fields, "summary")
		dueDate := getStringField(fields, "duedate")
		priority := getNestedStringField(fields, "priority", "name")
		assignee := getNestedStringField(fields, "assignee", "displayName")
		reporter := getNestedStringField(fields, "reporter", "displayName")

		labelsStr := ""
		if labels, ok := fields["labels"].([]interface{}); ok && len(labels) > 0 {
			labelsBytes, _ := json.Marshal(labels)
			labelsStr = string(labelsBytes)
		}

		description := ""
		if fields["description"] != nil {
			descBytes, _ := json.Marshal(fields["description"])
			description = string(descBytes)
		}

		// Extract Rich Metadata
		client := getCustomFieldString(fields, "Customer Name")
		team := getCustomFieldString(fields, "Team Name")
		taskType := getCustomFieldString(fields, "Request Type")
		sprint := getCustomFieldString(fields, "Sprint")
		parentKey := getCustomFieldString(fields, "Parent")
		if parentKey == "" {
			parentKey = getNestedStringField(fields, "parent", "key")
		}

		tasks = append(tasks, ExtractedTask{
			JiraKey:     key,
			Name:        name,
			Description: description,
			Status:      status,
			DueDate:     dueDate,
			Priority:    priority,
			Labels:      labelsStr,
			Assignee:    assignee,
			Reporter:    reporter,
			Project:     projectKey,
			Client:      client,
			Team:        team,
			TaskType:    taskType,
			Sprint:      sprint,
			ParentKey:   parentKey,
			SourceName:  "Jira: " + projectKey,
		})

		// Parse changelog for status transitions
		if changelogRaw, ok := issueData["changelog"]; ok && changelogRaw != nil {
			if changelog, ok := changelogRaw.(*JiraChangelog); ok && changelog != nil {
				for _, history := range changelog.Histories {
					for _, item := range history.Items {
						if item.Field == "status" {
							newStatus := "DRAFT"
							if item.ToString == "Done" || item.ToString == "Closed" || item.ToString == "Resolved" {
								newStatus = "DONE"
							} else if item.ToString == "In Progress" || item.ToString == "In Review" {
								newStatus = "IN_PROGRESS"
							}

							statusPayloadBytes, _ := json.Marshal(map[string]interface{}{
								"task_name": name,
								"status":    newStatus,
								"timestamp": history.Created,
							})

							events.Publish(events.OperationalEvent{
								Type:    events.TaskStatusChanged,
								Payload: statusPayloadBytes,
							})
						}
					}
				}
			}
		}
	}

	payloadBytes, _ := json.Marshal(map[string]interface{}{
		"tasks": tasks,
	})

	events.Publish(events.OperationalEvent{Type: events.BulkTasks, Payload: payloadBytes})

	return fmt.Sprintf("Extracted %d rich tasks from Jira project %s.", len(tasks), projectKey), nil
}
