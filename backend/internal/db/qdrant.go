package db

import (
	"context"
	"log"

	"github.com/qdrant/go-client/qdrant"
)

var QdrantClient *qdrant.Client

const TasksCollection = "tasks"
const EventsCollection = "events"
const VectorSize = 768 // Gemini text-embedding-004 size

// InitQdrant initializes the connection to Qdrant and creates the collections if they don't exist
func InitQdrant() {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: "localhost",
		Port: 6334,
	})
	if err != nil {
		log.Printf("[ERROR] Failed to connect to Qdrant: %v", err)
		return
	}

	QdrantClient = client
	ctx := context.Background()

	// Initialize TasksCollection
	initCollection(ctx, TasksCollection)

	// Initialize EventsCollection
	initCollection(ctx, EventsCollection)
}

func initCollection(ctx context.Context, name string) {
	exists, err := QdrantClient.CollectionExists(ctx, name)
	if err != nil {
		log.Printf("[ERROR] Failed to check Qdrant collection %s: %v", name, err)
		return
	}

	if !exists {
		log.Printf("[QDRANT] Creating collection '%s'", name)
		err = QdrantClient.CreateCollection(ctx, &qdrant.CreateCollection{
			CollectionName: name,
			VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
				Size:     VectorSize,
				Distance: qdrant.Distance_Cosine,
			}),
		})
		if err != nil {
			log.Printf("[ERROR] Failed to create Qdrant collection %s: %v", name, err)
		} else {
			log.Printf("[QDRANT] Collection '%s' created successfully", name)
		}
	} else {
		log.Printf("[QDRANT] Collection '%s' already exists", name)
	}
}
