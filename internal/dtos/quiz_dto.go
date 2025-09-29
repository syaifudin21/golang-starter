package dtos

// CreateQuizRequest defines the structure for creating a new quiz.
type QuestionContent struct {
	Type  string `json:"type"`  // e.g., "text", "image", "audio"
	Value string `json:"value"` // The text, URL for image/audio
}

type QuestionOption struct {
	ID    string `json:"id"`
	Type  string `json:"type"`  // e.g., "text", "image", "audio"
	Value string `json:"value"` // The text, URL for image/audio
}

// CreateQuizRequest defines the structure for creating a new quiz.
type CreateQuizRequest struct {
	Title       string `json:"title" validate:"required,min=5"`
	Description string `json:"description"`
}

// AddQuestionRequest defines the structure for adding a new question to a quiz.
type AddQuestionRequest struct {
	Content       QuestionContent  `json:"content" validate:"required"`
	Options       []QuestionOption `json:"options" validate:"required"`
	CorrectAnswer string           `json:"correct_answer" validate:"required"`
	Timer         int              `json:"timer" validate:"required,min=0"`
}

// UpdateQuizRequest defines the structure for updating an existing quiz.
type UpdateQuizRequest struct {
	Title       *string `json:"title" validate:"omitempty,min=5"`
	Description *string `json:"description"`
}

// UpdateQuestionRequest defines the structure for updating an existing question.
type UpdateQuestionRequest struct {
	Content       *QuestionContent  `json:"content" validate:"omitempty"`
	Options       []QuestionOption `json:"options" validate:"omitempty"`
	CorrectAnswer *string           `json:"correct_answer" validate:"omitempty"`
	Timer         *int             `json:"timer,omitempty" validate:"omitempty,min=0"`
}

// StartQuizRequest defines the structure for starting a quiz.
type StartQuizRequest struct {
	Mode string `json:"mode" validate:"required,oneof=sync parallel"`
}