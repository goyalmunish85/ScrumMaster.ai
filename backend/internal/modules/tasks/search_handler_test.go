package tasks

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchTasksHandler(t *testing.T) {
	// Test missing query
	req, _ := http.NewRequest("GET", "/api/search", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(SearchTasksHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}

	// Test valid query (will hit fallback since db/qdrant isn't fully mocked here, but shouldn't panic)
	req2, _ := http.NewRequest("GET", "/api/search?query=test", nil)
	rr2 := httptest.NewRecorder()

	handler.ServeHTTP(rr2, req2)

	if status := rr2.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code for valid query: got %v want %v",
			status, http.StatusOK)
	}

	var res []interface{}
	err := json.NewDecoder(rr2.Body).Decode(&res)
	if err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
}
