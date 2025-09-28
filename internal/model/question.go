package model

import (
	"time"

	"gorm.io/datatypes"
)

// Question represents a single question in a quiz
type Question struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	UUID           string         `gorm:"type:varchar(36);uniqueIndex" json:"uuid"`
	QuizID         uint           `json:"quiz_id"`
	Content        datatypes.JSON `gorm:"type:json" json:"content"` // Changed from QuestionText
	Options        datatypes.JSON `gorm:"type:json" json:"options"`
	CorrectAnswer  string         `gorm:"type:varchar(255)" json:"correct_answer"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}