package handler

import (
	"errors"
	"exam/internal/dtos"
	"exam/internal/service"
	"exam/internal/utils"
	"net/http"

	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Login(c echo.Context) error {
	req := new(dtos.LoginRequest)
	if err := c.Bind(req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	// Basic validation (can be enhanced with a validation library)
	if req.Email == "" || req.Password == "" {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Email and password are required")
	}

	// Extract User-Agent for device_info
	deviceInfo := c.Request().UserAgent()

	resp, err := h.authService.Login(*req, deviceInfo)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			return utils.ErrorResponse(c, http.StatusUnauthorized, err.Error())
		}
		return utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, "Login successful", resp)
}

func (h *AuthHandler) Logout(c echo.Context) error {
	jti, ok := c.Get("jti").(string)
	if !ok || jti == "" {
		return utils.ErrorResponse(c, http.StatusInternalServerError, "JTI not found in token claims")
	}

	if err := h.authService.LogoutDevice(jti); err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, "Logged out successfully", nil)
}

func (h *AuthHandler) Refresh(c echo.Context) error {
	req := new(dtos.RefreshTokenRequest)
	if err := c.Bind(req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	// Extract User-Agent for device_info
	deviceInfo := c.Request().UserAgent()

	resp, err := h.authService.RefreshAccessToken(req.RefreshToken, deviceInfo)
	if err != nil {
		if errors.Is(err, service.ErrInvalidRefreshToken) || errors.Is(err, service.ErrRefreshTokenLoggedOut) {
			return utils.ErrorResponse(c, http.StatusUnauthorized, err.Error())
		}
		return utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, "Token refreshed successfully", resp)
}
