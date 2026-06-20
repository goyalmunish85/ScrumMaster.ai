package integrations

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/aios/backend/internal/modules/events"
)

// SyncGoogleSheet downloads a public Google Sheet as CSV, parses it, and pushes BulkTasks to the Event Bus.
func SyncGoogleSheet(sheetID string) (string, error) {
	csvURL := fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s/export?format=csv", sheetID)

	resp, err := http.Get(csvURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch sheet, status: %d", resp.StatusCode)
	}

	reader := csv.NewReader(resp.Body)
	records, err := reader.ReadAll()
	if err != nil {
		return "", err
	}

	// Create a short prefix for the client name
	shortID := sheetID
	if len(shortID) > 5 {
		shortID = shortID[:5]
	}

	// Zero-Token CSV Mapping Logic
	var extractedTasks []map[string]interface{}

	// Map columns
	colMap := make(map[string]int)
	if len(records) > 0 {
		for i, header := range records[0] {
			colMap[strings.ToLower(strings.TrimSpace(header))] = i
		}
	}

	rowCount := 0
	for i, row := range records {
		if i == 0 {
			continue // skip header
		}

		task := make(map[string]interface{})
		task["source_name"] = "Sheet: " + shortID

		for colName, colIndex := range colMap {
			if colIndex < len(row) {
				val := strings.TrimSpace(row[colIndex])
				if val != "" {
					// Auto-map common standard columns
					switch {
					case strings.Contains(colName, "task") || strings.Contains(colName, "title") || strings.Contains(colName, "name"):
						task["name"] = val
					case strings.Contains(colName, "assign"):
						task["assignee"] = val
					case strings.Contains(colName, "status") || strings.Contains(colName, "state"):
						task["status"] = val
					case strings.Contains(colName, "due") || strings.Contains(colName, "date"):
						task["due_date"] = val
					case strings.Contains(colName, "desc"):
						task["description"] = val
					case strings.Contains(colName, "client"):
						task["client"] = val
					case strings.Contains(colName, "team"):
						task["team"] = val
					}
				}
			}
		}

		// Only add if it has a name
		if name, ok := task["name"].(string); ok && name != "" {
			extractedTasks = append(extractedTasks, task)
			rowCount++
		}
	}

	if len(extractedTasks) == 0 {
		return fmt.Sprintf("Skipped: Sheet %s has no valid task rows.", shortID), nil
	}

	payload := map[string]interface{}{
		"tasks": extractedTasks,
	}
	payloadBytes, _ := json.Marshal(payload)

	// Publish to Event Bus
	events.Publish(events.OperationalEvent{
		Type:    events.BulkTasks,
		Payload: payloadBytes,
	})

	return fmt.Sprintf("Extracted %d tasks/events from %d Sheet rows in %s.", len(extractedTasks), rowCount, shortID), nil
}
