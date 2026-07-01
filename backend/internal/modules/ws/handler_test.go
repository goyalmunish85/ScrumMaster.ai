package ws

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestWsHandler(t *testing.T) {
	// Start a test server
	server := httptest.NewServer(http.HandlerFunc(WsHandler))
	defer server.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect a client
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket server: %v", err)
	}
	defer conn.Close()

	// Wait briefly for connection to be registered on the server side
	// (not strictly necessary since we broadcast next and it might race,
	// but let's just make sure we send message)

	// Broadcast an event
	taskID := "123"
	testEvent := WsEvent{
		Event:   "task_updated",
		TaskID:  &taskID,
	}

	// Run broadcast in a goroutine because it writes to connections
	go BroadcastEvent(testEvent)

	// Read message from the client
	_, message, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read message from WebSocket server: %v", err)
	}

	var receivedEvent map[string]interface{}
	err = json.Unmarshal(message, &receivedEvent)
	if err != nil {
		t.Fatalf("Failed to parse message: %v", err)
	}

	if receivedEvent["event"] != "task_updated" {
		t.Errorf("Expected event 'task_updated', got %v", receivedEvent["event"])
	}
	if receivedEvent["task_id"] != "123" {
		t.Errorf("Expected task_id '123', got %v", receivedEvent["task_id"])
	}
}
