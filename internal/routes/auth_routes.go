package routes

import (
	"exam/internal/handler"

	"github.com/labstack/echo/v4"
)

func AuthRoutes(e *echo.Echo, authHandler *handler.AuthHandler) {
	e.POST("/login", authHandler.Login)
	e.POST("/refresh", authHandler.Refresh)
}
