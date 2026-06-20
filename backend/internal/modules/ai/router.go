package ai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aios/backend/internal/db"
	"github.com/aios/backend/internal/models"
	"github.com/aios/backend/internal/modules/events"
	"github.com/aios/backend/internal/modules/memory"
	"github.com/google/uuid"
	openai "github.com/sashabaranov/go-openai"
)

type Router struct {
	openrouterKey string
	deepseekKey   string
	geminiKey     string
	groqKey       string
	edenKey       string
}

// InitRouter initializes the 4-Tier AI Router
func InitRouter() *Router {
	return &Router{
		openrouterKey: os.Getenv("OPENROUTER_API_KEY"),
		deepseekKey:   os.Getenv("DEEPSEEK_API_KEY"),
		geminiKey:     os.Getenv("GEMINI_API_KEY"),
		groqKey:       os.Getenv("GROQ_API_KEY"),
		edenKey:       os.Getenv("EDEN_API_KEY"),
	}
}

type Provider struct {
	BaseURL string
	Key     string
	Model   string
	Name    string
}

func executeWithFallback(ctx context.Context, req openai.ChatCompletionRequest, providers []Provider) (openai.ChatCompletionResponse, error) {
	var lastErr error
	for _, p := range providers {
		if p.Key == "" {
			continue
		}
		log.Printf("[AI] Trying provider: %s (Model: %s)", p.Name, p.Model)
		cfg := openai.DefaultConfig(p.Key)
		cfg.BaseURL = p.BaseURL
		client := openai.NewClientWithConfig(cfg)
		req.Model = p.Model

		resp, err := client.CreateChatCompletion(ctx, req)
		if err == nil {
			return resp, nil
		}
		log.Printf("[AI] Provider %s failed: %v. Falling back...", p.Name, err)
		lastErr = err
	}
	return openai.ChatCompletionResponse{}, fmt.Errorf("all providers failed, last error: %v", lastErr)
}

