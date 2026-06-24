package main

import (
	"log"
	"time"

	"github.com/aios/backend/internal/db"
	"github.com/aios/backend/internal/models"
)

func main() {
	db.InitDB()

	// Add some mock tasks
	tasks := []models.Task{
		{ID: "task-1", Title: "Fix navigation bug", Status: "DONE", Priority: "High", Assignee: "Alice", DueDate: timePtr(time.Now().Add(24 * time.Hour)), JiraKey: "SAAS-1"},
		{ID: "task-2", Title: "Update API docs", Status: "IN_PROGRESS", Priority: "Medium", Assignee: "Bob", JiraKey: "SAAS-2"},
		{ID: "task-3", Title: "Deploy to production", Status: "BLOCKED", Priority: "High", Assignee: "Alice", JiraKey: "SAAS-3"},
	}

	for _, t := range tasks {
		if err := db.DB.Create(&t).Error; err != nil {
			log.Printf("Failed to insert task: %v", err)
		}
	}
	log.Println("Mock tasks inserted successfully.")
}

func timePtr(t time.Time) *time.Time {
	return &t
}
