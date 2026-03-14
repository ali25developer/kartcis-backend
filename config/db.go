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

	// GORM AutoMigrate
	fmt.Println("Running AutoMigrate...")
	err = DB.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Voucher{},
		&models.Event{},
		&models.TicketType{},
		&models.Order{},
		&models.Ticket{},
		&models.SocialAccount{},
		&models.OrderStatusHistory{},
		&models.PasswordReset{},
		&models.SiteSetting{},
		&models.RequestLog{},
		&models.FlashSale{},
		&models.BankTransaction{},
		&models.EmailVerification{},
		&models.ReferralCode{},
	)
	if err != nil {
		log.Println("AutoMigrate failed:", err)
	}

	// Manual Migration Fix: Pastikan kolom baru ada kalau AutoMigrate gagal (Existing pattern)
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
	if !DB.Migrator().HasColumn(&models.Order{}, "discount_amount") {
		DB.Migrator().AddColumn(&models.Order{}, "discount_amount")
	}
	if !DB.Migrator().HasColumn(&models.Order{}, "voucher_code") {
		DB.Migrator().AddColumn(&models.Order{}, "voucher_code")
	}
	if !DB.Migrator().HasColumn(&models.Order{}, "referral_code") {
		DB.Migrator().AddColumn(&models.Order{}, "referral_code")
	}

	// Multi-Role & Custom Fee
	if !DB.Migrator().HasColumn(&models.Event{}, "organizer_id") {
		DB.Migrator().AddColumn(&models.Event{}, "organizer_id")
	}
	if !DB.Migrator().HasColumn(&models.User{}, "role") {
		DB.Migrator().AddColumn(&models.User{}, "role")
	}
	if !DB.Migrator().HasColumn(&models.User{}, "custom_fee") {
		DB.Migrator().AddColumn(&models.User{}, "custom_fee")
	}

	// Ensure Tables (Existing pattern)
	if !DB.Migrator().HasTable(&models.OrderStatusHistory{}) {
		DB.Migrator().CreateTable(&models.OrderStatusHistory{})
	}
	if !DB.Migrator().HasTable(&models.RequestLog{}) {
		DB.Migrator().CreateTable(&models.RequestLog{})
	}
	if !DB.Migrator().HasTable(&models.Voucher{}) {
		DB.Migrator().CreateTable(&models.Voucher{})
	}
	if !DB.Migrator().HasTable(&models.ReferralCode{}) {
		DB.Migrator().CreateTable(&models.ReferralCode{})
	}
	// Force add columns if missing
	if !DB.Migrator().HasColumn(&models.ReferralCode{}, "partner_name") {
		DB.Migrator().AddColumn(&models.ReferralCode{}, "partner_name")
	}
	if !DB.Migrator().HasColumn(&models.ReferralCode{}, "reward_type") {
		DB.Migrator().AddColumn(&models.ReferralCode{}, "reward_type")
	}
	if !DB.Migrator().HasColumn(&models.ReferralCode{}, "reward_value") {
		DB.Migrator().AddColumn(&models.ReferralCode{}, "reward_value")
	}

	seedSettings(DB)

	// Data Migrations
	DB.Exec("UPDATE events SET status = 'completed' WHERE status = 'ended'")
	DB.Exec("UPDATE ticket_types SET available = quota WHERE available > quota OR available < 0")
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
			db.Create(&s)
			log.Println("Seeded setting:", s.Key)
		}
	}
}
