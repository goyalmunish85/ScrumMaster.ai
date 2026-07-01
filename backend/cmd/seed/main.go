package main

import (
	"log"

	"github.com/aios/backend/internal/db"
	"github.com/aios/backend/internal/models"
	"github.com/google/uuid"
)

func main() {
	db.InitDB()
	seedTasks()
}

func seedTasks() {
	tasks := []models.Task{
		{
			ID:         uuid.New().String(),
			Title:      "Fix Login Bug",
			Status:     models.StatusBlocked,
			Assignee:   "Alice",
			SourceName: "Jira",
		},
		{
			ID:         uuid.New().String(),
			Title:      "Implement TaskTable",
			Status:     models.StatusInProgress,
			Assignee:   "Bob",
			SourceName: "Notion",
		},
	}

	for _, task := range tasks {
		if err := db.DB.Create(&task).Error; err != nil {
			log.Printf("Failed to seed task %s: %v", task.Title, err)
		} else {
			log.Printf("Seeded task: %s", task.Title)
		}
	}
}
