package handler

import (
	"exam/internal/dtos"
	"exam/internal/service"
	"exam/internal/utils"
	"net/http"

	"github.com/labstack/echo/v4"
)

type QuizHandler struct {
	quizService *service.QuizService
}

func NewQuizHandler(quizService *service.QuizService) *QuizHandler {
	return &QuizHandler{quizService: quizService}
}

func (h *QuizHandler) CreateQuiz(c echo.Context) error {
	req := new(dtos.CreateQuizRequest)
	if err := c.Bind(req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	lang := c.Request().Header.Get("Accept-Language")
	if msg, ok := utils.ValidateStruct(req, lang); !ok {
		return utils.ErrorResponse(c, http.StatusBadRequest, msg)
	}

	teacherID := c.Get("userID").(uint)

	quiz, err := h.quizService.CreateQuiz(*req, teacherID)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, "Quiz created successfully", quiz)
}

func (h *QuizHandler) AddQuestion(c echo.Context) error {
	quizUUID := c.Param("quizUUID")

	req := new(dtos.AddQuestionRequest)
	if err := c.Bind(req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	lang := c.Request().Header.Get("Accept-Language")
	if msg, ok := utils.ValidateStruct(req, lang); !ok {
		return utils.ErrorResponse(c, http.StatusBadRequest, msg)
	}

	question, err := h.quizService.AddQuestion(*req, quizUUID)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, "Question added successfully", question)
}

func (h *QuizHandler) ListQuizzes(c echo.Context) error {
	quizzes, err := h.quizService.ListAllQuizzes()
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}
	return utils.SuccessResponse(c, "Quizzes retrieved successfully", quizzes)
}

func (h *QuizHandler) GetQuiz(c echo.Context) error {
	quizUUID := c.Param("quizUUID")
	if quizUUID == "" {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Quiz UUID is required")
	}

	quiz, err := h.quizService.GetQuizWithQuestions(quizUUID)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}
	return utils.SuccessResponse(c, "Quiz retrieved successfully", quiz)
}

func (h *QuizHandler) UpdateQuiz(c echo.Context) error {
	quizUUID := c.Param("quizUUID")
	if quizUUID == "" {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Quiz UUID is required")
	}

	req := new(dtos.UpdateQuizRequest)
	if err := c.Bind(req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	lang := c.Request().Header.Get("Accept-Language")
	if msg, ok := utils.ValidateStruct(req, lang); !ok {
		return utils.ErrorResponse(c, http.StatusBadRequest, msg)
	}

	updatedQuiz, err := h.quizService.UpdateQuiz(quizUUID, *req)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, "Quiz updated successfully", updatedQuiz)
}

func (h *QuizHandler) UpdateQuestion(c echo.Context) error {
	quizUUID := c.Param("quizUUID")
	if quizUUID == "" {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Quiz UUID is required")
	}

	questionUUID := c.Param("questionUUID")
	if questionUUID == "" {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Question UUID is required")
	}

	req := new(dtos.UpdateQuestionRequest)
	if err := c.Bind(req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	lang := c.Request().Header.Get("Accept-Language")
	if msg, ok := utils.ValidateStruct(req, lang); !ok {
		return utils.ErrorResponse(c, http.StatusBadRequest, msg)
	}

	updatedQuestion, err := h.quizService.UpdateQuestion(quizUUID, questionUUID, *req)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, "Question updated successfully", updatedQuestion)
}

func (h *QuizHandler) GetStudentCount(c echo.Context) error {
	quizUUID := c.Param("quizUUID")
	if quizUUID == "" {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Quiz UUID is required")
	}

	count := h.quizService.GetStudentCount(quizUUID)
	return utils.SuccessResponse(c, "Student count retrieved successfully", map[string]int{"count": count})
}

func (h *QuizHandler) ListStudents(c echo.Context) error {
	quizUUID := c.Param("quizUUID")
	if quizUUID == "" {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Quiz UUID is required")
	}

	students := h.quizService.ListConnectedStudents(quizUUID)
	return utils.SuccessResponse(c, "Connected students retrieved successfully", students)
}

func (h *QuizHandler) StartQuiz(c echo.Context) error {
	quizUUID := c.Param("quizUUID")
	if quizUUID == "" {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Quiz UUID is required")
	}

	req := new(dtos.StartQuizRequest)
	if err := c.Bind(req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	lang := c.Request().Header.Get("Accept-Language")
	if msg, ok := utils.ValidateStruct(req, lang); !ok {
		return utils.ErrorResponse(c, http.StatusBadRequest, msg)
	}

	err := h.quizService.StartQuiz(quizUUID, *req)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, "Quiz started successfully", nil)
}