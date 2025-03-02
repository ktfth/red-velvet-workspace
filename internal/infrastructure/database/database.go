package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/red-velvet-workspace/banco-digital/internal/domain/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(db *gorm.DB) error {
	DB = db

	// Auto Migrate the schema
	err := DB.AutoMigrate(
		&models.Account{},
		&models.Transaction{},
		&models.Notification{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %v", err)
	}

	log.Println("Successfully initialized database schema")
	return nil
}

func InitDBConnection() (*gorm.DB, error) {
	dsn := os.Getenv("DATABASE_URL")

	// Default connection string for local development
	defaultDSN := "host=postgres user=admin password=admin123 dbname=banco_digital port=5432 sslmode=disable"

	if dsn == "" {
		dsn = defaultDSN
	}

	var db *gorm.DB
	var err error
	maxRetries := 5

	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("Failed to connect to database (attempt %d/%d): %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			time.Sleep(5 * time.Second)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after %d attempts: %v", maxRetries, err)
	}

	return db, nil
}
