package utils

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

type Response struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// JSONResponse sends a standardized JSON response.
func JSONResponse(c echo.Context, statusCode int, message string, data interface{}) error {
	return c.JSON(statusCode, Response{
		Message: message,
		Data:    data,
	})
}

// SuccessResponse sends a standardized success JSON response.
func SuccessResponse(c echo.Context, message string, data interface{}) error {
	return JSONResponse(c, http.StatusOK, message, data)
}

// ErrorResponse sends a standardized error JSON response.
func ErrorResponse(c echo.Context, statusCode int, message string) error {
	return JSONResponse(c, statusCode, message, nil)
}
