package search

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aios/backend/internal/modules/memory"
)

// OptimizedTask payload minimizes JSON size for network speed
type OptimizedTask struct {
	ID       string `json:"id"`
	Title    string `json:"t,omitempty"`
	Assignee string `json:"a,omitempty"`
	Status   string `json:"s,omitempty"`
	JiraKey  string `json:"jk,omitempty"`
}

type CacheEntry struct {
	Data      []OptimizedTask
	ExpiresAt time.Time
}

var (
	searchCache sync.Map
	cacheTTL    = 5 * time.Minute
)

func getFromCache(key string) ([]OptimizedTask, bool) {
	if val, ok := searchCache.Load(key); ok {
		entry := val.(CacheEntry)
		if time.Now().Before(entry.ExpiresAt) {
			return entry.Data, true
		}
		// Expired
		searchCache.Delete(key)
	}
	return nil, false
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")
	query = strings.TrimSpace(query)
	if query == "" {
		http.Error(w, "Missing 'q' parameter", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := uint64(10) // default limit
	if limitStr != "" {
		if parsed, err := strconv.ParseUint(limitStr, 10, 64); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	cacheKey := query + "|" + strconv.FormatUint(limit, 10)
	if cachedData, ok := getFromCache(cacheKey); ok {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "HIT")
		json.NewEncoder(w).Encode(cachedData)
		return
	}

	tasks, err := memory.SearchEventsSemantic(query, limit)
	if err != nil {
		http.Error(w, "Search failed", http.StatusInternalServerError)
		return
	}

	// Optimize the payload mapping fields to shorter JSON keys
	optimizedTasks := make([]OptimizedTask, 0, len(tasks))
	for _, t := range tasks {
		optimizedTasks = append(optimizedTasks, OptimizedTask{
			ID:       t.ID,
			Title:    t.Title,
			Assignee: t.Assignee,
			Status:   string(t.Status),
			JiraKey:  t.JiraKey,
		})
	}

	searchCache.Store(cacheKey, CacheEntry{Data: optimizedTasks, ExpiresAt: time.Now().Add(cacheTTL)})

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "MISS")
	json.NewEncoder(w).Encode(optimizedTasks)
}
