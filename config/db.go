package config

import (
	"fmt"
	"log"
	"os"

	"kartcis-backend/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	godotenv.Load()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=kartcis port=5432 sslmode=disable"
		fmt.Println("Warning: DATABASE_URL not found, using default:", dsn)
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Println("Failed to connect to database:", err)
		log.Println("CRITICAL: Running without Database connection!")
		return
	}

	fmt.Println("Connected to Database!")

	// Sync Models with DB
	err = DB.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Event{},
		&models.TicketType{},
		&models.Order{},
		&models.Ticket{},
		&models.SocialAccount{},
		&models.OrderStatusHistory{},
		&models.PasswordReset{},
		&models.SiteSetting{}, // Added
		&models.RequestLog{},  // Added for smart logging
	)
	if err != nil {
		log.Println("Migration failed:", err)
	}

	// Manual Migration Fix: Ensure new columns exist if AutoMigrate fails
	if !DB.Migrator().HasColumn(&models.Order{}, "payment_url") {
		DB.Migrator().AddColumn(&models.Order{}, "payment_url")
	}
	if !DB.Migrator().HasColumn(&models.Order{}, "payment_data") {
		DB.Migrator().AddColumn(&models.Order{}, "payment_data")
	}
	if !DB.Migrator().HasColumn(&models.Order{}, "unique_code") {
		DB.Migrator().AddColumn(&models.Order{}, "unique_code")
	}
	if !DB.Migrator().HasColumn(&models.Order{}, "payment_instructions") {
		DB.Migrator().AddColumn(&models.Order{}, "payment_instructions")
	}
	// Ensure history table exists just in case
	if !DB.Migrator().HasTable(&models.OrderStatusHistory{}) {
		DB.Migrator().CreateTable(&models.OrderStatusHistory{})
	}

	seedSettings(DB)

	// Data Migration: Rename 'ended' to 'completed'
	DB.Exec("UPDATE events SET status = 'completed' WHERE status = 'ended'")

	// Data Cleanup: Cap Available at Quota (Fixes data corruption where available > quota)
	DB.Exec("UPDATE ticket_types SET available = quota WHERE available > quota OR available < 0")

	seedSettings(DB)
}

func seedSettings(db *gorm.DB) {
	defaults := []models.SiteSetting{
		{Key: "contact_email", Value: "support@kartcis.id"},
		{Key: "contact_phone", Value: "+628123456789"},
		{Key: "contact_address", Value: "Jl. Kaliurang Km 14, Yogyakarta"},
		{Key: "facebook_url", Value: "https://facebook.com/kartcis"},
		{Key: "twitter_url", Value: "https://twitter.com/kartcis"},
		{Key: "instagram_url", Value: "https://instagram.com/kartcis"},
	}

	for _, s := range defaults {
		var existing models.SiteSetting
		if err := db.Where("key = ?", s.Key).First(&existing).Error; err != nil {
			// Not found, create
			db.Create(&s)
			log.Println("Seeded setting:", s.Key)
		}
	}
}
