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

	// Check and create Tasks collection
	tasksExists, err := QdrantClient.CollectionExists(ctx, TasksCollection)
	if err != nil {
		log.Printf("[ERROR] Failed to check Qdrant tasks collection: %v", err)
		return
	}

	if !tasksExists {
		log.Printf("[QDRANT] Creating collection '%s'", TasksCollection)
		err = QdrantClient.CreateCollection(ctx, &qdrant.CreateCollection{
			CollectionName: TasksCollection,
			VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
				Size:     VectorSize,
				Distance: qdrant.Distance_Cosine,
			}),
		})
		if err != nil {
			log.Printf("[ERROR] Failed to create Qdrant tasks collection: %v", err)
		} else {
			log.Printf("[QDRANT] Collection '%s' created successfully", TasksCollection)
		}
	} else {
		log.Printf("[QDRANT] Collection '%s' already exists", TasksCollection)
	}

	// Check and create Events collection
	eventsExists, err := QdrantClient.CollectionExists(ctx, EventsCollection)
	if err != nil {
		log.Printf("[ERROR] Failed to check Qdrant events collection: %v", err)
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
			log.Printf("[ERROR] Failed to create Qdrant events collection: %v", err)
		} else {
			log.Printf("[QDRANT] Collection '%s' created successfully", EventsCollection)
		}
	} else {
		log.Printf("[QDRANT] Collection '%s' already exists", EventsCollection)
	}
}
