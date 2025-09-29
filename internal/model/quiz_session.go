package model

import (
	"time"

	"gorm.io/datatypes"
)

// QuizSession represents a single instance of a quiz being played.
type QuizSession struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	QuizUUID    string         `gorm:"type:varchar(36);not null" json:"quiz_uuid"`
	Mode        string         `gorm:"type:varchar(20);not null;default:'sync'" json:"mode"`
	StartedAt   time.Time      `json:"started_at"`
	EndedAt     *time.Time     `json:"ended_at,omitempty"`
	Participants datatypes.JSON `gorm:"type:json" json:"participants"` // Stores JSON array of ConnectedStudentDTO
	FinalScores datatypes.JSON `gorm:"type:json" json:"final_scores"`   // Stores JSON array of PlayerScore
}