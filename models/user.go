package models

import "time"

type User struct {
	ID                  uint      `gorm:"primarykey"`
	Name                string    `gorm:"size:100"`
	Email               string    `gorm:"size:100;unique;not null"`
	Phone               string    `gorm:"size:15"`
	Password            string    `gorm:"size:255;not null"`
	IsActive            bool      `gorm:"default:true"`
	CreatedAt           time.Time `gorm:"autoCreateTime"`
	UpdatedAt           time.Time `gorm:"autoUpdateTime"`
	SocialLoginProvider string    `gorm:"size:50"`
	SocialLoginId       string    `gorm:"size:100"`
	Addresses           []Address `gorm:"foreignKey:UserID"`
	Orders              []Order   `gorm:"foreignKey:UserID"`
	OtpExpiry           time.Time `gorm:"not null"`
}

type Address struct {
	ID           uint      `gorm:"primarykey"`
	UserID       uint      `gorm:"not null"`
	AddressLine1 string    `gorm:"size:225"`
	AddressLine2 string    `gorm:"size:225"`
	City         string    `gorm:"size:25"`
	State        string    `gorm:"size:25"`
	Zipcode      string    `gorm:"size:10"`
	Country      string    `gorm:"size:25"`
	IsDefault    bool      `gorm:"default:false"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

// type OTP struct {
// 	ID        uint      `gorm:"primarykey"`
// 	UserID    uint      `gorm:"not null"`
// 	CreatedAt time.Time `gorm:"autoCreateTime"`
// }
