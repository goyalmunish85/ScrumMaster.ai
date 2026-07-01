package search

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// We can add simple test to make sure it exists, but the test cannot hit Qdrant unless mocked.
func TestSearchEventsHandler_MissingQuery(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/search", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(SearchEventsHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	expected := "Missing 'q' query parameter\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}
