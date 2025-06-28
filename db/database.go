package db

import (
	"log"
	"recruitment-system/config"
	"recruitment-system/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	var err error
	DB, err = gorm.Open(postgres.Open(config.AppConfig.DBUrl), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connection established.")

	// Auto-migrate the schema
	err = DB.AutoMigrate(&models.User{}, &models.Profile{}, &models.Job{}, &models.JobApplication{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Println("Database migrated.")
}
