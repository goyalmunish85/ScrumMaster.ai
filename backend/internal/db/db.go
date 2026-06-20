package db

import (
	"log"

	"github.com/aios/backend/internal/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() {
	var err error
	// Use pure Go SQLite driver (no CGO required)
	DB, err = gorm.Open(sqlite.Open("aios.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("[DB] Connected to SQLite database successfully.")

	// Auto-migrate the schemas
	err = DB.AutoMigrate(
		&models.Task{},
		&models.TaskDependency{},
		&models.OperationalEventRecord{},
		&models.ChatMessage{},
		&models.AIEvaluation{},
		&models.SyncLog{},
		&models.CronLog{},
		&models.SlackSyncState{},
		&models.IntegrationTarget{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database schemas: %v", err)
	}

	log.Println("[DB] Database schemas migrated successfully.")
}
