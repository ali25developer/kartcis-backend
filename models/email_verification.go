package models

import (
"time"
)

type EmailVerification struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `json:"email" gorm:"index"`
	Token     string    `json:"token" gorm:"index"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}
