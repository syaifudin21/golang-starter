package model

import "time"

// Quiz represents a collection of questions created by a teacher
type Quiz struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	UUID        string     `gorm:"type:varchar(36);uniqueIndex" json:"uuid"`
	Title       string     `gorm:"type:varchar(255)" json:"title"`
	Description string     `gorm:"type:text" json:"description"`
	CreatedBy   uint       `json:"created_by"` // Foreign key to User ID
	Creator     User       `gorm:"foreignKey:CreatedBy" json:"creator"`
	Questions   []Question `gorm:"foreignKey:QuizID" json:"questions,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
