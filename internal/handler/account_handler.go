package handler

import (
	"errors"
	"exam/internal/service"
	"exam/internal/utils"
	"net/http"

	"github.com/labstack/echo/v4"
)

type AccountHandler struct {
	authService   *service.AuthService
	deviceService *service.DeviceService
}

func NewAccountHandler(authService *service.AuthService, deviceService *service.DeviceService) *AccountHandler {
	return &AccountHandler{authService: authService, deviceService: deviceService}
}

func (h *AccountHandler) GetAccountInfo(c echo.Context) error {
	userUUID, ok := c.Get("uuid").(string)
	if !ok {
		return utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid token claims")
	}

	user, err := h.authService.GetUserByUUID(userUUID)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			return utils.ErrorResponse(c, http.StatusNotFound, "User not found")
		}
		return utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve user information")
	}

	return utils.SuccessResponse(c, "Account information retrieved successfully", user)
}

func (h *AccountHandler) ListDevices(c echo.Context) error {
	userID := c.Get("userID").(uint)

	devices, err := h.deviceService.ListUserDevices(userID)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, "Devices retrieved successfully", devices)
}

func (h *AccountHandler) ForceDisconnect(c echo.Context) error {
	userID := c.Get("userID").(uint)
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
