package ops

import (
	"fmt"
	"log"
	"time"

	"github.com/aios/backend/internal/db"
	"github.com/aios/backend/internal/models"
	"github.com/google/uuid"
)

// StartCronEngine runs in the background and executes scheduled jobs
func StartCronEngine() {
	log.Println("[CRON] Starting Cron Engine...")

	// Initial check on startup (wait a few seconds for DB to settle)
	time.Sleep(5 * time.Second)

	runDailyBriefing()
	runWeeklyReport()

	// Keep checking every hour
	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		runDailyBriefing()
		runWeeklyReport()
	}
}

func runDailyBriefing() {
	// Check if a Daily Briefing was already generated today
	var lastLog models.CronLog
	err := db.DB.Where("job_name = ?", "daily_briefing").Order("run_at desc").First(&lastLog).Error

	if err == nil && time.Since(lastLog.RunAt) < 24*time.Hour {
		// Already ran today
		return
	}

	log.Println("[CRON] Generating Daily Briefing...")

	var overdueTasks []models.Task
	var blockedTasks []models.Task

	now := time.Now()
	// Find overdue (not done and due before now)
	db.DB.Where("status != ? AND status != ? AND due_date < ?", "DONE", "DEPRIORITIZED", now).Find(&overdueTasks)

	// Find blocked
	db.DB.Where("status = ?", "BLOCKED").Find(&blockedTasks)

	if len(overdueTasks) == 0 && len(blockedTasks) == 0 {
		log.Println("[CRON] No blockers or overdue tasks. Skipping daily briefing.")
		recordCronLog("daily_briefing", "SKIPPED")
		return
	}

	content := "### 🌅 Good morning! Here is your Daily Briefing:\n\n"

	if len(blockedTasks) > 0 {
		content += "🚨 **CRITICAL BLOCKERS** 🚨\n"
		for _, t := range blockedTasks {
			content += fmt.Sprintf("- **[%s] %s**: %s\n", t.JiraKey, t.Title, t.Assignee)
		}
		content += "\n"
	}

	if len(overdueTasks) > 0 {
		content += "⏰ **OVERDUE TASKS** ⏰\n"
		for _, t := range overdueTasks {
			content += fmt.Sprintf("- **[%s] %s**: Due on %s\n", t.JiraKey, t.Title, t.DueDate.Format("2006-01-02"))
		}
	}

	// Insert the message directly into chat history as an AI system message
	msg := models.ChatMessage{
		ID:        uuid.New().String(),
		Content:   content,
		SenderID:  "ai-system",
		Role:      "ai",
		CreatedAt: time.Now(),
	}
	db.DB.Create(&msg)

	recordCronLog("daily_briefing", "SUCCESS")
}

func runWeeklyReport() {
	// Check if a Weekly Report was already generated this week
	var lastLog models.CronLog
	err := db.DB.Where("job_name = ?", "weekly_report").Order("run_at desc").First(&lastLog).Error

	if err == nil && time.Since(lastLog.RunAt) < 7*24*time.Hour {
		// Already ran this week
		return
	}

	log.Println("[CRON] Generating Weekly Report Alert...")

	content := "📅 **Weekly Stakeholder Update Ready**\nIt has been 7 days since your last report! You can generate a fresh executive summary by clicking the 'Generate Weekly Report' button in the dashboard."

	msg := models.ChatMessage{
		ID:        uuid.New().String(),
		Content:   content,
		SenderID:  "ai-system",
		Role:      "ai",
		CreatedAt: time.Now(),
	}
	db.DB.Create(&msg)

	recordCronLog("weekly_report", "SUCCESS")
}

func recordCronLog(jobName string, status string) {
	db.DB.Create(&models.CronLog{
		ID:      uuid.New().String(),
		JobName: jobName,
		RunAt:   time.Now(),
		Status:  status,
	})
}
