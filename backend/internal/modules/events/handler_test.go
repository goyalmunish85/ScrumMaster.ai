package events

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aios/backend/internal/db"
	"github.com/joho/godotenv"
)

func TestGetEventsHandler(t *testing.T) {
	// Need to initialize db to prevent panic
	_ = godotenv.Load("../../../.env")
	db.InitDB()

	req, err := http.NewRequest("GET", "/api/v1/events", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(GetEventsHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}
