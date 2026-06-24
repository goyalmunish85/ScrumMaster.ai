package db

import (
	"fmt"
	"time"

	"github.com/aios/backend/internal/models"
)

// SeedTasks creates dummy tasks so we have something to see in the frontend
func SeedTasks() {
	var count int64
	DB.Model(&models.Task{}).Count(&count)
	if count > 0 {
		return
	}

	fmt.Println("Seeding tasks...")
	now := time.Now()
	nextWeek := now.Add(7 * 24 * time.Hour)

	tasks := []models.Task{
		{
			ID:       "uuid-1",
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
			ID:       "uuid-2",
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
			ID:       "uuid-3",
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
		DB.Create(&t)
	}
}
