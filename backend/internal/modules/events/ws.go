package events

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for the MVP/dashboard integration.
		// In production, this should validate against a list of allowed origins.
		return true
	},
}

// Global state to manage connected WebSocket clients safely.
var (
	clients      = make(map[*websocket.Conn]bool)
	clientsMutex sync.Mutex
)

// WSHandler handles incoming WebSocket connections.
func WSHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WebSocket] Upgrade failed: %v", err)
		return
	}

	clientsMutex.Lock()
	clients[conn] = true
	clientsMutex.Unlock()

	log.Printf("[WebSocket] Client connected: %v", conn.RemoteAddr())

	// Listen for incoming messages or closures
	go func() {
		defer func() {
			clientsMutex.Lock()
			delete(clients, conn)
			clientsMutex.Unlock()
			conn.Close()
			log.Printf("[WebSocket] Client disconnected: %v", conn.RemoteAddr())
		}()

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("[WebSocket] Unexpected close error: %v", err)
				}
				break
			}
		}
	}()
}

// BroadcastTaskUpdated sends a minimal payload to all connected clients.
func BroadcastTaskUpdated() {
	payload, err := json.Marshal(map[string]string{
		"event": "task_updated",
	})
	if err != nil {
		log.Printf("[WebSocket] Failed to marshal broadcast payload: %v", err)
		return
	}

	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	for conn := range clients {
		if err := conn.WriteMessage(websocket.TextMessage, payload); err != nil {
			log.Printf("[WebSocket] Failed to send message to client %v: %v", conn.RemoteAddr(), err)
			conn.Close()
			delete(clients, conn)
		}
	}
}
