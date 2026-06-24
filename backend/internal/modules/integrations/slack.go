package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aios/backend/internal/db"
	"github.com/aios/backend/internal/models"
	"github.com/aios/backend/internal/modules/ai"
	"github.com/aios/backend/internal/modules/events"
	"github.com/aios/backend/internal/modules/memory"
	"github.com/google/uuid"
)

type SlackMessage struct {
	Type     string `json:"type"`
	Subtype  string `json:"subtype"`
	Text     string `json:"text"`
	User     string `json:"user"`
	Ts       string `json:"ts"`
	ThreadTs string `json:"thread_ts"`
}

type SlackChannelInfoResponse struct {
	Ok      bool `json:"ok"`
	Channel struct {
		Name string `json:"name"`
	} `json:"channel"`
}

type SlackHistoryResponse struct {
	Ok               bool           `json:"ok"`
	Messages         []SlackMessage `json:"messages"`
	ResponseMetadata struct {
		NextCursor string `json:"next_cursor"`
	} `json:"response_metadata"`
}

func fetchSlackAPI(url string, token string, target interface{}) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch slack api, status: %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

// SyncSlackChannel fetches recent messages (3 days) and threads, and extracts tasks via AI
func SyncSlackChannel(channelID string) (string, error) {
	token := os.Getenv("SLACK_BOT_TOKEN")
	if token == "" {
		return "", fmt.Errorf("SLACK_BOT_TOKEN is missing")
	}

	// 1. Determine oldest timestamp to fetch
	var syncState models.SlackSyncState
	oldestTs := fmt.Sprintf("%d", time.Now().Add(-72*time.Hour).Unix()) // Default 3 days
	if err := db.DB.Where("channel_id = ?", channelID).First(&syncState).Error; err == nil && syncState.LastTimestamp != "" {
		oldestTs = syncState.LastTimestamp
	}

	var allMessages []SlackMessage
	cursor := ""

	for {
		historyURL := fmt.Sprintf("https://slack.com/api/conversations.history?channel=%s&oldest=%s&limit=200", channelID, oldestTs)
		if cursor != "" {
			historyURL += "&cursor=" + cursor
		}

		var slackResp SlackHistoryResponse
		if err := fetchSlackAPI(historyURL, token, &slackResp); err != nil {
			return "", err
		}

		if !slackResp.Ok {
			return "", fmt.Errorf("slack api returned not ok")
		}

		allMessages = append(allMessages, slackResp.Messages...)

		if slackResp.ResponseMetadata.NextCursor == "" {
			break
		}
		cursor = slackResp.ResponseMetadata.NextCursor
	}

	log.Printf("[SYNC] Fetched %d main messages from Slack channel %s", len(allMessages), channelID)

	// Fetch Channel Name for Client Context
	channelName := channelID
	var infoResp SlackChannelInfoResponse
	infoURL := fmt.Sprintf("https://slack.com/api/conversations.info?channel=%s", channelID)
	if err := fetchSlackAPI(infoURL, token, &infoResp); err == nil && infoResp.Ok {
		channelName = infoResp.Channel.Name
	}

	// 2. Build a unified transcript
	var transcriptBuilder strings.Builder
	transcriptBuilder.WriteString(fmt.Sprintf("--- Slack Conversation Transcript for Channel #%s (Context: Client or Team Name is likely '%s') ---\n\n", channelName, channelName))

	// Helper to insert ActivityLog
	logActivity := func(m SlackMessage) {
		if db.DB == nil {
			return
		}

		sec := int64(0)
		if parts := strings.Split(m.Ts, "."); len(parts) > 0 {
			fmt.Sscanf(parts[0], "%d", &sec)
		}

		ts := time.Now()
		if sec > 0 {
			ts = time.Unix(sec, 0)
		}

		activity := models.ActivityLog{
			ID:        uuid.New().String(),
			Platform:  "slack",
			Author:    m.User,
			Content:   m.Text,
			SourceRef: "slack_" + channelID + "_" + m.Ts,
			Timestamp: ts,
		}

		// Insert if not exists
		var existing models.ActivityLog
		if err := db.DB.Where("source_ref = ?", activity.SourceRef).First(&existing).Error; err != nil {
			db.DB.Create(&activity)
		}
	}

	// Slack returns history newest-first. We iterate in reverse to build chronological transcript.
	validMsgCount := 0
	for i := len(allMessages) - 1; i >= 0; i-- {
		msg := allMessages[i]
		if msg.Subtype != "" || msg.User == "" {
			continue // skip bots and system msgs
		}

		transcriptBuilder.WriteString(fmt.Sprintf("User %s: %s\n", msg.User, msg.Text))
		validMsgCount++

		// [PHASE 7] Log activity
		logActivity(msg)

		// Fetch threads if applicable
		if msg.ThreadTs != "" && msg.ThreadTs == msg.Ts {
			repliesURL := fmt.Sprintf("https://slack.com/api/conversations.replies?channel=%s&ts=%s", channelID, msg.ThreadTs)
			var repliesResp SlackHistoryResponse
			if err := fetchSlackAPI(repliesURL, token, &repliesResp); err == nil && repliesResp.Ok {
				// replies also come ordered. the first reply is the parent message itself.
				for j := 1; j < len(repliesResp.Messages); j++ {
					rMsg := repliesResp.Messages[j]
					if rMsg.Subtype == "" && rMsg.User != "" {
						transcriptBuilder.WriteString(fmt.Sprintf("  -> Reply by User %s: %s\n", rMsg.User, rMsg.Text))
						validMsgCount++

						// [PHASE 7] Log activity
						logActivity(rMsg)
					}
				}
			}
		}
	}

	if validMsgCount == 0 {
		log.Printf("[SYNC] Channel %s has no user messages. Skipping AI extraction to prevent hallucination.", channelID)
		return fmt.Sprintf("Skipped: Channel %s has no user messages in the last 3 days.", channelName), nil
	}

	fullTranscript := transcriptBuilder.String()
	log.Printf("[SYNC] Built Transcript (%d chars). Sending to AI for contextual extraction...", len(fullTranscript))

	// [PHASE 6] Store the full transcript in Episodic Memory
	go func(text string, chName string) {
		err := memory.UpsertEventToQdrant(text, map[string]string{
			"type":    "slack_transcript",
			"channel": chName,
		})
		if err != nil {
			log.Printf("[MEMORY ERROR] Failed to upsert episodic memory for %s: %v", chName, err)
		}
	}(fullTranscript, channelName)

	// 3. Extract tasks using full context
	router := ai.InitRouter()
	history := []map[string]string{
		{"role": "user", "content": fullTranscript},
	}

	extractedEvents, err := router.Extract(context.Background(), history)
	if err != nil {
		return "", fmt.Errorf("failed to extract from Slack transcript: %v", err)
	}

	for _, ev := range extractedEvents {
		if ev.Type == events.BulkTasks {
			var payload map[string]interface{}
			if err := json.Unmarshal(ev.Payload, &payload); err == nil {
				if tasks, ok := payload["tasks"].([]interface{}); ok {
					for _, tInterface := range tasks {
						if taskMap, ok := tInterface.(map[string]interface{}); ok {
							taskMap["source_name"] = "Slack: " + channelName
						}
					}
				}
				ev.Payload, _ = json.Marshal(payload)
			}
		}
		events.Publish(ev)
	}

	// 5. Update the LastTimestamp in the database
	if len(allMessages) > 0 {
		highestTs := oldestTs
		for _, msg := range allMessages {
			if msg.Ts > highestTs {
				highestTs = msg.Ts
			}
		}

		syncState.ChannelID = channelID
		syncState.LastTimestamp = highestTs
		db.DB.Save(&syncState)
	}

	return fmt.Sprintf("Extracted %d tasks/events from %d Slack messages in channel %s.", len(extractedEvents), validMsgCount, channelName), nil
}

// PostSlackMessage posts a message to a specific Slack channel
func PostSlackMessage(channelID string, message string) error {
	if os.Getenv("DRY_RUN") == "true" {
		log.Printf("[DRY-RUN] Would post to Slack channel %s: %s", channelID, message)
		return nil
	}

	token := os.Getenv("SLACK_BOT_TOKEN")
	if token == "" {
		return fmt.Errorf("SLACK_BOT_TOKEN is missing")
	}

	apiURL := "https://slack.com/api/chat.postMessage"

	payload := map[string]interface{}{
		"channel": channelID,
		"text":    message,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to post slack message, status: %d", resp.StatusCode)
	}

	var result struct {
		Ok    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if !result.Ok {
		return fmt.Errorf("slack api error: %s", result.Error)
	}

	log.Printf("[SLACK] Successfully posted message to channel %s", channelID)
	return nil
}
