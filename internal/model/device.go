package model

import (
	"time"
)

type Device struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	UserID       uint       `json:"user_id"`
	JTI          string     `gorm:"type:varchar(255);uniqueIndex" json:"jti"`
	RefreshToken *string    `gorm:"type:text" json:"refresh_token"`
	DeviceInfo   string     `gorm:"type:text" json:"device_info"`
	FCMToken     *string    `gorm:"type:text" json:"fcm_token"`
	LoginAt      time.Time  `json:"login_at"`
	LogoutAt     *time.Time `json:"logout_at"`
	Latitude     *float64   `json:"latitude"`
	Longitude    *float64   `json:"longitude"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}