package main

import (
	"log"
	"net/http"
	"os"

	"github.com/aios/backend/internal/db"
	"github.com/aios/backend/internal/modules/chat"
	"github.com/aios/backend/internal/modules/events"
	"github.com/aios/backend/internal/modules/identity"
	"github.com/aios/backend/internal/modules/integrations"
	"github.com/aios/backend/internal/middleware"
	"github.com/aios/backend/internal/modules/ops"
	"github.com/aios/backend/internal/modules/reports"
	"github.com/aios/backend/internal/modules/tasks"
	"github.com/joho/godotenv"
)

func main() {
	// Attempt to load .env file; ignore error if it doesn't exist
	_ = godotenv.Load()

	// Initialize the Database
	db.InitDB()
	db.InitQdrant()

	// Start internal event bus listener
	events.StartListener()

	// Start proactive Cron Engine
	go ops.StartCronEngine()

	mux := http.NewServeMux()

	// Identity/Auth Routes
	mux.HandleFunc("/api/activities", events.GetActivitiesHandler)

	mux.HandleFunc("/api/v1/auth/login", identity.LoginHandler)

	// Chat Routes
	mux.HandleFunc("/api/v1/chat/messages", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			chat.GetMessagesHandler(w, r)
		} else if r.Method == http.MethodPost {
			chat.SendMessageHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/v1/chat/evaluate", chat.EvaluateMessageHandler)

	mux.HandleFunc("/api/v1/integrations/sync", integrations.SyncIntegrationsHandler)
	mux.HandleFunc("/api/v1/integrations/logs", integrations.GetSyncLogsHandler)
	mux.HandleFunc("/api/v1/integrations/targets", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			integrations.GetTargetsHandler(w, r)
		case http.MethodPost:
			integrations.AddTargetHandler(w, r)
		case http.MethodDelete:
			integrations.DeleteTargetHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/tasks", tasks.GetTasksHandler)
	mux.HandleFunc("/api/v1/tasks/export", tasks.ExportTasksHandler)

	mux.HandleFunc("/api/v1/reports/weekly", reports.GenerateWeeklyReportHandler)

	// Simple CORS Middleware
	corsHandler := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	// Initialize Rate Limiter: 10 requests per second, burst size 20
	rateLimiter := middleware.NewRateLimiter(10, 20)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting AI-OS API server on port %s", port)

	// Apply rate limiter, then CORS
	finalHandler := corsHandler(rateLimiter.Handler(mux))

	if err := http.ListenAndServe(":"+port, finalHandler); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
