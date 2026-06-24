package events

import (
	"encoding/json"
	"log"
	"time"

	"github.com/aios/backend/internal/db"
	"github.com/aios/backend/internal/models"
	"github.com/aios/backend/internal/modules/memory"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventType string

const (
	TaskBlocked       EventType = "TASK_BLOCKED"
	TaskAssigned      EventType = "TASK_ASSIGNED"
	TaskCreated       EventType = "TASK_CREATED"
	TaskCompleted     EventType = "TASK_COMPLETED"
	TaskDependency    EventType = "TASK_DEPENDENCY"
	InfoRequest       EventType = "INFO_REQUEST"
	TaskDueDate       EventType = "TASK_DUE_DATE"
	TaskStatusChanged EventType = "TASK_STATUS_CHANGED"
	RiskDetected      EventType = "RISK_DETECTED"
	BulkTasks         EventType = "BULK_TASKS"
)

type OperationalEvent struct {
	Type    EventType       `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// In-Memory Event Bus for MVP (removes immediate Redis dependency for the user)
var eventStream = make(chan OperationalEvent, 100)

func Publish(event OperationalEvent) {
	log.Printf("[EVENT BUS] Publishing Event: %s", event.Type)
	eventStream <- event
}

func StartListener() {
	go func() {
		for event := range eventStream {
			log.Printf("[EVENT BUS] Processed Event: %s | Payload: %s", event.Type, string(event.Payload))

			// 1. Save the raw event for auditing
			eventRecord := models.OperationalEventRecord{
				ID:        uuid.New().String(),
				EventType: string(event.Type),
				Payload:   string(event.Payload),
			}

			// Helper to get or create a task
			getOrCreateTask := func(title string) *models.Task {
				var task models.Task
				if db.DB != nil {
					if err := db.DB.Where("title = ?", title).First(&task).Error; err != nil {
						// Create it
						task = models.Task{
							ID:     uuid.New().String(),
							Title:  title,
							Status: models.StatusDraft,
						}
						db.DB.Create(&task)
					}
					return &task
				}
				return nil
			}

			// 2. Materialize to the relational tables based on event type
			switch event.Type {
			case TaskCreated:
				var payload struct {
					TaskName    string `json:"task_name"`
					Description string `json:"description"`
					Assignee    string `json:"assignee"`
					Client      string `json:"client"`
					DueDate     string `json:"due_date"`
				}
				if err := json.Unmarshal(event.Payload, &payload); err == nil {
					task := getOrCreateTask(payload.TaskName)
					if task != nil {
						updates := map[string]interface{}{}
						if payload.Description != "" {
							updates["description"] = payload.Description
						}
						if payload.Assignee != "" {
							updates["assignee"] = payload.Assignee
						}
						if payload.Client != "" {
							updates["client"] = payload.Client
						}
						if len(updates) > 0 {
							db.DB.Model(task).Updates(updates)
						}
						eventRecord.TaskID = &task.ID
					}
				}
			case TaskAssigned:
				var payload struct {
					TaskName string `json:"task_name"`
					Assignee string `json:"assignee"`
				}
				if err := json.Unmarshal(event.Payload, &payload); err == nil {
					task := getOrCreateTask(payload.TaskName)
					if task != nil {
						db.DB.Model(task).Update("assignee", payload.Assignee)
						eventRecord.TaskID = &task.ID
					}
				}
			case TaskCompleted:
				var payload struct {
					TaskName string `json:"task_name"`
				}
				if err := json.Unmarshal(event.Payload, &payload); err == nil {
					task := getOrCreateTask(payload.TaskName)
					if task != nil {
						db.DB.Model(task).Update("status", models.StatusDone)
						eventRecord.TaskID = &task.ID
					}
				}
			case TaskBlocked:
				var payload struct {
					TaskName string `json:"task_name"`
				}
				if err := json.Unmarshal(event.Payload, &payload); err == nil {
					task := getOrCreateTask(payload.TaskName)
					if task != nil {
						db.DB.Model(task).Update("status", models.StatusBlocked)
						eventRecord.TaskID = &task.ID
					}
				}
			case TaskStatusChanged:
				var payload struct {
					TaskName string `json:"task_name"`
					Status   string `json:"status"`
				}
				if err := json.Unmarshal(event.Payload, &payload); err == nil {
					task := getOrCreateTask(payload.TaskName)
					if task != nil {
						db.DB.Model(task).Update("status", payload.Status)
						eventRecord.TaskID = &task.ID
					}
				}
			case TaskDueDate:
				var payload struct {
					TaskName string `json:"task_name"`
					DueDate  string `json:"due_date"`
				}
				if err := json.Unmarshal(event.Payload, &payload); err == nil {
					task := getOrCreateTask(payload.TaskName)
					if task != nil {
						db.DB.Model(task).Update("description", task.Description+"\nDue: "+payload.DueDate)
						eventRecord.TaskID = &task.ID
					}
				}
			case BulkTasks:
				var payload struct {
					Tasks []struct {
						JiraKey     string `json:"jira_key"`
						Name        string `json:"name"`
						Description string `json:"description"`
						Status      string `json:"status"`
						Priority    string `json:"priority"`
						Labels      string `json:"labels"`
						Assignee    string `json:"assignee"`
						Reporter    string `json:"reporter"`
						Project     string `json:"project"`
						Client      string `json:"client"`
						DueDate     string `json:"due_date"`
						Team        string `json:"team"`
						TaskType    string `json:"task_type"`
						Sprint      string `json:"sprint"`
						ParentKey   string `json:"parent_key"`
						SourceName  string `json:"source_name"`
					} `json:"tasks"`
					ActivityLogs []struct {
						JiraKey    string `json:"jira_key"`
						EventType  string `json:"event_type"`
						Field      string `json:"field"`
						FromString string `json:"from_string"`
						ToString   string `json:"to_string"`
						Author     string `json:"author"`
						CreatedAt  string `json:"created_at"`
					} `json:"activity_logs"`
				}
				if err := json.Unmarshal(event.Payload, &payload); err == nil {
					if db.DB != nil {
						var tasksToUpdate []models.Task

						txErr := db.DB.Transaction(func(tx *gorm.DB) error {
							// Pre-fetch tasks by Jira Key to avoid N+1 queries
							var jiraKeys []string
							var titles []string
							var sourceNames []string
							for _, t := range payload.Tasks {
								if t.JiraKey != "" {
									jiraKeys = append(jiraKeys, t.JiraKey)
								} else if t.Name != "" && t.SourceName != "" {
									titles = append(titles, t.Name)
									sourceNames = append(sourceNames, t.SourceName)
								}
							}

							var existingTasks []models.Task
							if len(jiraKeys) > 0 {
								tx.Where("jira_key IN ?", jiraKeys).Find(&existingTasks)
							}
							var existingTasksFallback []models.Task
							if len(titles) > 0 {
								tx.Where("title IN ? AND source_name IN ?", titles, sourceNames).Find(&existingTasksFallback)
							}
							existingTasks = append(existingTasks, existingTasksFallback...)

							existingTaskMap := make(map[string]*models.Task)
							existingTaskFallbackMap := make(map[string]*models.Task)
							for i, task := range existingTasks {
								if task.JiraKey != "" {
									existingTaskMap[task.JiraKey] = &existingTasks[i]
								} else {
									existingTaskFallbackMap[task.Title+"_"+task.SourceName] = &existingTasks[i]
								}
							}

							for _, t := range payload.Tasks {
								existingTask, ok := existingTaskMap[t.JiraKey]
								if !ok && t.JiraKey == "" {
									existingTask, ok = existingTaskFallbackMap[t.Name+"_"+t.SourceName]
								}

								if ok {
									// Task already exists, update fields
									updates := map[string]interface{}{
										"status":   t.Status,
										"priority": t.Priority,
										"labels":   t.Labels,
										"reporter": t.Reporter,
										"project":  t.Project,
									}
									// Only update description, client, assignee, due_date if they are not empty
									if t.Description != "" {
										updates["description"] = t.Description
									}
									if t.Client != "" {
										updates["client"] = t.Client
									}
									if t.Assignee != "" {
										updates["assignee"] = t.Assignee
									}
									if t.Team != "" {
										updates["team"] = t.Team
									}
									if t.TaskType != "" {
										updates["task_type"] = t.TaskType
									}
									if t.Sprint != "" {
										updates["sprint"] = t.Sprint
									}
									if t.ParentKey != "" {
										updates["parent_key"] = t.ParentKey
									}
									if t.SourceName != "" {
										updates["source_name"] = t.SourceName
									}

									if t.DueDate != "" {
										if pd, err := time.Parse("2 Jan 2006", t.DueDate); err == nil {
											updates["due_date"] = pd
										}
									}
									if existingTask.JiraKey == "" && t.JiraKey != "" {
										updates["jira_key"] = t.JiraKey
									}
									if err := tx.Model(existingTask).Updates(updates).Error; err != nil {
										return err
									}

									tasksToUpdate = append(tasksToUpdate, *existingTask)
									continue
								}

								// Parse DueDate if possible
								var dd *time.Time
								if t.DueDate != "" {
									if pd, err := time.Parse("2 Jan 2006", t.DueDate); err == nil {
										dd = &pd
									}
								}

								// Create new task
								newTask := models.Task{
									ID:          uuid.New().String(),
									JiraKey:     t.JiraKey,
									Title:       t.Name,
									Description: t.Description,
									Status:      models.TaskStatus(t.Status),
									Priority:    t.Priority,
									Labels:      t.Labels,
									Assignee:    t.Assignee,
									Reporter:    t.Reporter,
									Project:     t.Project,
									Client:      t.Client,
									Team:        t.Team,
									TaskType:    t.TaskType,
									Sprint:      t.Sprint,
									ParentKey:   t.ParentKey,
									SourceName:  t.SourceName,
									DueDate:     dd,
								}
								if err := tx.Create(&newTask).Error; err != nil {
									return err
								}
								if t.JiraKey != "" {
									existingTaskMap[t.JiraKey] = &newTask
								} else {
									existingTaskFallbackMap[t.Name+"_"+t.SourceName] = &newTask
								}
								tasksToUpdate = append(tasksToUpdate, newTask)
							}

							// Process Activity Logs
							if len(payload.ActivityLogs) > 0 {
								var logTasks []string
								for _, l := range payload.ActivityLogs {
									logTasks = append(logTasks, l.JiraKey)
								}

								// Pre-fetch activity logs to prevent duplicate entries
								var existingLogs []models.ActivityLog
								tx.Where("jira_key IN ?", logTasks).Find(&existingLogs)

								existingLogsMap := make(map[string]bool)
								for _, l := range existingLogs {
									// using a simple hash to prevent dups
									key := l.JiraKey + "_" + l.CreatedAt.Format(time.RFC3339) + "_" + l.Field + "_" + l.ToString
									existingLogsMap[key] = true
								}

								var newLogs []models.ActivityLog
								for _, logEntry := range payload.ActivityLogs {
									parsedTime, _ := time.Parse("2006-01-02T15:04:05.000-0700", logEntry.CreatedAt)

									key := logEntry.JiraKey + "_" + parsedTime.Format(time.RFC3339) + "_" + logEntry.Field + "_" + logEntry.ToString
									if !existingLogsMap[key] {
										taskID := ""
										if task, ok := existingTaskMap[logEntry.JiraKey]; ok {
											taskID = task.ID
										}

										if taskID != "" {
											newLogs = append(newLogs, models.ActivityLog{
												ID:         uuid.New().String(),
												TaskID:     taskID,
												JiraKey:    logEntry.JiraKey,
												EventType:  logEntry.EventType,
												Field:      logEntry.Field,
												FromString: logEntry.FromString,
												ToString:   logEntry.ToString,
												Author:     logEntry.Author,
												CreatedAt:  parsedTime,
											})
											existingLogsMap[key] = true
										}
									}
								}
								if len(newLogs) > 0 {
									if err := tx.CreateInBatches(newLogs, 100).Error; err != nil {
										return err
									}
								}
							}

							return nil
						})

						if txErr != nil {
							log.Printf("[EVENT BUS] Transaction failed: %v", txErr)
						} else {
							for _, t := range tasksToUpdate {
								go func(task models.Task) {
									memory.UpsertTaskToQdrant(&task)
								}(t)
							}
						}
					}
				}
			}

			// Save the audit record
			if db.DB != nil {
				db.DB.Create(&eventRecord)
			}
		}
	}()
}
