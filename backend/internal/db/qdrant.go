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

// InitQdrant initializes the connection to Qdrant and creates the tasks collection if it doesn't exist
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

	// Check if collection exists
	ctx := context.Background()
	exists, err := QdrantClient.CollectionExists(ctx, TasksCollection)
	if err != nil {
		log.Printf("[ERROR] Failed to check Qdrant collection: %v", err)
		return
	}

	if !exists {
		log.Printf("[QDRANT] Creating collection '%s'", TasksCollection)
		err = QdrantClient.CreateCollection(ctx, &qdrant.CreateCollection{
			CollectionName: TasksCollection,
			VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
				Size:     VectorSize,
				Distance: qdrant.Distance_Cosine,
			}),
		})
		if err != nil {
			log.Printf("[ERROR] Failed to create Qdrant collection: %v", err)
		} else {
			log.Printf("[QDRANT] Collection '%s' created successfully", TasksCollection)
		}
	} else {
		log.Printf("[QDRANT] Collection '%s' already exists", TasksCollection)
	}

	// Check if events collection exists
	eventsExists, err := QdrantClient.CollectionExists(ctx, EventsCollection)
	if err != nil {
		log.Printf("[ERROR] Failed to check Qdrant collection '%s': %v", EventsCollection, err)
		return
	}

	if !eventsExists {
		log.Printf("[QDRANT] Creating collection '%s'", EventsCollection)
		err = QdrantClient.CreateCollection(ctx, &qdrant.CreateCollection{
			CollectionName: EventsCollection,
			VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
				Size:     VectorSize,
				Distance: qdrant.Distance_Cosine,
			}),
		})
		if err != nil {
			log.Printf("[ERROR] Failed to create Qdrant collection '%s': %v", EventsCollection, err)
		} else {
			log.Printf("[QDRANT] Collection '%s' created successfully", EventsCollection)
		}
	} else {
		log.Printf("[QDRANT] Collection '%s' already exists", EventsCollection)
	}
}
