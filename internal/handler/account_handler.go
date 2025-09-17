package handler

import (
	"database/sql"
	"errors"
	"exam/internal/model"
	"exam/internal/service"
	"exam/internal/utils"
	"net/http"

	"github.com/labstack/echo/v4"
)

type AccountHandler struct {
	db *sql.DB
	authService *service.AuthService
	deviceService *service.DeviceService
}

func NewAccountHandler(db *sql.DB, authService *service.AuthService, deviceService *service.DeviceService) *AccountHandler {
	return &AccountHandler{db: db, authService: authService, deviceService: deviceService}
}

func (h *AccountHandler) GetAccountInfo(c echo.Context) error {
	userID := int(c.Get("userID").(float64))

	user := &model.User{}
	err := h.db.QueryRow("SELECT id, uuid, name, email, phone, role, created_at, updated_at FROM users WHERE id = ?", userID).Scan(
		&user.ID,
		&user.UUID,
		&user.Name,
		&user.Email,
		&user.Phone,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return utils.ErrorResponse(c, http.StatusNotFound, "User not found")
		}
		return utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve user information")
	}

	return utils.SuccessResponse(c, "Account information retrieved successfully", user)
}

func (h *AccountHandler) ListDevices(c echo.Context) error {
	userID := int(c.Get("userID").(float64))

	devices, err := h.deviceService.ListUserDevices(userID)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, "Devices retrieved successfully", devices)
}

func (h *AccountHandler) ForceDisconnect(c echo.Context) error {
	userID := int(c.Get("userID").(float64))
	jti := c.Param("jti")

	if jti == "" {
		return utils.ErrorResponse(c, http.StatusBadRequest, "JTI is required")
	}

	if err := h.deviceService.ForceDisconnectDevice(userID, jti); err != nil {
		if errors.Is(err, service.ErrDeviceNotFound) {
			return utils.ErrorResponse(c, http.StatusNotFound, err.Error())
		}
		if errors.Is(err, service.ErrDeviceNotOwned) {
			return utils.ErrorResponse(c, http.StatusForbidden, err.Error())
		}
		return utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, "Device disconnected successfully", nil)
}
