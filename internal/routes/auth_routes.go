package routes

import (
	"exam/internal/handler"

	"github.com/labstack/echo/v4"
)

func AuthRoutes(e *echo.Echo, authHandler *handler.AuthHandler) {
	e.POST("/register", authHandler.Register)
	e.POST("/login", authHandler.Login)
	e.POST("/refresh", authHandler.Refresh)

	// Google OAuth routes
	e.GET("/auth/google/login", authHandler.GoogleLogin)
	e.GET("/auth/google/callback", authHandler.GoogleCallback)
}
