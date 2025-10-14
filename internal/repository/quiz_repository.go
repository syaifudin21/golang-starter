package repository

import (
	"exam/internal/model"

	"gorm.io/gorm"
)

type QuizRepository interface {
	CreateQuiz(quiz *model.Quiz) error
	AddQuestion(question *model.Question) error
	GetQuizByUUID(uuid string) (*model.Quiz, error)
	GetQuizWithQuestionsByUUID(uuid string) (*model.Quiz, error)
			ListAllQuizzes(keyword string, page, pageSize int) ([]model.Quiz, error)
			CountAllQuizzes(keyword string) (int64, error)
			UpdateQuiz(quiz *model.Quiz) error
			GetQuestionByUUID(uuid string) (*model.Question, error)
			UpdateQuestion(question *model.Question) error
			CreateQuizSession(session *model.QuizSession) error
			UpdateQuizSession(session *model.QuizSession) error
			GetQuizSessionByID(sessionID uint) (*model.QuizSession, error)
			CreateQuizAnswer(answer *model.QuizAnswer) error
		}
		
		
		type quizRepository struct {
			db *gorm.DB
		}
		
		func NewQuizRepository(db *gorm.DB) QuizRepository {
			return &quizRepository{db: db}
		}
		
		func (r *quizRepository) CreateQuiz(quiz *model.Quiz) error {
			return r.db.Create(quiz).Error
		}
		
		func (r *quizRepository) AddQuestion(question *model.Question) error {
			return r.db.Create(question).Error
		}
		
		func (r *quizRepository) GetQuizByUUID(uuid string) (*model.Quiz, error) {
			var quiz model.Quiz
			err := r.db.Where("uuid = ?", uuid).First(&quiz).Error
			if err != nil {
				return nil, err
			}
			return &quiz, nil
		}
		
		func (r *quizRepository) GetQuizWithQuestionsByUUID(uuid string) (*model.Quiz, error) {
			var quiz model.Quiz
			err := r.db.Preload("Questions").Preload("Creator").Where("uuid = ?", uuid).First(&quiz).Error
			if err != nil {
				return nil, err
			}
			return &quiz, nil
		}
		
		func (r *quizRepository) ListAllQuizzes(keyword string, page, pageSize int) ([]model.Quiz, error) {
			var quizzes []model.Quiz
			db := r.db.Model(&model.Quiz{}).Preload("Creator")
		
			if keyword != "" {
				searchKeyword := "%" + keyword + "%"
				db = db.Where("title LIKE ? OR description LIKE ?", searchKeyword, searchKeyword)
			}
		
			offset := (page - 1) * pageSize
			err := db.Limit(pageSize).Offset(offset).Find(&quizzes).Error
			if err != nil {
				return nil, err
			}
			return quizzes, nil
		}
		
		func (r *quizRepository) CountAllQuizzes(keyword string) (int64, error) {
			var count int64
			db := r.db.Model(&model.Quiz{})
		
			if keyword != "" {
				searchKeyword := "%" + keyword + "%"
				db = db.Where("title LIKE ? OR description LIKE ?", searchKeyword, searchKeyword)
			}
		
			err := db.Count(&count).Error
			return count, err
		}
		
		func (r *quizRepository) UpdateQuiz(quiz *model.Quiz) error {
			return r.db.Save(quiz).Error
		}
		
		func (r *quizRepository) GetQuestionByUUID(uuid string) (*model.Question, error) {
			var question model.Question
			err := r.db.Where("uuid = ?", uuid).First(&question).Error
			if err != nil {
				return nil, err
			}
			return &question, nil
		}
		
		func (r *quizRepository) UpdateQuestion(question *model.Question) error {
			return r.db.Save(question).Error
		}
		
		func (r *quizRepository) CreateQuizSession(session *model.QuizSession) error {
			return r.db.Create(session).Error
		}
		
		func (r *quizRepository) UpdateQuizSession(session *model.QuizSession) error {
			return r.db.Save(session).Error
		}
		
		func (r *quizRepository) GetQuizSessionByID(sessionID uint) (*model.QuizSession, error) {
			var session model.QuizSession
			err := r.db.First(&session, sessionID).Error
			if err != nil {
				return nil, err
			}
			return &session, nil
		}
		
		func (r *quizRepository) CreateQuizAnswer(answer *model.QuizAnswer) error {
			return r.db.Create(answer).Error
		}
