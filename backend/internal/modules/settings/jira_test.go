package settings

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// setupMockConfig creates a temporary directory and a dummy integrations.json for testing
func setupMockConfig(t *testing.T) (string, func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "settings-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	tempFile := filepath.Join(tempDir, "integrations.json")

	initialConfig := IntegrationsConfig{
		SlackChannels:  []string{"C123"},
		JiraProjects:   []string{"OLD1", "OLD2"},
		GitlabProjects: []string{"G1"},
		GoogleSheets:   []string{"S1"},
	}

	data, err := json.MarshalIndent(initialConfig, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal initial config: %v", err)
	}

	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		t.Fatalf("Failed to write mock config: %v", err)
	}

	// Override cwd logic by changing the directory context
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get cwd: %v", err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}

	cleanup := func() {
		os.Chdir(origDir)
		os.RemoveAll(tempDir)
	}

	return tempFile, cleanup
}

func TestUpdateJiraConfigHandler(t *testing.T) {
	_, cleanup := setupMockConfig(t)
	defer cleanup()

	tests := []struct {
		name           string
		method         string
		contentType    string
		payload        interface{}
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Method Not Allowed (GET)",
			method:         http.MethodGet,
			contentType:    "application/json",
			payload:        JiraConfigRequest{JiraProjects: []string{"NEW"}},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Method not allowed\n",
		},
		{
			name:           "Method Not Allowed (PUT)",
			method:         http.MethodPut,
			contentType:    "application/json",
			payload:        JiraConfigRequest{JiraProjects: []string{"NEW"}},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Method not allowed\n",
		},
		{
			name:           "Unsupported Media Type",
			method:         http.MethodPost,
			contentType:    "text/plain",
			payload:        "some text",
			expectedStatus: http.StatusUnsupportedMediaType,
			expectedBody:   "Unsupported Media Type\n",
		},
		{
			name:           "Invalid JSON payload",
			method:         http.MethodPost,
			contentType:    "application/json",
			payload:        "{invalid_json}", // sending raw string will result in string payload if marshalled, so we'll construct the request manually below for raw bytes
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid JSON payload\n",
		},
		{
			name:           "Missing jira_projects (nil)",
			method:         http.MethodPost,
			contentType:    "application/json",
			payload:        map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "jira_projects is required\n",
		},
		{
			name:           "Empty string in jira_projects",
			method:         http.MethodPost,
			contentType:    "application/json",
			payload:        JiraConfigRequest{JiraProjects: []string{"SAAS", ""}},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "jira_projects cannot contain empty strings\n",
		},
		{
			name:           "Valid Payload - Empty Array",
			method:         http.MethodPost,
			contentType:    "application/json",
			payload:        JiraConfigRequest{JiraProjects: []string{}},
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
		{
			name:           "Valid Payload - New Projects",
			method:         http.MethodPost,
			contentType:    "application/json",
			payload:        JiraConfigRequest{JiraProjects: []string{"NEW1", "NEW2"}},
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyBytes []byte
			var err error

			if tt.name == "Invalid JSON payload" {
				bodyBytes = []byte("{invalid_json}")
			} else {
				bodyBytes, err = json.Marshal(tt.payload)
				if err != nil {
					t.Fatalf("Failed to marshal payload: %v", err)
				}
			}

			req := httptest.NewRequest(tt.method, "/api/settings/jira", bytes.NewReader(bodyBytes))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(UpdateJiraConfigHandler)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %v, got %v", tt.expectedStatus, rr.Code)
			}

			if tt.expectedStatus != http.StatusOK {
				if rr.Body.String() != tt.expectedBody {
					t.Errorf("Expected body %q, got %q", tt.expectedBody, rr.Body.String())
				}
			} else {
				// For 200 OK, verify the JSON response contains success
				var resp map[string]interface{}
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Fatalf("Failed to parse response body: %v", err)
				}
				if resp["status"] != "success" {
					t.Errorf("Expected status success, got %v", resp["status"])
				}

				// Verify the file was actually updated
				fileData, err := os.ReadFile("integrations.json")
				if err != nil {
					t.Fatalf("Failed to read updated integrations.json: %v", err)
				}
				var updatedConfig IntegrationsConfig
				if err := json.Unmarshal(fileData, &updatedConfig); err != nil {
					t.Fatalf("Failed to parse updated integrations.json: %v", err)
				}

				expectedProjects := tt.payload.(JiraConfigRequest).JiraProjects
				if len(updatedConfig.JiraProjects) != len(expectedProjects) {
					t.Errorf("Expected %d projects in file, got %d", len(expectedProjects), len(updatedConfig.JiraProjects))
				}
				for i, p := range expectedProjects {
					if updatedConfig.JiraProjects[i] != p {
						t.Errorf("Expected project %s at index %d, got %s", p, i, updatedConfig.JiraProjects[i])
					}
				}
				// Verify other fields remain unchanged
				if len(updatedConfig.SlackChannels) != 1 || updatedConfig.SlackChannels[0] != "C123" {
					t.Errorf("Slack channels were incorrectly modified")
				}
			}
		})
	}
}

func TestUpdateJiraConfigHandler_FileNotFound(t *testing.T) {
	// To reliably test file not found, we change to an empty temporary directory
	tempDir, err := os.MkdirTemp("", "empty-settings-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get cwd: %v", err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}

	defer func() {
		os.Chdir(origDir)
		os.RemoveAll(tempDir)
	}()

	payload := JiraConfigRequest{JiraProjects: []string{"NEW1"}}
	bodyBytes, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/settings/jira", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(UpdateJiraConfigHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500 when file not found, got %v", rr.Code)
	}
	expectedBody := "Failed to locate configuration file\n"
	if rr.Body.String() != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, rr.Body.String())
	}
}
