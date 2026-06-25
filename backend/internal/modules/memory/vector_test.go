package memory

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/sashabaranov/go-openai"
)

func TestGenerateEmbedding(t *testing.T) {
	// Save original env vars to restore later
	origKey := os.Getenv("GEMINI_API_KEY")
	origBaseURL := os.Getenv("GEMINI_BASE_URL")
	defer func() {
		os.Setenv("GEMINI_API_KEY", origKey)
		os.Setenv("GEMINI_BASE_URL", origBaseURL)
	}()

	t.Run("missing GEMINI_API_KEY", func(t *testing.T) {
		os.Setenv("GEMINI_API_KEY", "")
		_, err := GenerateEmbedding("test")
		if err == nil {
			t.Fatal("expected error for missing GEMINI_API_KEY, got nil")
		}
		if err.Error() != "GEMINI_API_KEY not found in .env" {
			t.Fatalf("unexpected error message: %v", err)
		}
	})

	t.Run("successful embedding generation", func(t *testing.T) {
		os.Setenv("GEMINI_API_KEY", "test-key")

		// Create a mock server
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/embeddings" {
				t.Errorf("expected path /embeddings, got %s", r.URL.Path)
			}

			resp := openai.EmbeddingResponse{
				Data: []openai.Embedding{
					{
						Embedding: []float32{0.1, 0.2, 0.3},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer ts.Close()

		os.Setenv("GEMINI_BASE_URL", ts.URL+"/")

		embedding, err := GenerateEmbedding("test text")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(embedding) != 3 {
			t.Fatalf("expected embedding length 3, got %d", len(embedding))
		}
		if embedding[0] != 0.1 || embedding[1] != 0.2 || embedding[2] != 0.3 {
			t.Fatalf("unexpected embedding data: %v", embedding)
		}
	})

	t.Run("no embedding data returned", func(t *testing.T) {
		os.Setenv("GEMINI_API_KEY", "test-key")

		// Create a mock server
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := openai.EmbeddingResponse{
				Data: []openai.Embedding{},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer ts.Close()

		os.Setenv("GEMINI_BASE_URL", ts.URL+"/")

		_, err := GenerateEmbedding("test text")
		if err == nil {
			t.Fatal("expected error for no embedding data, got nil")
		}
		if err.Error() != "no embedding data returned" {
			t.Fatalf("unexpected error message: %v", err)
		}
	})

	t.Run("api error", func(t *testing.T) {
		os.Setenv("GEMINI_API_KEY", "test-key")

		// Create a mock server that returns an error
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": {"message": "internal server error"}}`))
		}))
		defer ts.Close()

		os.Setenv("GEMINI_BASE_URL", ts.URL+"/")

		_, err := GenerateEmbedding("test text")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}
