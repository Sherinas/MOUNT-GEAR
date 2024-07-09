package models

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func SetDatabase(db *gorm.DB) {
	DB = db
}

func DatabaseSetup() *gorm.DB {
	dsn := "host=localhost port=5432 user=sherinascdlm password=admin123 dbname=mgdata sslmode=disable"
	var err error
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}
	return db
}

func AutoMigrate(db *gorm.DB) error {
	err := db.AutoMigrate(
		&User{},
		&Address{},
		//	&OTP{},
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
		&Stock{},
	)
	return err
}
