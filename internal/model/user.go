package model

import (
	"time"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UUID      string    `gorm:"type:varchar(36);uniqueIndex" json:"uuid"`
	Name      string    `gorm:"type:varchar(255)" json:"name"`
	Email     string    `gorm:"type:varchar(255);uniqueIndex" json:"email"`
	Phone     *string   `gorm:"type:varchar(20)" json:"phone,omitempty"`
	Password  string    `gorm:"type:varchar(255)" json:"-"`
	Role      string    `gorm:"type:varchar(50)" json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Devices   []Device  `gorm:"foreignKey:UserID" json:"devices,omitempty"`
}