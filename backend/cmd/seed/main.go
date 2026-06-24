package main

import (
	"github.com/aios/backend/internal/db"
	"github.com/aios/backend/internal/models"
	"time"
	"fmt"
)

func main() {
	db.InitDB()
	var count int64
	db.DB.Model(&models.Task{}).Count(&count)
	if count > 0 {
		fmt.Println("Database already has tasks.")
		return
	}

	now := time.Now()
	nextWeek := now.Add(7 * 24 * time.Hour)

	tasks := []models.Task{
		{
			ID:       "00000000-0000-0000-0000-000000000001",
			Title:    "Setup Qdrant Connection",
			Status:   models.StatusDone,
			Assignee: "Jules",
			Priority: "High",
			DueDate:  &now,
			CreatedAt: now,
			UpdatedAt: now,
			Team: "Backend",
			Client: "Internal",
		},
		{
			ID:       "00000000-0000-0000-0000-000000000002",
			Title:    "Refactor React Frontend Components",
			Status:   models.StatusInProgress,
			Assignee: "Jules",
			Priority: "Medium",
			DueDate:  &nextWeek,
			CreatedAt: now,
			UpdatedAt: now,
			Team: "Frontend",
			Client: "External",
			SourceName: "Jira:PROJ-123",
		},
		{
			ID:       "00000000-0000-0000-0000-000000000003",
			Title:    "Investigate Playwright Timeout Issue",
			Status:   models.StatusBlocked,
			Assignee: "",
			Priority: "High",
			CreatedAt: now,
			UpdatedAt: now,
			Team: "QA",
			Client: "Internal",
		},
	}

	for _, t := range tasks {
		db.DB.Create(&t)
	}
	fmt.Println("Tasks seeded successfully.")
}
