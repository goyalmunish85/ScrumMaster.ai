package ws

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for MVP
	},
}

var (
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
)

// HandleConnections upgrades the HTTP connection to a WebSocket connection and registers the client.
func HandleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[WS ERROR] Upgrade failed: %v", err)
		return
	}

	clientsMu.Lock()
	clients[ws] = true
	clientsMu.Unlock()
	log.Println("[WS] Client connected")

	// Read loop to detect disconnects
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			log.Printf("[WS] Client disconnected: %v", err)
			clientsMu.Lock()
			delete(clients, ws)
			clientsMu.Unlock()
			ws.Close()
			break
		}
	}
}

// Broadcast sends a message to all connected clients.
func Broadcast(message []byte) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("[WS ERROR] Failed to send message, closing connection: %v", err)
			client.Close()
			delete(clients, client)
		}
	}
}
