package service

import (
	"exam/internal/dtos"
	"exam/internal/model"
	"exam/internal/repository"
	"errors"
	"fmt"
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
)

type AuthService struct {
	userRepo   repository.UserRepository
	deviceRepo repository.DeviceRepository
}

func NewAuthService(userRepo repository.UserRepository, deviceRepo repository.DeviceRepository) *AuthService {
	return &AuthService{userRepo: userRepo, deviceRepo: deviceRepo}
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

	// Generate JTI and Refresh Token
	jti := uuid.New().String()
	refreshToken := uuid.New().String()

	// Insert device info into devices table
	device := &model.Device{
		UserID:     user.ID,
		JTI:        jti,
		RefreshToken: &refreshToken,
		DeviceInfo: deviceInfo,
		FCMToken:   req.FCMToken,
		Latitude:   req.Latitude,
		Longitude:  req.Longitude,
	}
	err = s.deviceRepo.InsertDevice(device)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to record device login: %v", ErrDatabase, err)
	}

	// Generate JWT token
	claims := jwt.MapClaims{
		"id":    user.ID,
		"uuid":  user.UUID,
		"email": user.Email,
		"role":  user.Role,
		"exp":   time.Now().Add(time.Hour * 72).Unix(), // Token expires in 72 hours
		"jti":   jti,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		// Fallback for development, but should be set in .env for production
		jwtSecret = "supersecretjwtkey"
	}

	t, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign token: %w", err)
	}

	return &dtos.LoginResponse{Token: t, RefreshToken: refreshToken}, nil
}

func (s *AuthService) LogoutDevice(jti string) error {
	// Also clear the refresh_token when logging out
	err := s.deviceRepo.UpdateDeviceLogout(jti)
	if err != nil {
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

	// Get user details for new access token claims
	user, err := s.userRepo.GetUserByID(device.UserID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
	}
	if user == nil {
		return nil, fmt.Errorf("%w: user not found for refresh token", ErrUserNotFound)
	}

	// Generate new JTI and Access Token
	newJTI := uuid.New().String()
	newRefreshToken := uuid.New().String()

	claims := jwt.MapClaims{
		"id":    user.ID,
		"uuid":  user.UUID,
		"email": user.Email,
		"role":  user.Role,
		"exp":   time.Now().Add(time.Hour * 72).Unix(), // New access token expires in 72 hours
		"jti":   newJTI,
	}

	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "supersecretjwtkey"
	}

	accessToken, err := newToken.SignedString([]byte(jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign new access token: %w", err)
	}

	// Update device record with new JTI and Refresh Token
	device.JTI = newJTI
	device.RefreshToken = &newRefreshToken // Assign pointer to string
	device.DeviceInfo = deviceInfo
	device.UpdatedAt = time.Now()

	err = s.deviceRepo.UpdateDevice(device)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to update device record for refresh: %v", ErrDatabase, err)
	}

	return &dtos.LoginResponse{Token: accessToken, RefreshToken: newRefreshToken}, nil
}

func (s *AuthService) UpdateUserAccount(userID int, name string, phone *string) error {
	err := s.userRepo.UpdateUser(userID, name, phone)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabase, err)
	}
	return nil
}

func (s *AuthService) UpdateUserPassword(userID int, oldPassword string, newPassword string) error {
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

	err = s.userRepo.UpdatePassword(userID, string(hashedNewPassword))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	return nil
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

	// Fetch devices for the user
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

	// TODO: Add validation for newRole (e.g., check against a list of valid roles)

	err = s.userRepo.UpdateUserRole(user.ID, newRole)
	if err != nil {
		return fmt.Errorf("%w: failed to update user role: %v", ErrDatabase, err)
	}

	return nil
}
