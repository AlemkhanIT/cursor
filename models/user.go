package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID                uint           `json:"id" gorm:"primaryKey"`
	Email             string         `json:"email" gorm:"unique;not null"`
	Password          string         `json:"-" gorm:"not null"`
	FirstName         string         `json:"first_name" gorm:"not null"`
	LastName          string         `json:"last_name" gorm:"not null"`
	IsEmailConfirmed  bool           `json:"is_email_confirmed" gorm:"default:false"`
	EmailConfirmToken string         `json:"-"`
	ResetPasswordToken string        `json:"-"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Products []Product `json:"products,omitempty" gorm:"foreignKey:UserID"`
	Orders   []Order   `json:"orders,omitempty" gorm:"foreignKey:UserID"`
	Reviews  []Review  `json:"reviews,omitempty" gorm:"foreignKey:UserID"`
	SentMessages     []Message `json:"sent_messages,omitempty" gorm:"foreignKey:FromUserID"`
	ReceivedMessages []Message `json:"received_messages,omitempty" gorm:"foreignKey:ToUserID"`
}
