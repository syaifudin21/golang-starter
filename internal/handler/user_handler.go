package handler

import (
	"errors"
	"exam/internal/dtos"
	"exam/internal/service"
	"exam/internal/utils"
	"net/http"

	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	authService *service.AuthService
}

func NewUserHandler(authService *service.AuthService) *UserHandler {
	return &UserHandler{authService: authService}
}

func (h *UserHandler) UpdateAccount(c echo.Context) error {
	userID := int(c.Get("userID").(float64))

	req := new(dtos.UpdateAccountRequest)
	if err := c.Bind(req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	if req.Name == "" {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Name is required")
	}

	if err := h.authService.UpdateUserAccount(int(userID), req.Name, req.Phone); err != nil {
		if errors.Is(err, service.ErrDatabase) {
			return utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		}
		return utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, "Account updated successfully", nil)
}

func (h *UserHandler) UpdatePassword(c echo.Context) error {
	userID := int(c.Get("userID").(float64))

	req := new(dtos.UpdatePasswordRequest)
	if err := c.Bind(req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Old password and new password are required")
	}

	if err := h.authService.UpdateUserPassword(int(userID), req.OldPassword, req.NewPassword); err != nil {
		if errors.Is(err, service.ErrUserNotFound) || errors.Is(err, service.ErrOldPasswordMismatch) {
			return utils.ErrorResponse(c, http.StatusUnauthorized, err.Error())
		}
		return utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, "Password updated successfully", nil)
}

func (h *UserHandler) ListUsers(c echo.Context) error {
	users, err := h.authService.ListAllUsers()
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, "Users retrieved successfully", users)
}

func (h *UserHandler) GetUser(c echo.Context) error {
	userUUID := c.Param("uuid") // Get UUID from URL parameter
	if userUUID == "" {
		return utils.ErrorResponse(c, http.StatusBadRequest, "User UUID is required")
	}

	user, err := h.authService.GetUserByUUID(userUUID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return utils.ErrorResponse(c, http.StatusNotFound, err.Error())
		}
		return utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, "User retrieved successfully", user)
}

func (h *UserHandler) UpdateUserRole(c echo.Context) error {
	userUUID := c.Param("uuid")
	if userUUID == "" {
		return utils.ErrorResponse(c, http.StatusBadRequest, "User UUID is required")
	}

	req := new(dtos.UpdateUserRoleRequest)
	if err := c.Bind(req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	if req.Role == "" {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Role is required")
	}

	if err := h.authService.UpdateUserRole(userUUID, req.Role); err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return utils.ErrorResponse(c, http.StatusNotFound, err.Error())
		}
		return utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, "User role updated successfully", nil)
}
