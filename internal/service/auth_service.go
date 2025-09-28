package service

import (
	"encoding/json"
	"errors"
	"exam/internal/dtos"
	"exam/internal/model"
	"exam/internal/repository"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Custom error types
var (
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrUserNotFound          = errors.New("user not found")
	ErrOldPasswordMismatch   = errors.New("old password does not match")
	ErrRefreshTokenLoggedOut = errors.New("refresh token has been logged out")
	ErrInvalidRefreshToken   = errors.New("invalid refresh token")
	ErrDatabase              = errors.New("database error")
	ErrDeviceNotFound        = errors.New("device not found")
	ErrDeviceNotOwned        = errors.New("device not owned by user")
	ErrEmailExists           = errors.New("email already exists")
)

type AuthService struct {
	userRepo   repository.UserRepository
	deviceRepo repository.DeviceRepository
}

func NewAuthService(userRepo repository.UserRepository, deviceRepo repository.DeviceRepository) *AuthService {
	return &AuthService{userRepo: userRepo, deviceRepo: deviceRepo}
}

func (s *AuthService) Register(req dtos.RegisterRequest) (*model.User, error) {
	existingUser, err := s.userRepo.GetUserByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
	}
	if existingUser != nil {
		return nil, ErrEmailExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &model.User{
		UUID:     uuid.New().String(),
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     "student", // Default role
	}

	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	return user, nil
}

func (s *AuthService) Login(req dtos.LoginRequest, deviceInfo string) (*dtos.LoginResponse, error) {
	user, err := s.userRepo.GetUserByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.generateAndSaveTokens(user, req.FCMToken, req.Latitude, req.Longitude, deviceInfo)
}

func (s *AuthService) LogoutDevice(jti string) error {
	device, err := s.deviceRepo.GetDeviceByJTI(jti)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabase, err)
	}
	if device == nil {
		return ErrDeviceNotFound
	}

	now := time.Now()
	device.LogoutAt = &now
	device.RefreshToken = nil

	if err := s.deviceRepo.UpdateDevice(device); err != nil {
		return fmt.Errorf("%w: failed to record device logout: %v", ErrDatabase, err)
	}
	return nil
}

func (s *AuthService) RefreshAccessToken(oldRefreshToken string, deviceInfo string) (*dtos.LoginResponse, error) {
	device, err := s.deviceRepo.GetDeviceByRefreshToken(oldRefreshToken)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
	}
	if device == nil {
		return nil, ErrInvalidRefreshToken
	}

	if device.LogoutAt != nil {
		return nil, ErrRefreshTokenLoggedOut
	}

	user, err := s.userRepo.GetUserByID(device.UserID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
	}
	if user == nil {
		return nil, fmt.Errorf("%w: user not found for refresh token", ErrUserNotFound)
	}

	return s.generateAndSaveTokens(user, device.FCMToken, device.Latitude, device.Longitude, deviceInfo)
}

func (s *AuthService) UpdateUserAccount(userID uint, name string, phone *string) error {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabase, err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	user.Name = name
	user.Phone = phone

	return s.userRepo.UpdateUser(user)
}

func (s *AuthService) UpdateUserPassword(userID uint, oldPassword string, newPassword string) error {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabase, err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword))
	if err != nil {
		return ErrOldPasswordMismatch
	}

	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
			if err != nil {
				return fmt.Errorf("failed to hash new password: %w", err)
			}
	user.Password = string(hashedNewPassword)

	return s.userRepo.UpdateUser(user)
}

func (s *AuthService) ListAllUsers() ([]model.User, error) {
	users, err := s.userRepo.ListAllUsers()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
	}
	return users, nil
}

func (s *AuthService) GetUserByUUID(userUUID string) (*model.User, error) {
	user, err := s.userRepo.GetUserByUUID(userUUID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	devices, err := s.deviceRepo.ListUserDevices(user.ID)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to list devices for user: %v", ErrDatabase, err)
	}
	user.Devices = devices

	return user, nil
}

func (s *AuthService) UpdateUserRole(userUUID string, newRole string) error {
	user, err := s.userRepo.GetUserByUUID(userUUID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabase, err)
	}
	if user == nil {
		return ErrUserNotFound
	}

	user.Role = newRole

	return s.userRepo.UpdateUser(user)
}

// GetGoogleUserInfo fetches user info from Google using the access token.
func (s *AuthService) GetGoogleUserInfo(accessToken string) (*dtos.GoogleUserInfo, error) {
	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info from Google: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info from Google, status: %s", resp.Status)
	}

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Google user info response: %w", err)
	}

	var userInfo dtos.GoogleUserInfo
	if err := json.Unmarshal(contents, &userInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Google user info: %w", err)
	}

	return &userInfo, nil
}

// LoginWithGoogle handles user login/registration via Google OAuth.
func (s *AuthService) LoginWithGoogle(userInfo *dtos.GoogleUserInfo, deviceInfo string) (*dtos.LoginResponse, error) {
	user, err := s.userRepo.GetUserByEmail(userInfo.Email)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	if user == nil {
		// User does not exist, register them
		user = &model.User{
			UUID:     uuid.New().String(),
			Name:     userInfo.Name,
			Email:    userInfo.Email,
			Password: "", // No password for OAuth users
			Role:     "student", // Default role
		}
		if err := s.userRepo.CreateUser(user); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
		}
	}

	// Generate and save tokens for the user
	return s.generateAndSaveTokens(user, nil, nil, nil, deviceInfo)
}

// generateAndSaveTokens is a helper to create JWT and refresh tokens and save device info.
func (s *AuthService) generateAndSaveTokens(user *model.User, fcmToken *string, latitude *float64, longitude *float64, deviceInfo string) (*dtos.LoginResponse, error) {
	jti := uuid.New().String()
	refreshToken := uuid.New().String()

	device := &model.Device{
		UserID:       user.ID,
		JTI:          jti,
		RefreshToken: &refreshToken,
		DeviceInfo:   deviceInfo,
		FCMToken:     fcmToken,
		Latitude:     latitude,
		Longitude:    longitude,
		LoginAt:      time.Now(),
	}
	if err := s.deviceRepo.CreateDevice(device); err != nil {
		return nil, fmt.Errorf("%w: failed to record device login: %v", ErrDatabase, err)
	}

	claims := jwt.MapClaims{
		"id":    user.ID,
		"uuid":  user.UUID,
		"email": user.Email,
		"role":  user.Role,
		"exp":   time.Now().Add(time.Hour * 72).Unix(),
		"jti":   jti,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	jwtSecret := os.Getenv("JWT_SECRET")
	accessToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign new access token: %w", err)
	}

	return &dtos.LoginResponse{Token: accessToken, RefreshToken: refreshToken}, nil
}