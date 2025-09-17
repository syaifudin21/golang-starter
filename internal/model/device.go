package model

import (
	"time"
)

type Device struct {
	ID           int        `json:"id"`
	UserID       int        `json:"user_id"`
	JTI          string     `json:"jti"`
	RefreshToken *string    `json:"refresh_token"` // Changed to *string
	DeviceInfo   string     `json:"device_info"`
	FCMToken     *string    `json:"fcm_token"` // Use pointer for nullable string
	LoginAt      time.Time  `json:"login_at"`
	LogoutAt     *time.Time `json:"logout_at"` // Use pointer for nullable time
	Latitude     *float64   `json:"latitude"`  // Use pointer for nullable decimal
	Longitude    *float64   `json:"longitude"` // Use pointer for nullable decimal
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
