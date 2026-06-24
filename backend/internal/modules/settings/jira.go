package settings

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"sync"
)

var configMutex sync.Mutex

// IntegrationsConfig represents the full structure of integrations.json
type IntegrationsConfig struct {
	SlackChannels  []string `json:"slack_channels"`
	JiraProjects   []string `json:"jira_projects"`
	GitlabProjects []string `json:"gitlab_projects"`
	GoogleSheets   []string `json:"google_sheets"`
}

// JiraConfigRequest represents the incoming request payload
type JiraConfigRequest struct {
	JiraProjects []string `json:"jira_projects"`
}

func getIntegrationsPath() (string, error) {
	// Simple lookup relative to cwd. In a real environment, you might
	// pass this from main or use an env variable. We try a couple common places.
	paths := []string{"integrations.json", "../integrations.json", "../../integrations.json", "../../../integrations.json", "../../../../integrations.json", "backend/integrations.json"}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	// For tests, return a dummy path or error
	return "", fmt.Errorf("integrations.json not found")
}

// UpdateJiraConfigHandler handles POST /api/settings/jira
func UpdateJiraConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Validate content type
	contentType := r.Header.Get("Content-Type")
	if contentType != "" && contentType != "application/json" {
		http.Error(w, "Unsupported Media Type", http.StatusUnsupportedMediaType)
		return
	}

	var req JiraConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	// Input validation: ensure it's an array, not nil (though Decode will make it nil if not present, we can just allow empty arrays but reject invalid types)
	if req.JiraProjects == nil {
		http.Error(w, "jira_projects is required", http.StatusBadRequest)
		return
	}

	for _, proj := range req.JiraProjects {
		if proj == "" {
			http.Error(w, "jira_projects cannot contain empty strings", http.StatusBadRequest)
			return
		}
	}

	// Lock for writing
	configMutex.Lock()
	defer configMutex.Unlock()

	path, err := getIntegrationsPath()
	if err != nil {
		http.Error(w, "Failed to locate configuration file", http.StatusInternalServerError)
		return
	}

	fileData, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, "Failed to read configuration file", http.StatusInternalServerError)
		return
	}

	var config IntegrationsConfig
	if err := json.Unmarshal(fileData, &config); err != nil {
		http.Error(w, "Failed to parse configuration file", http.StatusInternalServerError)
		return
	}

	// Update only JiraProjects
	config.JiraProjects = req.JiraProjects

	newData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		http.Error(w, "Failed to marshal configuration file", http.StatusInternalServerError)
		return
	}

	// Write back
	if err := os.WriteFile(path, newData, 0644); err != nil {
		http.Error(w, "Failed to write configuration file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Jira configuration updated successfully",
		"config":  config,
	})
}
