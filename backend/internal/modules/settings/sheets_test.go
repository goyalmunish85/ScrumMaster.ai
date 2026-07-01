package settings

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/aios/backend/internal/db"
	"github.com/aios/backend/internal/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) {
	t.Helper()
	var err error
	db.DB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	err = db.DB.AutoMigrate(&models.IntegrationTarget{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	// Insert some initial data
	initialTarget := models.IntegrationTarget{
		ID:       "test-id-1",
		Platform: "sheets",
		TargetID: "OLD_SHEET_1",
	}
	if err := db.DB.Create(&initialTarget).Error; err != nil {
		t.Fatalf("Failed to insert initial data: %v", err)
	}

	otherTarget := models.IntegrationTarget{
		ID:       "test-id-2",
		Platform: "slack",
		TargetID: "SLACK_CH_1",
	}
	if err := db.DB.Create(&otherTarget).Error; err != nil {
		t.Fatalf("Failed to insert initial data: %v", err)
	}
}

func TestUpdateSheetsConfigHandler(t *testing.T) {
	setupTestDB(t)

	// Clean up environment variable after tests
	origEnv := os.Getenv("GOOGLE_SHEETS")
	defer os.Setenv("GOOGLE_SHEETS", origEnv)
	os.Setenv("GOOGLE_SHEETS", "")

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
			payload:        SheetsConfigRequest{GoogleSheets: []string{"NEW"}},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Method not allowed\n",
		},
		{
			name:           "Method Not Allowed (PUT)",
			method:         http.MethodPut,
			contentType:    "application/json",
			payload:        SheetsConfigRequest{GoogleSheets: []string{"NEW"}},
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
			payload:        "{invalid_json}",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid JSON payload\n",
		},
		{
			name:           "Missing google_sheets (nil)",
			method:         http.MethodPost,
			contentType:    "application/json",
			payload:        map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "google_sheets is required\n",
		},
		{
			name:           "Empty string in google_sheets",
			method:         http.MethodPost,
			contentType:    "application/json",
			payload:        SheetsConfigRequest{GoogleSheets: []string{"SAAS", ""}},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "google_sheets cannot contain empty strings\n",
		},
		{
			name:           "Valid Payload - Empty Array",
			method:         http.MethodPost,
			contentType:    "application/json",
			payload:        SheetsConfigRequest{GoogleSheets: []string{}},
			expectedStatus: http.StatusOK,
			expectedBody:   "success",
		},
		{
			name:           "Valid Payload - New Sheets",
			method:         http.MethodPost,
			contentType:    "application/json",
			payload:        SheetsConfigRequest{GoogleSheets: []string{"NEW1", "NEW2"}},
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

			req := httptest.NewRequest(tt.method, "/api/settings/sheets", bytes.NewReader(bodyBytes))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(UpdateSheetsConfigHandler)
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

				// Verify DB state
				var targets []models.IntegrationTarget
				if err := db.DB.Where("platform = ?", "sheets").Find(&targets).Error; err != nil {
					t.Fatalf("Failed to query DB: %v", err)
				}

				expectedSheets := tt.payload.(SheetsConfigRequest).GoogleSheets
				if len(targets) != len(expectedSheets) {
					t.Errorf("Expected %d targets in DB, got %d", len(expectedSheets), len(targets))
				}

				// Collect actual sheet IDs from DB
				actualSheets := make([]string, 0, len(targets))
				for _, target := range targets {
					actualSheets = append(actualSheets, target.TargetID)
				}

				// They should match the input
				for _, expectedSheet := range expectedSheets {
					found := false
					for _, actualSheet := range actualSheets {
						if expectedSheet == actualSheet {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected sheet %s not found in DB", expectedSheet)
					}
				}

				// Verify other platforms are not affected
				var slackTargets []models.IntegrationTarget
				if err := db.DB.Where("platform = ?", "slack").Find(&slackTargets).Error; err != nil {
					t.Fatalf("Failed to query DB for slack: %v", err)
				}
				if len(slackTargets) != 1 || slackTargets[0].TargetID != "SLACK_CH_1" {
					t.Errorf("Slack targets were incorrectly modified")
				}

				// Verify Env Var
				envVar := os.Getenv("GOOGLE_SHEETS")
				expectedEnvVar := ""
				for i, s := range expectedSheets {
					if i > 0 {
						expectedEnvVar += ","
					}
					expectedEnvVar += s
				}

				if envVar != expectedEnvVar {
					t.Errorf("Expected GOOGLE_SHEETS env var to be %q, got %q", expectedEnvVar, envVar)
				}
			}
		})
	}
}

func TestUpdateSheetsConfigHandler_DBError(t *testing.T) {
	setupTestDB(t)

	// Break the DB connection temporarily by closing the underlying sqlite db
	// For GORM, we can drop the table to force an error on delete/insert
	db.DB.Migrator().DropTable(&models.IntegrationTarget{})

	payload := SheetsConfigRequest{GoogleSheets: []string{"NEW1"}}
	bodyBytes, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/settings/sheets", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(UpdateSheetsConfigHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500 when DB error occurs, got %v", rr.Code)
	}
	expectedBody := "Failed to update sheets configuration in database\n"
	if rr.Body.String() != expectedBody {
		t.Errorf("Expected body %q, got %q", expectedBody, rr.Body.String())
	}
}