// 1. Chat + Extraction (The Parser)
// Strictly handles Tool Calling to parse JSON events from the unstructured chat.
func (r *Router) Extract(ctx context.Context, history []map[string]string) ([]events.OperationalEvent, error) {
	providers := []Provider{
		{BaseURL: "https://api.edenai.run/v3", Key: r.edenKey, Model: "openai/gpt-4o-mini", Name: "EdenAI"},
		{BaseURL: "https://openrouter.ai/api/v1", Key: r.openrouterKey, Model: "google/gemini-2.0-flash-lite-001", Name: "OpenRouter"},
		{BaseURL: "https://generativelanguage.googleapis.com/v1beta/openai/", Key: r.geminiKey, Model: "gemini-1.5-flash", Name: "Gemini"},
		{BaseURL: "https://api.groq.com/openai/v1", Key: r.groqKey, Model: "llama-3.3-70b-versatile", Name: "Groq"},
	}

	tools := []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "log_bulk_tasks",
				Description: "Log a large list of tasks at once. Use this when the user pastes a spreadsheet, table, or long list of items.",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"tasks": {
							"type": "array",
							"items": {
								"type": "object",
								"properties": {
									"name": { "type": "string" },
									"status": { "type": "string", "description": "DRAFT, IN_PROGRESS, DONE, BLOCKED" },
									"description": { "type": "string", "description": "Detailed requirements or summary." },
									"assignee": { "type": "string", "description": "Who the task is assigned to." },
									"client": { "type": "string", "description": "Client name, often inferred from the channel name or sheet name." },
									"due_date": { "type": "string", "description": "When the task is due." }
								},
								"required": ["name", "status"]
							}
						}
					},
					"required": ["tasks"]
				}`),
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "log_task_blocked",
				Description: "Log that a task is blocked and cannot proceed.",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"task_name": { "type": "string", "description": "The name of the blocked task." },
						"reason": { "type": "string", "description": "The exact reason why it is blocked." }
					},
					"required": ["task_name", "reason"]
				}`),
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "log_task_assigned",
				Description: "Log that a task has been assigned to a person.",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"task_name": { "type": "string" },
						"assignee": { "type": "string" }
					},
					"required": ["task_name", "assignee"]
				}`),
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "log_task_created",
				Description: "Log that a new task or list of tasks has been created by the user.",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"task_name": { "type": "string", "description": "The title or description of the new task." },
						"description": { "type": "string" },
						"assignee": { "type": "string" },
						"client": { "type": "string" },
						"due_date": { "type": "string" }
					},
					"required": ["task_name"]
				}`),
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "log_task_completed",
				Description: "Log that an existing task has been completed.",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"task_name": { "type": "string", "description": "The name of the completed task." }
					},
					"required": ["task_name"]
				}`),
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "log_task_dependency",
				Description: "Log that a task is blocked by or depends on another task.",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"blocked_task": { "type": "string", "description": "The task that is blocked." },
						"blocking_task": { "type": "string", "description": "The task that is blocking it." }
					},
					"required": ["blocked_task", "blocking_task"]
				}`),
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "log_info_request",
				Description: "Log that a task requires more information or that info was received.",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"task_name": { "type": "string" },
						"info_needed": { "type": "string", "description": "What info is needed or was received." }
					},
					"required": ["task_name", "info_needed"]
				}`),
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "log_task_due_date",
				Description: "Log or update the due date or deadline for a task.",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"task_name": { "type": "string" },
						"due_date": { "type": "string", "description": "The due date, e.g. 'Friday', 'next week', '12th Oct'." }
					},
					"required": ["task_name", "due_date"]
				}`),
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "log_task_status_change",
				Description: "Log that a task status changed to IN_PROGRESS, DELAYED, DEPRIORITIZED.",
				Parameters: json.RawMessage(`{
					"type": "object",
					"properties": {
						"task_name": { "type": "string" },
						"status": { "type": "string", "description": "IN_PROGRESS, DELAYED, DEPRIORITIZED" }
					},
					"required": ["task_name", "status"]
				}`),
			},
		},
	}

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "You are the silent extraction brain for an Operational OS. Extract operational events from the user's message using the provided tools. IMPORTANT: When extracting 'description', extract the exact full text verbatim. NEVER summarize it and NEVER truncate it with '...'. If no operational intent exists, do not call any tools. NEVER reply with normal text, only use tools.",
		},
	}

	for _, msg := range history {
		role := openai.ChatMessageRoleUser
		if msg["role"] == "ai" {
			role = openai.ChatMessageRoleAssistant
		}
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    role,
			Content: msg["content"],
		})
	}

	req := openai.ChatCompletionRequest{
		Messages:    messages,
		Tools:       tools,
		MaxTokens:   500,
		Temperature: 0.1,
	}
	resp, err := executeWithFallback(ctx, req, providers)

	if err != nil {
		return nil, err
	}

	choice := resp.Choices[0]
	var extractedEvents []events.OperationalEvent

	for _, toolCall := range choice.Message.ToolCalls {
		switch toolCall.Function.Name {
		case "log_bulk_tasks":
			extractedEvents = append(extractedEvents, events.OperationalEvent{Type: events.BulkTasks, Payload: []byte(toolCall.Function.Arguments)})
		case "log_task_blocked":
			extractedEvents = append(extractedEvents, events.OperationalEvent{Type: events.TaskBlocked, Payload: []byte(toolCall.Function.Arguments)})
		case "log_task_assigned":
			extractedEvents = append(extractedEvents, events.OperationalEvent{Type: events.TaskAssigned, Payload: []byte(toolCall.Function.Arguments)})
		case "log_task_created":
			extractedEvents = append(extractedEvents, events.OperationalEvent{Type: events.TaskCreated, Payload: []byte(toolCall.Function.Arguments)})
		case "log_task_completed":
			extractedEvents = append(extractedEvents, events.OperationalEvent{Type: events.TaskCompleted, Payload: []byte(toolCall.Function.Arguments)})
		case "log_task_dependency":
			extractedEvents = append(extractedEvents, events.OperationalEvent{Type: events.TaskDependency, Payload: []byte(toolCall.Function.Arguments)})
		case "log_info_request":
			extractedEvents = append(extractedEvents, events.OperationalEvent{Type: events.InfoRequest, Payload: []byte(toolCall.Function.Arguments)})
		case "log_task_due_date":
			extractedEvents = append(extractedEvents, events.OperationalEvent{Type: events.TaskDueDate, Payload: []byte(toolCall.Function.Arguments)})
		case "log_task_status_change":
			extractedEvents = append(extractedEvents, events.OperationalEvent{Type: events.TaskStatusChanged, Payload: []byte(toolCall.Function.Arguments)})
		}
	}

	return extractedEvents, nil
}

