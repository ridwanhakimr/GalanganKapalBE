package config

import (
	"log"

	"github.com/shipyard-system/backend/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDatabase(dbURL string) {
	if dbURL == "" {
		log.Println("Database URL is empty! Please check your .env file.")
		return
	}

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	DB = db
	log.Println("Successfully connected to the database!")
}

func MigrateDB() {
	if DB == nil {
		log.Fatal("Database not initialized for migration")
	}

	log.Println("Running AutoMigration...")

	err := DB.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Warehouse{},
		&models.Item{},
		&models.ItemRequest{},
		&models.StockMovement{},
		&models.AuditLog{},
		&models.Notification{},
	)

	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migration completed successfully!")
}
