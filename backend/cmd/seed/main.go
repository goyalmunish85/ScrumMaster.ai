package main

import (
	"fmt"
	"time"

	"github.com/aios/backend/internal/db"
	"github.com/aios/backend/internal/models"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	db.InitDB()

	events := []models.OperationalEventRecord{
		{
			ID:        uuid.New().String(),
			EventType: "TASK_CREATED",
			Payload:   `{"task_name": "Implement Timeline", "assignee": "Artemis"}`,
			CreatedAt: time.Now().Add(-1 * time.Hour),
		},
		{
			ID:        uuid.New().String(),
			EventType: "TASK_COMPLETED",
			Payload:   `{"task_name": "Setup Repos"}`,
			CreatedAt: time.Now().Add(-2 * time.Hour),
		},
		{
			ID:        uuid.New().String(),
			EventType: "TASK_BLOCKED",
			Payload:   `{"task_name": "Database Migration", "reason": "Waiting for DBA"}`,
			CreatedAt: time.Now().Add(-3 * time.Hour),
		},
	}

	for _, ev := range events {
		err := db.DB.Create(&ev).Error
		if err != nil {
			fmt.Printf("Error creating event: %v\n", err)
		}
	}
	fmt.Println("Seeded events.")
}
