package models

import (
	"time"

	"gorm.io/gorm"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusShipped   OrderStatus = "shipped"
	OrderStatusDelivered OrderStatus = "delivered"
	OrderStatusCancelled OrderStatus = "cancelled"
	OrderStatusInProcess OrderStatus = "in_process"
)

type Order struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	UserID        uint           `json:"user_id" gorm:"not null"`
	Status        OrderStatus    `json:"status" gorm:"default:'pending'"`
	TotalAmount   float64        `json:"total_amount" gorm:"not null"`
	ShippingAddress string       `json:"shipping_address" gorm:"not null"`
	PaymentIntentID string       `json:"payment_intent_id"`
	StripeSessionID string       `json:"stripe_session_id"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	User       User        `json:"user,omitempty" gorm:"foreignKey:UserID"`
	OrderItems []OrderItem `json:"order_items,omitempty" gorm:"foreignKey:OrderID"`
}

type OrderItem struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	OrderID   uint           `json:"order_id" gorm:"not null"`
	ProductID uint           `json:"product_id" gorm:"not null"`
	Quantity  int            `json:"quantity" gorm:"not null"`
	Price     float64        `json:"price" gorm:"not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Order   Order   `json:"order,omitempty" gorm:"foreignKey:OrderID"`
	Product Product `json:"product,omitempty" gorm:"foreignKey:ProductID"`
}
