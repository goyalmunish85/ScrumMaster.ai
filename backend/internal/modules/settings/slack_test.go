package settings

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aios/backend/internal/db"
	"github.com/aios/backend/internal/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDBSlack(t *testing.T) {
	var err error
	db.DB, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	err = db.DB.AutoMigrate(&models.IntegrationTarget{})
	if err != nil {
		t.Fatalf("Failed to migrate database schemas: %v", err)
	}
}

func TestUpdateSlackConfigHandler(t *testing.T) {
	setupTestDBSlack(t)

	tests := []struct {
		name           string
		method         string
		contentType    string
		payload        interface{}
		expectedStatus int
		expectedBody   string
		expectedDB     []string
	}{
		{
			name:           "Method Not Allowed (GET)",
			method:         http.MethodGet,
			contentType:    "application/json",
			payload:        SlackConfigRequest{SlackChannels: []string{"C1"}},
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
			payload:        "{invalid_json}", // constructed manually below
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid JSON payload\n",
		},
		{
			name:           "Missing slack_channels",
			method:         http.MethodPost,
			contentType:    "application/json",
			payload:        map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "slack_channels is required\n",
		},
		{
			name:           "Empty string in slack_channels",
			method:         http.MethodPost,
			contentType:    "application/json",
			payload:        SlackConfigRequest{SlackChannels: []string{"C1", ""}},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "slack_channels cannot contain empty strings\n",
		},
		{
			name:           "Valid Payload - New Channels",
			method:         http.MethodPost,
			contentType:    "application/json",
			payload:        SlackConfigRequest{SlackChannels: []string{"C1", "C2"}},
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
			expectedDB:     []string{"C1", "C2"},
		},
		{
			name:           "Valid Payload - Update Channels",
			method:         http.MethodPost,
			contentType:    "application/json",
			payload:        SlackConfigRequest{SlackChannels: []string{"C3"}},
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
			expectedDB:     []string{"C3"},
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

			req := httptest.NewRequest(tt.method, "/api/settings/slack", bytes.NewReader(bodyBytes))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(UpdateSlackConfigHandler)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %v, got %v", tt.expectedStatus, rr.Code)
			}

			if tt.expectedStatus != http.StatusOK {
				if rr.Body.String() != tt.expectedBody {
					t.Errorf("Expected body %q, got %q", tt.expectedBody, rr.Body.String())
				}
			} else {
				var resp map[string]interface{}
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Fatalf("Failed to parse response body: %v", err)
				}
				if resp["status"] != "success" {
					t.Errorf("Expected status success, got %v", resp["status"])
				}

				// Check DB
				var targets []models.IntegrationTarget
				if err := db.DB.Where("platform = ?", "slack").Find(&targets).Error; err != nil {
					t.Fatalf("Failed to query DB: %v", err)
				}

				if len(targets) != len(tt.expectedDB) {
					t.Errorf("Expected %d DB records, got %d", len(tt.expectedDB), len(targets))
				}

				dbTargets := make(map[string]bool)
				for _, t := range targets {
					dbTargets[t.TargetID] = true
				}
				for _, expected := range tt.expectedDB {
					if !dbTargets[expected] {
						t.Errorf("Expected DB to contain %s", expected)
					}
				}
			}
		})
	}
}
