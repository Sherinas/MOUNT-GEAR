package models

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func SetDatabase(db *gorm.DB) {
	DB = db
}

func DatabaseSetup() *gorm.DB {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbSSLMode := os.Getenv("DB_SSLMODE")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	return db
}

func AutoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&User{},
		&Address{},
		&Wallet{},
		&Product{},
		&Image{},
		&Review{},
		&Category{},
		&Order{},
		&OrderItem{},
		&Payment{},
		&Coupon{},
		&Offer{},
		&AdminUser{},
		&Banner{},
		&SalesReport{},
		&Wishlist{},
		&CouponUsage{},
		&Cart{}, // Add this line
		&CartItem{},
	)
	return err
}
