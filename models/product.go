package models

import "time"

type Product struct {
	ID            uint      `gorm:"Primarykey"`
	CategoryID    uint      `gorm:"not null"`
	Category      Category  `gorm:"foreignKey:CategoryID"`
	Name          string    `gorm:"size:100"`
	Description   string    `gorm:"type:text"`
	Price         float64   `gorm:"type:decimal(10,2)" validate:"gte=0"`
	DiscountPrice float64   `gorm:"type:decimal(10,2)"`
	Stock         int32     `gorm:"not null" validate:"gte=0"`
	IsActive      bool      `gorm:"default:true"`
	Images        []Image   `gorm:"foreignKey:ProductID"`
	Reviews       []Review  `gorm:"foreignKey:ProductID"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}

type Image struct {
	ID        uint      `gorm:"primaryKey"`
	ProductID uint      `gorm:"not null"`
	FilePath  string    `gorm:"type:text"`
	ImageURL  string    `gorm:"size:255"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type Review struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null"`
	ProductID uint      `gorm:"not null"`
	Rating    int       `gorm:"check:rating >= 1 AND rating <= 5"`
	Comment   string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type Category struct {
	ID          uint      `gorm:"primaryKey"`
	Name        string    `gorm:"size:100"`
	Description string    `gorm:"type:text"`
	IsActive    bool      `gorm:"default:true"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}
