package service

import (
	"encoding/json"
	"exam/internal/dtos"
	"exam/internal/model"
	"exam/internal/repository"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// QuizRoomManager defines the interface for managing quiz rooms (e.g., getting student count).
type QuizRoomManager interface {
	GetRoomClientCount(quizUUID string) int
	GetRoomClients(quizUUID string) []dtos.ConnectedStudentDTO
	StartQuizInRoom(quizUUID string, sessionID uint, mode string) error
}

// ... (rest of QuizService struct and NewQuizService function)

func (s *QuizService) ListConnectedStudents(quizUUID string) []dtos.ConnectedStudentDTO {
	return s.hub.GetRoomClients(quizUUID)
}

func (s *QuizService) StartQuiz(quizUUID string, req dtos.StartQuizRequest) error {
	// First, check if the quiz exists and is valid
	quiz, err := s.quizRepo.GetQuizByUUID(quizUUID)
	if err != nil {
		return fmt.Errorf("failed to get quiz by UUID: %w", err)
	}
	if quiz == nil {
		return fmt.Errorf("quiz not found with UUID: %s", quizUUID)
	}

	// Create a new quiz session record
	now := time.Now()
	participantsJSON, _ := json.Marshal(s.hub.GetRoomClients(quizUUID))
	session := &model.QuizSession{
		QuizUUID:     quizUUID,
		Mode:         req.Mode,
		StartedAt:    now,
		Participants: datatypes.JSON(participantsJSON),
	}
	if err := s.quizRepo.CreateQuizSession(session); err != nil {
		return fmt.Errorf("failed to create quiz session: %w", err)
	}

	// Then, tell the hub to start the quiz in the room, passing the session ID and mode
	return s.hub.StartQuizInRoom(quizUUID, session.ID, req.Mode)
}

func (s *QuizService) EndQuizSession(sessionID uint, finalScores []dtos.PlayerScore) error {
	// Retrieve the session
	session, err := s.quizRepo.GetQuizSessionByID(sessionID)
	if err != nil {
		return fmt.Errorf("failed to get quiz session: %w", err)
	}
	if session == nil {
		return fmt.Errorf("quiz session %d not found", sessionID)
	}

	// Update session with end time and final scores
	now := time.Now()
	finalScoresJSON, _ := json.Marshal(finalScores)
	session.EndedAt = &now
	session.FinalScores = datatypes.JSON(finalScoresJSON)

	if err := s.quizRepo.UpdateQuizSession(session); err != nil {
		return fmt.Errorf("failed to update quiz session: %w", err)
	}

	return nil
}

type QuizService struct {
	quizRepo repository.QuizRepository
	hub      QuizRoomManager
}

func NewQuizService(quizRepo repository.QuizRepository, hub QuizRoomManager) *QuizService {
	return &QuizService{quizRepo: quizRepo, hub: hub}
}

// ... (rest of the file)

func (s *QuizService) GetStudentCount(quizUUID string) int {
	return s.hub.GetRoomClientCount(quizUUID)
}

func (s *QuizService) CreateQuiz(req dtos.CreateQuizRequest, teacherID uint) (*model.Quiz, error) {
	quiz := &model.Quiz{
		UUID:        uuid.New().String(),
		Title:       req.Title,
		Description: req.Description,
		CreatedBy:   teacherID,
	}

	if err := s.quizRepo.CreateQuiz(quiz); err != nil {
		return nil, fmt.Errorf("failed to create quiz: %w", err)
	}

	return quiz, nil
}

func (s *QuizService) AddQuestion(req dtos.AddQuestionRequest, quizUUID string) (*model.Question, error) {
	quiz, err := s.quizRepo.GetQuizByUUID(quizUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quiz by UUID: %w", err)
	}
	if quiz == nil {
		return nil, fmt.Errorf("quiz not found with UUID: %s", quizUUID)
	}

	contentJSON, err := json.Marshal(req.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal content: %w", err)
	}
	optionsJSON, err := json.Marshal(req.Options)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal options: %w", err)
	}

	question := &model.Question{
		UUID:          uuid.New().String(),
		QuizID:        quiz.ID,
		Content:       datatypes.JSON(contentJSON),
		Options:       datatypes.JSON(optionsJSON),
		CorrectAnswer: req.CorrectAnswer,
		Timer:         req.Timer,
	}

	if err := s.quizRepo.AddQuestion(question); err != nil {
		return nil, fmt.Errorf("failed to add question: %w", err)
	}

	return question, nil
}

func (s *QuizService) GetQuizWithQuestions(quizUUID string) (*model.Quiz, error) {
	quiz, err := s.quizRepo.GetQuizWithQuestionsByUUID(quizUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quiz: %w", err)
	}
	return quiz, nil
}

func (s *QuizService) ListAllQuizzes() ([]model.Quiz, error) {
	quizzes, err := s.quizRepo.ListAllQuizzes()
	if err != nil {
		return nil, fmt.Errorf("failed to list quizzes: %w", err)
	}
	return quizzes, nil
}

func (s *QuizService) UpdateQuiz(quizUUID string, req dtos.UpdateQuizRequest) (*model.Quiz, error) {
	quiz, err := s.quizRepo.GetQuizByUUID(quizUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quiz by UUID: %w", err)
	}
	if quiz == nil {
		return nil, fmt.Errorf("quiz not found with UUID: %s", quizUUID)
	}

	if req.Title != nil {
		quiz.Title = *req.Title
	}
	if req.Description != nil {
		quiz.Description = *req.Description
	}

	if err := s.quizRepo.UpdateQuiz(quiz); err != nil {
		return nil, fmt.Errorf("failed to update quiz: %w", err)
	}

	// Fetch the updated quiz with Creator preloaded for the response
	updatedQuiz, err := s.quizRepo.GetQuizWithQuestionsByUUID(quizUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated quiz with creator: %w", err)
	}

	return updatedQuiz, nil
}

func (s *QuizService) UpdateQuestion(quizUUID string, questionUUID string, req dtos.UpdateQuestionRequest) (*model.Question, error) {
	quiz, err := s.quizRepo.GetQuizByUUID(quizUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quiz by UUID: %w", err)
	}
	if quiz == nil {
		return nil, fmt.Errorf("quiz not found with UUID: %s", quizUUID)
	}

	questionToUpdate, err := s.quizRepo.GetQuestionByUUID(questionUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get question by UUID: %w", err)
	}
	if questionToUpdate == nil {
		return nil, fmt.Errorf("question not found with UUID: %s in quiz %s", questionUUID, quizUUID)
	}

	if req.Content != nil {
		contentJSON, err := json.Marshal(req.Content)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal content for update: %w", err)
		}
		questionToUpdate.Content = datatypes.JSON(contentJSON)
	}
	if req.Options != nil {
		optionsJSON, err := json.Marshal(req.Options)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal options for update: %w", err)
		}
		questionToUpdate.Options = datatypes.JSON(optionsJSON)
	}
	if req.CorrectAnswer != nil {
		questionToUpdate.CorrectAnswer = *req.CorrectAnswer
	}
	if req.Timer != nil {
		questionToUpdate.Timer = *req.Timer
	}

	if err := s.quizRepo.UpdateQuestion(questionToUpdate); err != nil {
		return nil, fmt.Errorf("failed to update question: %w", err)
	}

	return questionToUpdate, nil
}

func (s *QuizService) RecordQuizAnswer(answer *model.QuizAnswer) error {
	if err := s.quizRepo.CreateQuizAnswer(answer); err != nil {
		return fmt.Errorf("failed to record quiz answer: %w", err)
	}
	return nil
}