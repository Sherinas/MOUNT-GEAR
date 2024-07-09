package models

import (
	"time"
)

type AdminUser struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"size:100"`
	Email     string    `gorm:"size:100;unique;not null"`
	Password  string    `gorm:"size:100;not null"`
	IsActive  bool      `gorm:"default:true"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type Banner struct {
	ID        uint      `gorm:"primaryKey"`
	Title     string    `gorm:"size:100"`
	ImageURL  string    `gorm:"size:255"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
