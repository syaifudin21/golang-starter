package routes

import (
	"exam/internal/handler"

	"github.com/labstack/echo/v4"
)

func APIRoutes(g *echo.Group, authHandler *handler.AuthHandler, accountHandler *handler.AccountHandler, userHandler *handler.UserHandler, quizHandler *handler.QuizHandler, websocketHandler *handler.WebsocketHandler) {
	g.GET("/account", accountHandler.GetAccountInfo)
	g.PUT("/account", userHandler.UpdateAccount)
	g.PUT("/password", userHandler.UpdatePassword)
	g.POST("/logout", authHandler.Logout)

	g.GET("/devices", accountHandler.ListDevices)
	g.DELETE("/devices/:jti", accountHandler.ForceDisconnect)

	g.GET("/users", userHandler.ListUsers)
	g.GET("/users/:uuid", userHandler.GetUser)
	g.PUT("/users/:uuid", userHandler.UpdateUserRole)

	// Quiz routes
	g.GET("/quizzes", quizHandler.ListQuizzes)
	g.GET("/quizzes/:quizUUID", quizHandler.GetQuiz)
	g.POST("/quizzes", quizHandler.CreateQuiz)
	g.PUT("/quizzes/:quizUUID", quizHandler.UpdateQuiz)
	g.POST("/quizzes/:quizUUID/questions", quizHandler.AddQuestion)
	g.PUT("/quizzes/:quizUUID/questions/:questionUUID", quizHandler.UpdateQuestion)
	g.GET("/quizzes/:quizUUID/students/count", quizHandler.GetStudentCount)
	g.GET("/quizzes/:quizUUID/students", quizHandler.ListStudents)
	g.POST("/quizzes/:quizUUID/start", quizHandler.StartQuiz)

	// Websocket route
	g.GET("/quiz/join/:quizUUID", websocketHandler.ServeWs)
}
