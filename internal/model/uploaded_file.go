package model

import (
	"time"
)

type UploadedFile struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UUID      string    `gorm:"type:char(36);uniqueIndex;not null" json:"uuid"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	FileName  string    `gorm:"type:varchar(255);not null" json:"file_name"`
	FilePath  string    `gorm:"type:varchar(255);not null" json:"file_path"`
	FileSize  int64     `gorm:"not null" json:"file_size"`
	MimeType  string    `gorm:"type:varchar(255);not null" json:"mime_type"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Associations
	User User `gorm:"foreignKey:UserID;references:ID" json:"-"`
}
