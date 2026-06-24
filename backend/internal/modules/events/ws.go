package events

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		// Allow specific origins for production security
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:8080",
		}
		for _, o := range allowedOrigins {
			if origin == o {
				return true
			}
		}
		// Also allow if origin is empty (e.g. same-origin request)
		return origin == ""
	},
}

var (
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
)

// WsHandler handles incoming WebSocket connections
func WsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to set websocket upgrade: %v", err)
		return
	}

	clientsMu.Lock()
	clients[conn] = true
	clientsMu.Unlock()

	defer func() {
		clientsMu.Lock()
		delete(clients, conn)
		clientsMu.Unlock()
		conn.Close()
	}()

	// Keep connection alive
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

// BroadcastEvent sends a message to all connected clients
func BroadcastEvent(event string) {
	clientsMu.Lock()
	// Copy clients to local slice to release lock before I/O
	var localClients []*websocket.Conn
	for client := range clients {
		localClients = append(localClients, client)
	}
	clientsMu.Unlock()

	for _, client := range localClients {
		// Optimization: strict timeout boundaries, avoid hanging connections
		client.SetWriteDeadline(time.Now().Add(5 * time.Second))
		err := client.WriteJSON(map[string]string{"event": event})
		if err != nil {
			log.Printf("Websocket error: %v", err)
			client.Close()
			clientsMu.Lock()
			delete(clients, client)
			clientsMu.Unlock()
		}
	}
}
