package ops

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aios/backend/internal/db"
	"github.com/aios/backend/internal/models"
	"github.com/aios/backend/internal/modules/integrations"
	"github.com/google/uuid"
)

// StartCronEngine runs in the background and executes scheduled jobs (Agent Loops)
func StartCronEngine() {
	log.Println("[CRON] Starting Agent Cron Engine...")

	// Initial check on startup (wait a few seconds for DB to settle)
	time.Sleep(5 * time.Second)

	// Tier 2: Heavy-Lifter Agents (Run once a day on startup or interval)
	runDailyStandupAgent()
	runWeeklyReport()

	// Tier 1: High-Frequency Agents
	// Ticker for Slack polling every 5 minutes
	go func() {
		slackTicker := time.NewTicker(5 * time.Minute)
		for range slackTicker.C {
			runSlackListenerAgent()
		}
	}()

	// Ticker for Excel/Jira syncing every 15 minutes
	go func() {
		syncTicker := time.NewTicker(15 * time.Minute)
		for range syncTicker.C {
			runExcelSyncAgent()
			runJiraManagerAgent()
		}
	}()

	// Keep checking Daily/Weekly agents every hour
	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		runDailyStandupAgent()
		runWeeklyReport()
	}
}

// ---------------------------------------------------------
// TIER 1: High-Frequency Agents
// ---------------------------------------------------------

// runSlackListenerAgent uses Groq/Cerebras to poll Slack channels for action items
func runSlackListenerAgent() {
	log.Println("[AGENT: Tier 1] Slack Listener Agent starting...")

	channelsEnv := os.Getenv("SLACK_LISTENER_CHANNELS")
	if channelsEnv == "" {
		log.Println("[AGENT: Tier 1] SLACK_LISTENER_CHANNELS not set. Skipping.")
		return
	}

	channels := strings.Split(channelsEnv, ",")
	for _, ch := range channels {
		ch = strings.TrimSpace(ch)
		if ch == "" {
			continue
		}
		log.Printf("[AGENT: Tier 1] Syncing Slack channel %s...", ch)
		transcript, err := integrations.SyncSlackChannel(ch)
		if err != nil {
			log.Printf("[AGENT: Tier 1] Error syncing channel %s: %v", ch, err)
		} else {
			log.Printf("[AGENT: Tier 1] Successfully fetched transcript for %s (len: %d). Events emitted.", ch, len(transcript))
		}
	}

	log.Println("[AGENT: Tier 1] Slack Listener Agent finished.")
}

// ---------------------------------------------------------
// TIER 2: Heavy-Lifter Agents
// ---------------------------------------------------------

// runExcelSyncAgent reads the Client Excel to pull new requests into the AI OS (V1 is Read-Only)
func runExcelSyncAgent() {
	log.Println("[AGENT: Tier 2] Excel Sync Agent starting...")

	sheetID := os.Getenv("GOOGLE_SHEET_ID")
	if sheetID == "" {
		log.Println("[AGENT: Tier 2] GOOGLE_SHEET_ID not set. Skipping.")
		return
	}

	log.Printf("[AGENT: Tier 2] Syncing Google Sheet %s...", sheetID)
	res, err := integrations.SyncGoogleSheet(sheetID)
	if err != nil {
		log.Printf("[AGENT: Tier 2] Error syncing sheet: %v", err)
	} else {
		log.Printf("[AGENT: Tier 2] %s", res)
	}

	log.Println("[AGENT: Tier 2] Excel Sync Agent finished.")
}

// runJiraManagerAgent scans for 'Approved' tasks and pushes them to Jira, and pulls Jira updates
func runJiraManagerAgent() {
	log.Println("[AGENT: Tier 2] Jira Manager Agent starting...")

	// 1. Pull updates from Jira
	jiraKeysEnv := os.Getenv("JIRA_PROJECT_KEYS")
	if jiraKeysEnv != "" {
		keys := strings.Split(jiraKeysEnv, ",")
		for _, key := range keys {
			key = strings.TrimSpace(key)
			if key == "" {
				continue
			}
			log.Printf("[AGENT: Tier 2] Syncing Jira project %s...", key)
			res, err := integrations.SyncJiraProject(key, false)
			if err != nil {
				log.Printf("[AGENT: Tier 2] Error syncing Jira project %s: %v", key, err)
			} else {
				log.Printf("[AGENT: Tier 2] %s", res)
			}
		}
	}

	// 2. Push approved tasks to Jira
	// [V1 CONSTRAINT] Write operations disabled. AI OS strictly acts as a local source of truth.
	/*
		var approvedTasks []models.Task
		db.DB.Where("status = ? AND jira_key = ?", "APPROVED", "").Find(&approvedTasks)

		for _, task := range approvedTasks {
			log.Printf("Found approved task ready for Jira: %s", task.Title)

			// In a real scenario, Tier 2 Gemini might be used here to expand the task description
			// into a full Jira User Story before posting.

			jiraKey, err := integrations.CreateJiraTicket(task, "AIOS") // Assume "AIOS" is the default project
			if err != nil {
				log.Printf("Error creating Jira ticket for task %s: %v", task.ID, err)
				continue
			}

			if jiraKey != "" && !strings.HasPrefix(jiraKey, "DRYRUN") {
				task.JiraKey = jiraKey
				db.DB.Save(&task)
			}
		}
	*/

	log.Println("[AGENT: Tier 2] Jira Manager Agent finished.")
}

func runDailyStandupAgent() {
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

	// Post to Slack if channel is configured
	// [V1 CONSTRAINT] Write operations disabled. AI OS strictly acts as a local source of truth.
	/*
		standupChannel := os.Getenv("STANDUP_CHANNEL_ID")
		if standupChannel != "" {
			err := integrations.PostSlackMessage(standupChannel, content)
			if err != nil {
				log.Printf("[CRON] Failed to post standup to Slack: %v", err)
			}
		}
	*/

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
