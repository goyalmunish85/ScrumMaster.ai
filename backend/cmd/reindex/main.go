package main

import (
	"context"
	"log"
	"time"

	"github.com/aios/backend/internal/db"
	"github.com/aios/backend/internal/models"
	"github.com/aios/backend/internal/modules/memory"
	"github.com/joho/godotenv"
	"github.com/qdrant/go-client/qdrant"
	"gorm.io/gorm"
)

func main() {
	log.Println("[REINDEX] Starting vector re-indexing script...")

	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("[WARNING] Error loading .env file, continuing with existing environment")
	}

	// Initialize DB and Qdrant
	db.InitDB()
	db.InitQdrant()

	if db.QdrantClient == nil {
		log.Fatalf("[ERROR] Qdrant client is not initialized")
	}

	ctx := context.Background()

	// 1. Delete the existing collection
	log.Printf("[REINDEX] Deleting collection '%s'...", db.TasksCollection)
	err = db.QdrantClient.DeleteCollection(ctx, db.TasksCollection)
	if err != nil {
		log.Printf("[WARNING] Failed to delete collection (might not exist): %v", err)
	}

	// 2. Re-create the collection
	log.Printf("[REINDEX] Re-creating collection '%s'...", db.TasksCollection)
	err = db.QdrantClient.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: db.TasksCollection,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     db.VectorSize,
			Distance: qdrant.Distance_Cosine,
		}),
	})
	if err != nil {
		log.Fatalf("[ERROR] Failed to re-create Qdrant collection: %v", err)
	}
	log.Printf("[REINDEX] Collection '%s' re-created successfully.", db.TasksCollection)

	// 3. Query all tasks from SQLite and upsert
	var totalProcessed int
	var totalFailed int

	var tasks []models.Task
	result := db.DB.FindInBatches(&tasks, 50, func(tx *gorm.DB, batch int) error {
		for _, task := range tasks {
			err := memory.UpsertTaskToQdrant(&task)
			if err != nil {
				log.Printf("[ERROR] Failed to upsert task %s: %v", task.ID, err)
				totalFailed++
			} else {
				totalProcessed++
			}

			// Add a precise delay to avoid rate limits with Gemini API
			time.Sleep(500 * time.Millisecond)
		}
		log.Printf("[REINDEX] Processed batch %d...", batch)
		return nil
	})

	if result.Error != nil {
		log.Fatalf("[ERROR] Failed to fetch tasks from database: %v", result.Error)
	}

	log.Printf("[REINDEX] Re-indexing complete. Successfully processed: %d, Failed: %d", totalProcessed, totalFailed)
}
