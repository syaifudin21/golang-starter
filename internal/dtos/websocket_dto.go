package dtos

import "encoding/json"

// WebsocketMessage is the generic structure for all websocket messages.
type WebsocketMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// --- Client-to-Server Payloads ---

// SubmitAnswerPayload is the payload for a 'submit_answer' message.
type SubmitAnswerPayload struct {
	QuestionID uint   `json:"question_id"`
	Answer     string `json:"answer"`
}

// --- Server-to-Client Payloads ---

// PlayerInfoPayload is used for player join/leave notifications.
type PlayerInfoPayload struct {
	UserID   uint   `json:"user_id"`
	UserName string `json:"user_name"`
}

// QuizQuestionDTO is a safe version of model.Question to be sent to clients (without the correct answer).
type QuizQuestionDTO struct {
	ID      uint            `json:"id"`
	Content json.RawMessage `json:"content"`
	Options json.RawMessage `json:"options"`
}

// AnswerResultPayload announces the result of an answer submission.
type AnswerResultPayload struct {
	QuestionID    uint   `json:"question_id"`
	IsCorrect     bool   `json:"is_correct"`
	PlayerID      uint   `json:"player_id"` // The player who answered
	PlayerName    string `json:"player_name"`
	IsFirstAnswer bool   `json:"is_first_answer"`
}

// PlayerScore holds the score for a single player.
type PlayerScore struct {
	UserID   uint   `json:"user_id"`
	UserName string `json:"user_name"`
	Score    int    `json:"score"`
}

// ScoreUpdatePayload sends the current leaderboard.
type ScoreUpdatePayload struct {
	Scores []PlayerScore `json:"scores"`
}

// GameOverPayload announces the end of the game.
type GameOverPayload struct {
	Winner  PlayerScore   `json:"winner"`
	Scores  []PlayerScore `json:"scores"`
}

// ErrorPayload sends an error message to a client.
type ErrorPayload struct {
	Message string `json:"message"`
}

// ConnectedStudentDTO represents a connected student in a quiz room.
type ConnectedStudentDTO struct {
	UserID   uint   `json:"user_id"`
	UserName string `json:"user_name"`
}
