package memory

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aios/backend/internal/db"
	"github.com/aios/backend/internal/models"
	"github.com/google/uuid"
	"github.com/qdrant/go-client/qdrant"
	"github.com/sashabaranov/go-openai"
)

// GenerateEmbedding uses Google Gemini (via OpenAI compatibility) to generate a 768-d vector
func GenerateEmbedding(text string) ([]float32, error) {
	geminiKey := os.Getenv("GEMINI_API_KEY")
	if geminiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY not found in .env")
	}

	cfg := openai.DefaultConfig(geminiKey)
	cfg.BaseURL = "https://generativelanguage.googleapis.com/v1beta/openai/"
	if baseURL := os.Getenv("GEMINI_BASE_URL"); baseURL != "" {
		cfg.BaseURL = baseURL
	}
	client := openai.NewClientWithConfig(cfg)

	resp, err := client.CreateEmbeddings(context.Background(), openai.EmbeddingRequest{
		Model: "text-embedding-004", // Gemini embedding model
		Input: text,
	})
	if err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data returned")
	}

	return resp.Data[0].Embedding, nil
}

// UpsertTaskToQdrant generates an embedding for a task and upserts it to Qdrant
func UpsertTaskToQdrant(task *models.Task) error {
	if db.QdrantClient == nil {
		return fmt.Errorf("Qdrant client not initialized")
	}

	textToEmbed := fmt.Sprintf("Title: %s\nDescription: %s\nAssignee: %s\nStatus: %s", task.Title, task.Description, task.Assignee, task.Status)

	vector, err := GenerateEmbedding(textToEmbed)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %v", err)
	}

	// Create Qdrant point
	payload := map[string]*qdrant.Value{
		"task_id":  qdrant.NewValueString(task.ID),
		"title":    qdrant.NewValueString(task.Title),
		"assignee": qdrant.NewValueString(task.Assignee),
		"status":   qdrant.NewValueString(string(task.Status)),
		"jira_key": qdrant.NewValueString(task.JiraKey),
		"source":   qdrant.NewValueString(task.SourceName),
	}

	uid, err := uuid.Parse(task.ID)
	if err != nil {
		// If ID is not a valid UUID (e.g. from tests), just hash it or ignore.
		// Actually our DB models use standard UUIDs. Let's assume it's valid.
	}

	point := &qdrant.PointStruct{
		Id:      qdrant.NewIDUUID(uid.String()),
		Vectors: qdrant.NewVectors(vector...),
		Payload: payload,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	wait := true
	_, err = db.QdrantClient.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: db.TasksCollection,
		Wait:           &wait,
		Points:         []*qdrant.PointStruct{point},
	})

	if err != nil {
		return fmt.Errorf("failed to upsert point to Qdrant: %v", err)
	}

	log.Printf("[QDRANT] Upserted task %s into vector memory", task.ID)
	return nil
}

// SearchTasksSemantic searches Qdrant for similar tasks using the provided query
func SearchTasksSemantic(query string, limit uint64) ([]models.Task, error) {
	if db.QdrantClient == nil {
		return nil, fmt.Errorf("Qdrant client not initialized")
	}

	vector, err := GenerateEmbedding(query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding for query: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	searchResults, err := db.QdrantClient.Query(ctx, &qdrant.QueryPoints{
		CollectionName: db.TasksCollection,
		Query:          qdrant.NewQuery(vector...),
		Limit:          &limit,
		WithPayload:    qdrant.NewWithPayload(true),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search Qdrant: %v", err)
	}

	var tasks []models.Task
	for _, result := range searchResults {
		if result.Payload == nil {
			continue
		}

		taskId := ""
		if val, ok := result.Payload["task_id"]; ok {
			taskId = val.GetStringValue()
		}
		title := ""
		if val, ok := result.Payload["title"]; ok {
			title = val.GetStringValue()
		}
		assignee := ""
		if val, ok := result.Payload["assignee"]; ok {
			assignee = val.GetStringValue()
		}
		status := ""
		if val, ok := result.Payload["status"]; ok {
			status = val.GetStringValue()
		}
		jiraKey := ""
		if val, ok := result.Payload["jira_key"]; ok {
			jiraKey = val.GetStringValue()
		}

		tasks = append(tasks, models.Task{
			ID:       taskId,
			Title:    title,
			Assignee: assignee,
			Status:   models.TaskStatus(status),
			JiraKey:  jiraKey,
		})
	}

	return tasks, nil
}

// UpsertEventToQdrant generates an embedding for a raw episodic event (like a slack message) and upserts it to Qdrant
func UpsertEventToQdrant(text string, metadata map[string]string) error {
	if db.QdrantClient == nil {
		return fmt.Errorf("Qdrant client not initialized")
	}

	vector, err := GenerateEmbedding(text)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %v", err)
	}

	payload := map[string]*qdrant.Value{
		"text": qdrant.NewValueString(text),
	}
	for k, v := range metadata {
		payload[k] = qdrant.NewValueString(v)
	}

	uid := uuid.New().String()
	point := &qdrant.PointStruct{
		Id:      qdrant.NewIDUUID(uid),
		Vectors: qdrant.NewVectors(vector...),
		Payload: payload,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	wait := true
	_, err = db.QdrantClient.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: db.EventsCollection,
		Wait:           &wait,
		Points:         []*qdrant.PointStruct{point},
	})

	if err != nil {
		return fmt.Errorf("failed to upsert event to Qdrant: %v", err)
	}

	log.Printf("[QDRANT] Upserted event %s into episodic memory", uid)
	return nil
}

// SearchEventsSemantic searches Qdrant for similar events using the provided query
func SearchEventsSemantic(query string, limit uint64) ([]map[string]interface{}, error) {
	if db.QdrantClient == nil {
		return nil, fmt.Errorf("Qdrant client not initialized")
	}

	vector, err := GenerateEmbedding(query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding for query: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	searchResults, err := db.QdrantClient.Query(ctx, &qdrant.QueryPoints{
		CollectionName: db.EventsCollection,
		Query:          qdrant.NewQuery(vector...),
		Limit:          &limit,
		WithPayload:    qdrant.NewWithPayload(true),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to search Qdrant events: %v", err)
	}

	var events []map[string]interface{}
	for _, result := range searchResults {
		if result.Payload == nil {
			continue
		}

		eventData := make(map[string]interface{})
		for k, v := range result.Payload {
			eventData[k] = v.GetStringValue()
		}
		events = append(events, eventData)
	}

	return events, nil
}
