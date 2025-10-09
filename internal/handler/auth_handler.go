package handler

import (
	"errors"
	"exam/internal/dtos"
	"exam/internal/service"
	"exam/internal/utils"
	"net/http"

	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
)

type AuthHandler struct {
	authService *service.AuthService
	googleOauthConfig *oauth2.Config
}

func NewAuthHandler(authService *service.AuthService, googleOauthConfig *oauth2.Config) *AuthHandler {
	return &AuthHandler{authService: authService, googleOauthConfig: googleOauthConfig}
}

func (h *AuthHandler) Register(c echo.Context) error {
	req := new(dtos.RegisterRequest)
	if err := c.Bind(req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	lang := c.Request().Header.Get("Accept-Language")
	if msg, ok := utils.ValidateStruct(req, lang); !ok {
		return utils.ErrorResponse(c, http.StatusBadRequest, msg)
	}

	user, err := h.authService.Register(*req)
	if err != nil {
		if errors.Is(err, service.ErrEmailExists) {
			return utils.ErrorResponse(c, http.StatusConflict, err.Error())
		}
		return utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, "Registration successful", user)
}

func (h *AuthHandler) Login(c echo.Context) error {
	req := new(dtos.LoginRequest)
	if err := c.Bind(req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	lang := c.Request().Header.Get("Accept-Language")
	if msg, ok := utils.ValidateStruct(req, lang); !ok {
		return utils.ErrorResponse(c, http.StatusBadRequest, msg)
	}

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

	lang := c.Request().Header.Get("Accept-Language")
	if msg, ok := utils.ValidateStruct(req, lang); !ok {
		return utils.ErrorResponse(c, http.StatusBadRequest, msg)
	}

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

func (h *AuthHandler) GoogleLogin(c echo.Context) error {
	url := h.googleOauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *AuthHandler) GoogleCallback(c echo.Context) error {
	state := c.QueryParam("state")
	if state != "state" {
		return utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid state parameter")
	}

	code := c.QueryParam("code")
	if code == "" {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Authorization code not provided")
	}

	token, err := h.googleOauthConfig.Exchange(c.Request().Context(), code)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to exchange code for token: "+err.Error())
	}

	// Use the token to get user info from Google
	userInfo, err := h.authService.GetGoogleUserInfoFromAccessToken(token.AccessToken)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get user info from Google: "+err.Error())
	}

	// Find or create user in our DB
	loginResp, err := h.authService.LoginWithGoogle(userInfo, c.Request().UserAgent())
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to login/register with Google: "+err.Error())
	}

	return utils.SuccessResponse(c, "Google login successful", loginResp)
}

func (h *AuthHandler) GoogleLoginWithToken(c echo.Context) error {
	req := new(dtos.GoogleLoginRequest)
	if err := c.Bind(req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	lang := c.Request().Header.Get("Accept-Language")
	if msg, ok := utils.ValidateStruct(req, lang); !ok {
		return utils.ErrorResponse(c, http.StatusBadRequest, msg)
	}

	deviceInfo := c.Request().UserAgent()

	resp, err := h.authService.GoogleLoginWithToken(req.Credential, deviceInfo)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
	}

	return utils.SuccessResponse(c, "Google login successful", resp)
}
