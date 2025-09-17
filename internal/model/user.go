package model

import (
	"time"
)

type User struct {
	ID        int       `json:"id"`
	UUID      string    `json:"uuid"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     *string   `json:"phone"` // Added Phone field, nullable
	Password  string    `json:"-"` // Exclude from JSON output
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Devices   []Device  `json:"devices"` // Added Devices field
}
