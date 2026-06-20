package reports

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aios/backend/internal/db"
	"github.com/aios/backend/internal/models"
	"github.com/aios/backend/internal/modules/ai"
)

// GenerateWeeklyReportHandler fetches events from the last 7 days and uses AI to generate an executive summary.
func GenerateWeeklyReportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	var events []models.OperationalEventRecord

	err := db.DB.Where("created_at >= ?", sevenDaysAgo).Order("created_at asc").Find(&events).Error
	if err != nil {
		http.Error(w, "Failed to fetch event history", http.StatusInternalServerError)
		return
	}

	if len(events) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"report": "No operational events found in the last 7 days to summarize.",
		})
		return
	}

	// Prepare the prompt for the Reasoning Engine
	eventLog := "OPERATIONAL EVENT LOG (LAST 7 DAYS):\n"
	for _, e := range events {
		eventLog += fmt.Sprintf("[%s] %s: %s\n", e.CreatedAt.Format("Jan 02 15:04"), e.EventType, e.Payload)
	}

	prompt := fmt.Sprintf(`You are an elite Chief of Staff and Operational Excellence expert. 
Based on the raw log of events below, synthesize a professional "Weekly Executive Status Update".

STRUCTURE:
1. **Executive Summary**: 1-2 sentences on overall velocity.
2. **Key Achievements**: Bullet points of what was actually completed or significantly progressed.
3. **Critical Blockers & Risks**: What is currently stalled and why.
4. **Next Week Priorities**: Suggested focus areas based on recent activity.

Use clean, professional markdown formatting with bold headers and bullet points.

%s`, eventLog)

	// Initialize AI Router
	router := ai.InitRouter()

	// Call the Reasoning Engine (DeepSeek)
	report, err := router.Reason(context.Background(), prompt)
	if err != nil {
		http.Error(w, "AI Reasoning Engine failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"report": report,
	})
}
