package model

import (
	"time"
)

// QuizAnswer represents a single answer submitted by a participant for a question in a quiz session.
type QuizAnswer struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	QuizSessionID uint      `gorm:"not null" json:"quiz_session_id"`
	QuestionID    uint      `gorm:"not null" json:"question_id"`
	UserID        uint      `gorm:"not null" json:"user_id"`
	Answer        string    `gorm:"type:varchar(255)" json:"answer"` // The ID of the option chosen by the user
	IsCorrect     bool      `gorm:"not null" json:"is_correct"`
	SubmittedAt   time.Time `gorm:"not null" json:"submitted_at"`
}