package models

import "time"

type SalesReport struct {
	ID             uint      `gorm:"primaryKey"`
	ReportDate     time.Time `gorm:"type:date"`
	TotalSales     float64   `gorm:"type:decimal(10,2)"`
	TotalOrders    int
	TotalDiscounts float64   `gorm:"type:decimal(10,2)"`
	CreatedAt      time.Time `gorm:"autoCreateTime"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime"`
}
