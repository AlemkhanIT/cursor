package models

import (
	"time"

	"gorm.io/gorm"
)

type Cart struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uint           `json:"user_id" gorm:"unique;not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	User      User       `json:"user,omitempty" gorm:"foreignKey:UserID"`
	CartItems []CartItem `json:"cart_items,omitempty" gorm:"foreignKey:CartID"`
}

type CartItem struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CartID    uint           `json:"cart_id" gorm:"not null"`
	ProductID uint           `json:"product_id" gorm:"not null"`
	Quantity  int            `json:"quantity" gorm:"not null;default:1"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Cart    Cart    `json:"cart,omitempty" gorm:"foreignKey:CartID"`
	Product Product `json:"product,omitempty" gorm:"foreignKey:ProductID"`
}
