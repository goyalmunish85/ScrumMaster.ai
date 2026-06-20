package models

import (
	"time"

	"gorm.io/gorm"
)

type TaskStatus string

const (
	StatusDraft         TaskStatus = "DRAFT"
	StatusInProgress    TaskStatus = "IN_PROGRESS"
	StatusBlocked       TaskStatus = "BLOCKED"
	StatusNeedsInfo     TaskStatus = "NEEDS_INFO"
	StatusDelayed       TaskStatus = "DELAYED"
	StatusDone          TaskStatus = "DONE"
	StatusDeprioritized TaskStatus = "DEPRIORITIZED"
)

type Task struct {
	ID              string     `gorm:"type:uuid;primaryKey" json:"id"`
	Title           string     `gorm:"not null" json:"title"`
	Description     string     `json:"description"`
	Status          TaskStatus `gorm:"default:'DRAFT'" json:"status"`
	Priority        string     `json:"priority"`
	Labels          string     `json:"labels"`
	Assignee        string     `json:"assignee"`
	Reporter        string     `json:"reporter"`
	Project         string     `json:"project"`
	Client          string     `json:"client"`
	JiraKey         string     `gorm:"index" json:"jira_key"`
	SlackReference  string     `json:"slack_reference"`
	GitlabReference string     `json:"gitlab_reference"`
	SheetReference  string     `json:"sheet_reference"`

	// Rich Metadata
	Team       string `json:"team"`
	TaskType   string `json:"task_type"`
	Sprint     string `json:"sprint"`
	ParentKey  string `json:"parent_key"`
	SourceName string `json:"source_name"`

	DueDate   *time.Time     `json:"due_date"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type SyncLog struct {
	ID        string    `gorm:"type:uuid;primaryKey" json:"id"`
	Platform  string    `gorm:"not null" json:"platform"`  // "slack", "jira", "sheets"
	TargetID  string    `gorm:"not null" json:"target_id"` // e.g. "SAAS", "C0AQMS8J0P3"
	Status    string    `gorm:"not null" json:"status"`    // "SUCCESS" or "ERROR"
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type TaskDependency struct {
	ID             string    `gorm:"type:uuid;primaryKey" json:"id"`
	BlockedTaskID  string    `gorm:"type:uuid;not null;index" json:"blocked_task_id"`
	BlockingTaskID string    `gorm:"type:uuid;not null;index" json:"blocking_task_id"`
	CreatedAt      time.Time `json:"created_at"`
}

type OperationalEventRecord struct {
	ID        string    `gorm:"type:uuid;primaryKey" json:"id"`
	TaskID    *string   `gorm:"type:uuid;index" json:"task_id"` // Nullable if the event is a general info update
	EventType string    `gorm:"not null" json:"event_type"`
	Payload   string    `gorm:"type:json" json:"payload"`
	CreatedAt time.Time `json:"created_at"`
}

type ChatMessage struct {
	ID        string    `gorm:"type:uuid;primaryKey" json:"id"`
	Content   string    `gorm:"not null" json:"content"`
	SenderID  string    `gorm:"not null" json:"sender_id"`
	Role      string    `gorm:"not null" json:"role"` // 'user' or 'ai'
	CreatedAt time.Time `json:"created_at"`
}

type CronLog struct {
	ID      string    `gorm:"type:uuid;primaryKey" json:"id"`
	JobName string    `json:"job_name"`
	RunAt   time.Time `json:"run_at"`
	Status  string    `json:"status"`
}

type SlackSyncState struct {
	ChannelID     string `gorm:"primaryKey" json:"channel_id"`
	LastTimestamp string `json:"last_timestamp"`
}

type IntegrationTarget struct {
	ID        string    `gorm:"type:uuid;primaryKey" json:"id"`
	Platform  string    `gorm:"not null;index" json:"platform"` // e.g., "slack", "sheets", "jira"
	TargetID  string    `gorm:"not null" json:"target_id"`      // channel id, sheet id, project key
	CreatedAt time.Time `json:"created_at"`
}

type AIEvaluation struct {
	ID           string    `gorm:"type:uuid;primaryKey" json:"id"`
	MessageID    string    `gorm:"type:uuid;not null;index" json:"message_id"` // Which AI message was evaluated
	FeedbackText string    `gorm:"not null" json:"feedback_text"`              // "Stop using bullet points", "Be more concise"
	CreatedAt    time.Time `json:"created_at"`
}
