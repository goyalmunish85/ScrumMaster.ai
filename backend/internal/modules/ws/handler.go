package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"strings"

	"github.com/gorilla/websocket"
)

// WsEvent defines a strictly typed event for WebSocket messages.
type WsEvent struct {
	Event  string  `json:"event"`
	Type   string  `json:"type,omitempty"`
	TaskID *string `json:"task_id,omitempty"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		// Require strict origin matching for safety
		if origin != "" && !strings.HasPrefix(origin, "http://localhost:") && !strings.HasPrefix(origin, "https://localhost:") {
			return false
		}
		return true
	},
}

var (
	clients      = make(map[*websocket.Conn]bool)
	clientsMutex sync.Mutex
)

// WsHandler handles incoming WebSocket connections.
func WsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("[WS] Error upgrading to WebSocket:", err)
		return
	}
	defer conn.Close()

	clientsMutex.Lock()
	clients[conn] = true
	clientsMutex.Unlock()

	log.Println("[WS] Client connected")

	// Keep connection alive and listen for close
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			log.Println("[WS] Client disconnected or error:", err)
			clientsMutex.Lock()
			delete(clients, conn)
			clientsMutex.Unlock()
			break
		}
	}
}

// BroadcastEvent sends a JSON message to all connected clients.
func BroadcastEvent(event WsEvent) {
	message, err := json.Marshal(event)
	if err != nil {
		log.Println("[WS] Error marshaling event:", err)
		return
	}

	clientsMutex.Lock()
	var activeClients []*websocket.Conn
	for client := range clients {
		activeClients = append(activeClients, client)
	}
	clientsMutex.Unlock()

	for _, client := range activeClients {
		err := client.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("[WS] Error writing to client: %v", err)
			clientsMutex.Lock()
			client.Close()
			delete(clients, client)
			clientsMutex.Unlock()
		}
	}
}
