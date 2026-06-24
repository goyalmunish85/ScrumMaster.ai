package integrations

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

)

func TestFetchSlackAPI(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		resp := SlackHistoryResponse{
			Ok: true,
			Messages: []SlackMessage{
				{
					Type: "message",
					Text: "hello",
					User: "U123",
					Ts:   "12345.000",
					Files: []SlackFile{
						{
							ID:       "F123",
							Name:     "test.png",
							Filetype: "png",
						},
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	var target SlackHistoryResponse
	err := fetchSlackAPI(ts.URL, "test-token", &target)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !target.Ok || len(target.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(target.Messages))
	}

	if len(target.Messages[0].Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(target.Messages[0].Files))
	}

	f := target.Messages[0].Files[0]
	if f.ID != "F123" || f.Name != "test.png" || f.Filetype != "png" {
		t.Fatalf("file attributes do not match, got %+v", f)
	}
}

func TestSlackMessageParsing(t *testing.T) {
	// Directly test parsing of Slack messages with files (since SyncSlackChannel requires AI and DB dependencies that are hard to mock completely without refactoring).
	// But we can verify json parsing.
	rawJSON := `{
		"type": "message",
		"subtype": "",
		"text": "Check out this file",
		"user": "U123",
		"ts": "123456.789",
		"files": [
			{
				"id": "F1",
				"name": "doc.pdf",
				"filetype": "pdf"
			},
			{
				"id": "F2",
				"name": "img.jpg",
				"filetype": "jpg"
			}
		]
	}`

	var msg SlackMessage
	if err := json.Unmarshal([]byte(rawJSON), &msg); err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(msg.Files) != 2 {
		t.Fatalf("Expected 2 files, got %d", len(msg.Files))
	}

	if msg.Files[0].ID != "F1" || msg.Files[0].Name != "doc.pdf" {
		t.Fatalf("File 1 parsed incorrectly")
	}

	if msg.Files[1].ID != "F2" || msg.Files[1].Name != "img.jpg" || msg.Files[1].Filetype != "jpg" {
		t.Fatalf("File 2 parsed incorrectly")
	}
}
