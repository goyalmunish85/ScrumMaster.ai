package chat

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/aios/backend/internal/db"
	"github.com/aios/backend/internal/models"
	"github.com/aios/backend/internal/modules/ai"
	"github.com/aios/backend/internal/modules/events"
	"github.com/google/uuid"
)

// GetDailyBriefingHandler returns the most recent Daily Briefing message
func GetDailyBriefingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var message models.ChatMessage
	err := db.DB.Where("sender_id = ? AND content LIKE ?", "ai-system", "### 🌅 Good morning! Here is your Daily Briefing%").
		Order("created_at desc").
		First(&message).Error

	if err != nil {
		http.Error(w, "Daily briefing not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(message)
}

// GetMessagesHandler returns all messages for a conversation
func GetMessagesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var messages []models.ChatMessage
	if err := db.DB.Order("created_at asc").Find(&messages).Error; err != nil {
		http.Error(w, "Failed to load messages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

// SendMessageHandler handles incoming user messages and their AI extracted events
func SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Content string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Save User Message
	userMsg := models.ChatMessage{
		ID:       uuid.New().String(),
		Content:  req.Content,
		SenderID: "user-1",
		Role:     "user",
	}
	db.DB.Create(&userMsg)

	// Prepare history (limit to last 10 messages to prevent token bloat)
	var allMessages []models.ChatMessage
	db.DB.Order("created_at desc").Limit(10).Find(&allMessages)

	// Reverse the slice to make it chronological again
	for i, j := 0, len(allMessages)-1; i < j; i, j = i+1, j-1 {
		allMessages[i], allMessages[j] = allMessages[j], allMessages[i]
	}

	var history []map[string]string
	for _, msg := range allMessages {
		history = append(history, map[string]string{
			"role":    msg.Role,
			"content": msg.Content,
		})
	}

	// Initialize Router
	aiRouter := ai.InitRouter()

	// 1. Extract Events ONLY from the newest user message to save tokens
	latestMessageOnly := []map[string]string{
		{"role": "user", "content": req.Content},
	}
	extractedEvents, err := aiRouter.Extract(context.Background(), latestMessageOnly)
	if err != nil {
		log.Printf("[ERROR] AI Extraction failed: %v", err)
	}

	// 2. Publish extracted events to the Event Bus
	for _, evt := range extractedEvents {
		events.Publish(evt)
	}

	// 3. Generate Conversational Response
	aiResponseText, err := aiRouter.Converse(context.Background(), history, extractedEvents)
	if err != nil {
		log.Printf("[ERROR] AI Conversation failed: %v", err)
		aiResponseText = "Sorry, I had trouble generating a response. Details: " + err.Error()
	}

	// Save AI Response Message
	aiMsg := models.ChatMessage{
		ID:       uuid.New().String(),
		Content:  aiResponseText,
		SenderID: "ai-system",
		Role:     "ai",
	}
	db.DB.Create(&aiMsg)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(aiMsg)
}

// EvaluateMessageHandler handles feedback for self-improvement
func EvaluateMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		MessageID string `json:"message_id"`
		Feedback  string `json:"feedback"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Feedback == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	eval := models.AIEvaluation{
		ID:           uuid.New().String(),
		MessageID:    req.MessageID,
		FeedbackText: req.Feedback,
	}

	if err := db.DB.Create(&eval).Error; err != nil {
		http.Error(w, "Failed to save evaluation", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}