// 2. Fast Interactions (The Voice)
// Instantly generates the natural language UI response.
func (r *Router) Converse(ctx context.Context, history []map[string]string, extractedEvents []events.OperationalEvent) (string, error) {
	providers := []Provider{
		{BaseURL: "https://api.edenai.run/v3", Key: r.edenKey, Model: "openai/gpt-4o-mini", Name: "EdenAI"},
		{BaseURL: "https://openrouter.ai/api/v1", Key: r.openrouterKey, Model: "google/gemini-2.0-flash-lite-001", Name: "OpenRouter"},
		{BaseURL: "https://generativelanguage.googleapis.com/v1beta/openai/", Key: r.geminiKey, Model: "gemini-1.5-flash", Name: "Gemini"},
		{BaseURL: "https://api.groq.com/openai/v1", Key: r.groqKey, Model: "llama-3.3-70b-versatile", Name: "Groq"},
	}

	if len(providers) == 0 {
		if len(extractedEvents) > 0 {
			return "I've successfully logged the operational events from your update.", nil
		}
		return "Message received.", nil
	}
	// No context stuffing! The AI will use the search_tasks tool dynamically.

	// Pull past evaluations to self-improve
	var evals []models.AIEvaluation
	db.DB.Limit(10).Order("created_at desc").Find(&evals)

	evaluationContext := ""
	if len(evals) > 0 {
		evaluationContext = "\n\nCRITICAL USER FEEDBACK YOU MUST FOLLOW:\n"
		for _, e := range evals {
			evaluationContext += "- " + e.FeedbackText + "\n"
		}
	}

	extractedSummary := ""
	if len(extractedEvents) > 0 {
		extractedSummary = "Behind the scenes, the system automatically logged the following operational events based on the user's update: "
		for _, ev := range extractedEvents {
			extractedSummary += string(ev.Type) + ", "
		}
		extractedSummary += "Acknowledge what was done naturally and concisely."
	}

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "You are Antigravity OS, an AI Operational Execution assistant. You are deeply integrated with the user's Slack, Jira, GitLab, and Google Sheets. When the user clicks 'Sync Integrations', your backend automatically pulls tickets and messages from those platforms, and extracts tasks into your database.\n\nCRITICAL RULE 1: NEVER say you don't know something without first using the `search_tasks` tool to query the database! You MUST use the `search_tasks` tool to retrieve relevant tasks before answering queries about tasks. The database is large, so use specific keywords or client names to filter.\nCRITICAL RULE 2: DO NOT say you cannot access Jira, Slack or external systems. You ALREADY have the data synced into your database! When you use the `search_tasks` tool, look at the `JiraKey` or `Cross-Platform` tags in the results to tell the user if a task is from Jira or Slack.\nCRITICAL RULE 3: You are the definitive Lead Project Manager. If the user asks if a Sheet task is on Jira, and your search returns NO Jira key for it, you MUST state definitively with authority: 'No, this task is not currently tracked on Jira. It is only in the Sheet.' Do NOT hedge or say 'I don't have a task with that exact name'. Speak with absolute authority and connect the dots!\n\nRespond naturally and conversationally. " + extractedSummary + " Keep it concise, helpful, and professional." + evaluationContext,
		},
	}

	for _, msg := range history {
		role := openai.ChatMessageRoleUser
		if msg["role"] == "ai" {
			role = openai.ChatMessageRoleAssistant
		}
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    role,
			Content: msg["content"],
		})
	}

	searchTool := openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "search_tasks",
			Description: "Search the database for tasks based on EXACT keyword, client, or status. If you know the Jira Key, pass it explicitly.",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"keyword": { "type": "string", "description": "Exact keyword to search in title or description" },
					"jira_key": { "type": "string", "description": "Exact Jira ticket ID (e.g. SAAS-540)" },
					"client": { "type": "string", "description": "Client name to filter by" },
					"status": { "type": "string", "description": "Task status to filter by (e.g. IN_PROGRESS, DONE)" }
				}
			}`),
		},
	}

	semanticSearchTool := openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        "semantic_search_tasks",
			Description: "Search the database using AI Vector Memory to find tasks by their semantic MEANING and INTENT, even if exact keywords don't match (e.g. 'money' -> 'invoice').",
			Parameters: json.RawMessage(`{
				"type": "object",
				"properties": {
					"query": { "type": "string", "description": "Natural language query to search for conceptually" }
				},
				"required": ["query"]
			}`),
		},
	}

	req := openai.ChatCompletionRequest{
		Messages:    messages,
		Temperature: 0.1,
		MaxTokens:   500,
		Tools:       []openai.Tool{searchTool, semanticSearchTool},
	}

	resp, err := executeWithFallback(ctx, req, providers)
	if err != nil {
		return "", err
	}

	// Step 2: Handle Tool Call
	if len(resp.Choices) > 0 && len(resp.Choices[0].Message.ToolCalls) > 0 {
		toolCall := resp.Choices[0].Message.ToolCalls[0]
		if toolCall.Function.Name == "search_tasks" {
			var args struct {
				Keyword string `json:"keyword"`
				JiraKey string `json:"jira_key"`
				Client  string `json:"client"`
				Status  string `json:"status"`
			}
			json.Unmarshal([]byte(toolCall.Function.Arguments), &args)

			// Perform DB search using SQLite
			query := db.DB.Model(&models.Task{})

			if args.JiraKey != "" {
				query = query.Where("jira_key = ?", args.JiraKey)
			} else {
				if args.Keyword != "" {
					query = query.Where("title LIKE ? OR description LIKE ? OR jira_key LIKE ?", "%"+args.Keyword+"%", "%"+args.Keyword+"%", "%"+args.Keyword+"%")
				}
				if args.Client != "" {
					query = query.Where("client LIKE ?", "%"+args.Client+"%")
				}
				if args.Status != "" {
					query = query.Where("status = ?", args.Status)
				}
			}

			var results []models.Task
			query.Limit(20).Find(&results)

			// Phase B: AI Fallback Tool logic
			if len(results) == 0 && args.JiraKey != "" {
				log.Printf("[AI] Task %s not found in SQLite. Fetching live from Jira...", args.JiraKey)
				liveTask, err := fetchLiveJiraTicket(args.JiraKey)
				if err == nil && liveTask != nil {
					// Save it to DB so we have it next time
					liveTask.ID = uuid.New().String()
					db.DB.Create(liveTask)
					results = append(results, *liveTask)
					log.Printf("[AI] Successfully fetched and cached live ticket %s", args.JiraKey)
				} else {
					log.Printf("[AI] Failed to fetch live ticket %s: %v", args.JiraKey, err)
				}
			}

			searchResults := "Search Results:\n"
			if len(results) == 0 {
				searchResults += "No tasks found matching criteria. If the user provided a Jira Key, tell them it definitely doesn't exist in Jira.\n"
			}
			for _, t := range results {
				searchResults += fmt.Sprintf("- [%s] %s (Status: %s, Client: %s)\n  Desc: %s\n  Assignee: %s | Source: %s | Team: %s | Sprint: %s | Type: %s | Parent: %s\n",
					t.JiraKey, t.Title, t.Status, t.Client, t.Description,
					t.Assignee, t.SourceName, t.Team, t.Sprint, t.TaskType, t.ParentKey)
				if t.DueDate != nil {
					searchResults += fmt.Sprintf("  Due: %s\n", t.DueDate.Format("2006-01-02"))
				}
			}

			// Append tool response
			messages = append(messages, resp.Choices[0].Message)
			messages = append(messages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    searchResults,
				Name:       toolCall.Function.Name,
				ToolCallID: toolCall.ID,
			})

			// Make second call for final answer
			req.Messages = messages
			req.Tools = nil // No more tools to prevent infinite loop
			resp2, err := executeWithFallback(ctx, req, providers)
			if err != nil {
				return "", err
			}
			return resp2.Choices[0].Message.Content, nil
		} else if toolCall.Function.Name == "semantic_search_tasks" {
			var args struct {
				Query string `json:"query"`
			}
			json.Unmarshal([]byte(toolCall.Function.Arguments), &args)

			log.Printf("[AI] Semantic Search triggered for: %s", args.Query)
			results, err := memory.SearchTasksSemantic(args.Query, 10)

			searchResults := "Semantic Search Results:\n"
			if err != nil {
				searchResults += fmt.Sprintf("Error performing semantic search: %v\n", err)
			} else if len(results) == 0 {
				searchResults += "No tasks found matching the semantic meaning.\n"
			} else {
				for _, t := range results {
					searchResults += fmt.Sprintf("- [%s] %s (Status: %s)\n  Assignee: %s\n",
						t.JiraKey, t.Title, t.Status, t.Assignee)
				}
			}

			// Append tool response
			messages = append(messages, resp.Choices[0].Message)
			messages = append(messages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    searchResults,
				Name:       toolCall.Function.Name,
				ToolCallID: toolCall.ID,
			})

			// Make second call for final answer
			req.Messages = messages
			req.Tools = nil
			resp2, err := executeWithFallback(ctx, req, providers)
			if err != nil {
				return "", err
			}
			return resp2.Choices[0].Message.Content, nil
		}
	}

	return resp.Choices[0].Message.Content, nil
}

// 3. Complex Reasoning (The Analyzer)
// Used for deep analysis, report generation, and priority conflict resolution.
func (r *Router) Reason(ctx context.Context, prompt string) (string, error) {
	providers := []Provider{
		{BaseURL: "https://api.edenai.run/v3", Key: r.edenKey, Model: "openai/gpt-4o", Name: "EdenAI-Reasoning"},
		{BaseURL: "https://openrouter.ai/api/v1", Key: r.openrouterKey, Model: "deepseek/deepseek-chat", Name: "OpenRouter-DeepSeek"},
		{BaseURL: "https://api.deepseek.com/v1", Key: r.deepseekKey, Model: "deepseek-chat", Name: "DeepSeek-Direct"},
		{BaseURL: "https://generativelanguage.googleapis.com/v1beta/openai/", Key: r.geminiKey, Model: "gemini-1.5-pro", Name: "Gemini-Pro"},
	}

	if len(providers) == 0 {
		return "", fmt.Errorf("no reasoning engine available")
	}

	req := openai.ChatCompletionRequest{
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are the advanced reasoning engine for an Operational OS. Analyze the following request carefully.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0.2, // low temp for analytical reasoning
	}

	resp, err := executeWithFallback(ctx, req, providers)

	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}

// fetchLiveJiraTicket fetches a single ticket directly by its ID
func fetchLiveJiraTicket(jiraKey string) (*models.Task, error) {
	domain := os.Getenv("JIRA_DOMAIN")
	email := os.Getenv("JIRA_EMAIL")
	token := os.Getenv("JIRA_API_TOKEN")

	if domain == "" || email == "" || token == "" {
		return nil, fmt.Errorf("JIRA credentials missing in .env")
	}

	apiURL := fmt.Sprintf("https://%s/rest/api/3/issue/%s?expand=names&fields=*all", domain, jiraKey)
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", email, token)))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch ticket %s, status: %d", jiraKey, resp.StatusCode)
	}

	var issue struct {
		Key    string                 `json:"key"`
		Fields map[string]interface{} `json:"fields"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, err
	}

	title, _ := issue.Fields["summary"].(string)
	description, _ := issue.Fields["description"].(string)

	statusVal := "DRAFT"
	if statusMap, ok := issue.Fields["status"].(map[string]interface{}); ok {
		if statusName, ok := statusMap["name"].(string); ok {
			if statusName == "Done" || statusName == "Closed" || statusName == "Resolved" {
				statusVal = "DONE"
			} else if statusName == "In Progress" || statusName == "In Review" {
				statusVal = "IN_PROGRESS"
			}
		}
	}

	assigneeName := "Unassigned"
	if assigneeMap, ok := issue.Fields["assignee"].(map[string]interface{}); ok && assigneeMap != nil {
		if dn, ok := assigneeMap["displayName"].(string); ok {
			assigneeName = dn
		}
	}

	task := models.Task{
		Title:       title,
		Description: description,
		Status:      models.TaskStatus(statusVal),
		Assignee:    assigneeName,
		JiraKey:     issue.Key,
		SourceName:  "Jira",
	}

	return &task, nil
}
