package models

import "time"

type Coupon struct {
	ID        uint      `gorm:"primaryKey"`
	Code      string    `gorm:"size:50"`
	Discount  float64   `gorm:"type:decimal(5,2)"`
	ValidFrom time.Time `gorm:"not null"`
	ValidTo   time.Time `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type Offer struct {
	ID                 uint      `gorm:"primaryKey"`
	OfferType          string    `gorm:"size:50"`
	ProductID          *uint     `gorm:"not null"`
	CategoryID         *uint     `gorm:"not null"`
	ReferralCode       string    `gorm:"size:50"`
	DiscountPercentage float64   `gorm:"type:decimal(5,2)"`
	ValidFrom          time.Time `gorm:"not null"`
	ValidTo            time.Time `gorm:"not null"`
	CreatedAt          time.Time `gorm:"autoCreateTime"`
	UpdatedAt          time.Time `gorm:"autoUpdateTime"`
}
