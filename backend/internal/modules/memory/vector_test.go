package memory

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/aios/backend/internal/db"
	"github.com/aios/backend/internal/models"
)

func TestGenerateEmbedding(t *testing.T) {
	// Test 1: Missing API key
	os.Unsetenv("GEMINI_API_KEY")
	_, err := GenerateEmbedding("test")
	if err == nil || !strings.Contains(err.Error(), "GEMINI_API_KEY not found") {
		t.Errorf("Expected missing GEMINI_API_KEY error, got: %v", err)
	}

	// Test 2: Successful mock server response
	os.Setenv("GEMINI_API_KEY", "dummy-key")
	defer os.Unsetenv("GEMINI_API_KEY")

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify standard paths that OpenAI client uses
		if r.URL.Path != "/embeddings" {
			t.Errorf("Expected path /embeddings, got %s", r.URL.Path)
		}

		response := map[string]interface{}{
			"object": "list",
			"data": []map[string]interface{}{
				{
					"object":    "embedding",
					"embedding": []float32{0.1, 0.2, 0.3},
					"index":     0,
				},
			},
			"model": "text-embedding-004",
			"usage": map[string]int{"prompt_tokens": 8, "total_tokens": 8},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	os.Setenv("GEMINI_BASE_URL", mockServer.URL+"/")
	defer os.Unsetenv("GEMINI_BASE_URL")

	embedding, err := GenerateEmbedding("test text")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(embedding) != 3 || embedding[0] != 0.1 {
		t.Errorf("Unexpected embedding values: %v", embedding)
	}

	// Test 3: API Error response (empty data)
	errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"object": "list",
			"data":   []map[string]interface{}{},
			"model":  "text-embedding-004",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer errorServer.Close()

	os.Setenv("GEMINI_BASE_URL", errorServer.URL+"/")

	_, err = GenerateEmbedding("test text")
	if err == nil || !strings.Contains(err.Error(), "no embedding data returned") {
		t.Errorf("Expected 'no embedding data returned' error, got: %v", err)
	}
}

func TestUpsertTaskToQdrant_NilClient(t *testing.T) {
	db.QdrantClient = nil
	task := &models.Task{
		ID:          "test-id",
		Title:       "Test",
		Description: "Desc",
		Assignee:    "John",
		Status:      models.StatusDraft,
	}

	err := UpsertTaskToQdrant(task)
	if err == nil || !strings.Contains(err.Error(), "Qdrant client not initialized") {
		t.Errorf("Expected 'Qdrant client not initialized' error, got: %v", err)
	}
}

func TestSearchTasksSemantic_NilClient(t *testing.T) {
	db.QdrantClient = nil

	_, err := SearchTasksSemantic("test query", 10)
	if err == nil || !strings.Contains(err.Error(), "Qdrant client not initialized") {
		t.Errorf("Expected 'Qdrant client not initialized' error, got: %v", err)
	}
}
