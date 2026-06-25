package main

import (
	"fmt"
	"log"
	"github.com/aios/backend/internal/db"
	"github.com/aios/backend/internal/models"
	"github.com/google/uuid"
)

func main() {
	db.InitDB()

	task := models.Task{
		ID:          uuid.New().String(),
		Title:       "Implement Auth",
		Description: "Add JWT authentication to the platform.",
		Status:      models.StatusInProgress,
		Assignee:    "John Doe",
		Priority:    "High",
		JiraKey:     "PROJ-123",
	}
	result := db.DB.Create(&task)
	if result.Error != nil {
		log.Fatalf("failed to seed task: %v", result.Error)
	}
	fmt.Printf("Seeded task %s\n", task.ID)
}
