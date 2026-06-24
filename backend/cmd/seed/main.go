package main

import (
	"log"
	"time"

	"github.com/aios/backend/internal/db"
	"github.com/aios/backend/internal/models"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load("../../.env")
	db.InitDB()

	events := []models.OperationalEventRecord{
		{
			ID:        uuid.New().String(),
			EventType: "TASK_CREATED",
			Payload:   `{"task_name": "Update login page", "assignee": "alice@example.com"}`,
			CreatedAt: time.Now().Add(-10 * time.Minute),
		},
		{
			ID:        uuid.New().String(),
			EventType: "TASK_ASSIGNED",
			Payload:   `{"task_name": "Update login page", "assignee": "bob@example.com"}`,
			CreatedAt: time.Now().Add(-5 * time.Minute),
		},
		{
			ID:        uuid.New().String(),
			EventType: "TASK_COMPLETED",
			Payload:   `{"task_name": "Update login page"}`,
			CreatedAt: time.Now().Add(-1 * time.Minute),
		},
	}

	for _, e := range events {
		if err := db.DB.Create(&e).Error; err != nil {
			log.Fatal(err)
		}
	}
	log.Println("Seeded events successfully!")
}
