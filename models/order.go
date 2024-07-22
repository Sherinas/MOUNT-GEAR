package models

import "time"

type Order struct {
	ID            uint    `gorm:"primaryKey"`
	UserID        uint    `gorm:"not null"`
	AddressID     uint    `gorm:"not null"`
	TotalAmount   float64 `gorm:"type:decimal(10,2)"`
	Discount      float64 `gorm:"type:decimal(10,2)"`
	FinalAmount   float64 `gorm:"type:decimal(10,2)"`
	PaymentMethod string  `gorm:"size:50"`
	Status        string  `gorm:"size:50;check:status IN ('Pending', 'Confirmed', 'Shipped', 'Delivered', 'Canceled', 'Partially Canceled')"`

	CancellationReason string `gorm:"type:text"`
	ReturnReason       string `gorm:"type:text"`
	AdminReturnNote    string `gorm:"type:text"`

	Items     []OrderItem `gorm:"foreignKey:OrderID"`
	CreatedAt time.Time   `gorm:"autoCreateTime"`
	UpdatedAt time.Time   `gorm:"autoUpdateTime"`
}

type OrderItem struct {
	ID               uint      `gorm:"primaryKey"`
	OrderID          uint      `gorm:"not null"`
	ProductID        uint      `gorm:"not null"`
	Quantity         int       `gorm:"not null"`
	Price            float64   `gorm:"type:decimal(10,2)"`
	CreatedAt        time.Time `gorm:"autoCreateTime"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime"`
	IsCanceled       bool      `gorm:"default:false"`
	CanceledQuantity int       `gorm:"default:0"`
}

type Payment struct {
	ID            uint      `gorm:"primaryKey"`
	OrderID       uint      `gorm:"not null"`
	PaymentMethod string    `gorm:"size:50"`
	Amount        float64   `gorm:"type:decimal(10,2)"`
	Status        string    `gorm:"size:50"`
	TransactionID string    `gorm:"size:100"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}

type Wishlist struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null"`
	User      User      `gorm:"foreignKey:UserID"`
	ProductID uint      `gorm:"not null"`
	Product   Product   `gorm:"foreignKey:ProductID"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
