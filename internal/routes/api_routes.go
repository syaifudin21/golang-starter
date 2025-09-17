package routes

import (
	"exam/internal/handler"

	"github.com/labstack/echo/v4"
)

func APIRoutes(g *echo.Group, authHandler *handler.AuthHandler, accountHandler *handler.AccountHandler, userHandler *handler.UserHandler) {
	g.GET("/account", accountHandler.GetAccountInfo)
	g.PUT("/account", userHandler.UpdateAccount)
	g.PUT("/password", userHandler.UpdatePassword)
	g.POST("/logout", authHandler.Logout)

	g.GET("/devices", accountHandler.ListDevices)
	g.DELETE("/devices/:jti", accountHandler.ForceDisconnect)

	g.GET("/users", userHandler.ListUsers)
	g.GET("/users/:uuid", userHandler.GetUser)
	g.PUT("/users/:uuid", userHandler.UpdateUserRole)
}
