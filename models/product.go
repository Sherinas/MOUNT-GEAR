package models

import "time"

type Product struct {
	ID          uint     `gorm:"Primarykey"`
	CategoryID  uint     `gorm:"not null"`
	Category    Category `gorm:"foreignKey:CategoryID"`
	Name        string   `gorm:"size:100"`
	Description string   `gorm:"type:text"`
	Price       float64  `gorm:"type:decimal(10,2);check:price >= 0"`
	Discount    float64  `gorm:"type:decimal(10,2)"`
	// DiscountPercentage float64 `gorm:"type:decimal(5,2);default:0;check:discount_percentage >= 0 AND discount_percentage <= 99"`
	Stock int32 `gorm:"not null; check:stock >= 0"`

	IsActive  bool      `gorm:"default:true"`
	Images    []Image   `gorm:"foreignKey:ProductID"`
	Reviews   []Review  `gorm:"foreignKey:ProductID"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
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

type Cart struct {
	ID        uint       `gorm:"primaryKey"`
	UserID    uint       `gorm:"not null"`
	User      User       `gorm:"foreignKey:UserID"`
	CartItems []CartItem `gorm:"foreignKey:CartID"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime"`
}

type CartItem struct {
	ID        uint      `gorm:"primaryKey"`
	CartID    uint      `gorm:"not null"`
	ProductID uint      `gorm:"not null"`
	Product   Product   `gorm:"foreignKey:ProductID"`
	Quantity  int       `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// GetDiscountedPrice calculates and returns the discounted price of the product
func (p *Product) GetDiscountedPrice() float64 {
	if p.Discount > 0 {
		discountAmount := p.Price * (p.Discount / 100)
		return p.Price - discountAmount
	}
	return p.Price
}
