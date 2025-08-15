package models

import (
	"time"

	"gorm.io/gorm"
)

type Message struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	FromUserID uint           `json:"from_user_id" gorm:"not null"`
	ToUserID   uint           `json:"to_user_id" gorm:"not null"`
	Content    string         `json:"content" gorm:"not null"`
	IsRead     bool           `json:"is_read" gorm:"default:false"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	FromUser User `json:"from_user,omitempty" gorm:"foreignKey:FromUserID"`
	ToUser   User `json:"to_user,omitempty" gorm:"foreignKey:ToUserID"`
}
