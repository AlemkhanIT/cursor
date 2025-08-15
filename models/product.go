package models

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	Price       float64        `json:"price" gorm:"not null"`
	Stock       int            `json:"stock" gorm:"not null;default:0"`
	ImageURL    string         `json:"image_url"`
	Category    string         `json:"category"`
	UserID      uint           `json:"user_id" gorm:"not null"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	User       User         `json:"user,omitempty" gorm:"foreignKey:UserID"`
	OrderItems []OrderItem  `json:"order_items,omitempty" gorm:"foreignKey:ProductID"`
	CartItems  []CartItem   `json:"cart_items,omitempty" gorm:"foreignKey:ProductID"`
	Reviews    []Review     `json:"reviews,omitempty" gorm:"foreignKey:ProductID"`
}
