package repository

import (
	"database/sql"
	"errors"
	"exam/internal/model"
	"fmt"
	"time"
)

// DeviceRepository defines methods for interacting with device data in the database.
type DeviceRepository interface {
	InsertDevice(device *model.Device) error
	UpdateDeviceLogout(jti string) error
	UpdateDevice(device *model.Device) error // Added UpdateDevice method
	GetDeviceByRefreshToken(refreshToken string) (*model.Device, error)
	GetDeviceByJTI(jti string) (*model.Device, error) // Added GetDeviceByJTI method
	ListUserDevices(userID int) ([]model.Device, error)
	ForceDisconnectDevice(userID int, jti string) (int64, error) // Returns rows affected
	CheckDeviceExistence(jti string) (bool, error)
}

type deviceRepository struct {
	db *sql.DB
}

func NewDeviceRepository(db *sql.DB) DeviceRepository {
	return &deviceRepository{db: db}
}

func (r *deviceRepository) InsertDevice(device *model.Device) error {
	_, err := r.db.Exec(
		"INSERT INTO devices (user_id, jti, refresh_token, device_info, fcm_token, latitude, longitude) VALUES (?, ?, ?, ?, ?, ?, ?)",
		device.UserID,
		device.JTI,
		device.RefreshToken,
		device.DeviceInfo,
		device.FCMToken,
		device.Latitude,
		device.Longitude,
	)
	if err != nil {
		return fmt.Errorf("failed to insert device: %w", err)
	}
	return nil
}

func (r *deviceRepository) UpdateDeviceLogout(jti string) error {
	_, err := r.db.Exec("UPDATE devices SET logout_at = ?, refresh_token = NULL WHERE jti = ?", time.Now(), jti)
	if err != nil {
		return fmt.Errorf("failed to update device logout: %w", err)
	}
	return nil
}

func (r *deviceRepository) UpdateDevice(device *model.Device) error {
	_, err := r.db.Exec("UPDATE devices SET jti = ?, refresh_token = ?, device_info = ?, updated_at = ? WHERE id = ?",
		device.JTI,
		device.RefreshToken,
		device.DeviceInfo,
		device.UpdatedAt,
		device.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update device: %w", err)
	}
	return nil
}

func (r *deviceRepository) GetDeviceByRefreshToken(refreshToken string) (*model.Device, error) {
	device := &model.Device{}
	err := r.db.QueryRow("SELECT id, user_id, jti, refresh_token, logout_at FROM devices WHERE refresh_token = ?", refreshToken).Scan(
		&device.ID,
		&device.UserID,
		&device.JTI,
		&device.RefreshToken,
		&device.LogoutAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Device not found, return nil device and nil error
		}
		return nil, fmt.Errorf("failed to get device by refresh token: %w", err)
	}
	return device, nil
}

func (r *deviceRepository) GetDeviceByJTI(jti string) (*model.Device, error) {
	device := &model.Device{}
	err := r.db.QueryRow("SELECT id, user_id, jti, refresh_token, logout_at FROM devices WHERE jti = ?", jti).Scan(
		&device.ID,
		&device.UserID,
		&device.JTI,
		&device.RefreshToken,
		&device.LogoutAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Device not found, return nil device and nil error
		}
		return nil, fmt.Errorf("failed to get device by JTI: %w", err)
	}
	return device, nil
}

func (r *deviceRepository) ListUserDevices(userID int) ([]model.Device, error) {
	rows, err := r.db.Query("SELECT id, user_id, jti, refresh_token, device_info, fcm_token, login_at, logout_at, latitude, longitude, created_at, updated_at FROM devices WHERE user_id = ? AND logout_at IS NULL", userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query devices: %w", err)
	}
	defer rows.Close()

	var devices []model.Device
	for rows.Next() {
		var device model.Device
		err := rows.Scan(
			&device.ID,
			&device.UserID,
			&device.JTI,
			&device.RefreshToken,
			&device.DeviceInfo,
			&device.FCMToken,
			&device.LoginAt,
			&device.LogoutAt,
			&device.Latitude,
			&device.Longitude,
			&device.CreatedAt,
			&device.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device: %w", err)
		}
		devices = append(devices, device)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return devices, nil
}

func (r *deviceRepository) ForceDisconnectDevice(userID int, jti string) (int64, error) {
	result, err := r.db.Exec("UPDATE devices SET logout_at = ?, refresh_token = NULL WHERE jti = ? AND user_id = ?", time.Now(), jti, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to update device for disconnect: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

func (r *deviceRepository) CheckDeviceExistence(jti string) (bool, error) {
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM devices WHERE jti = ?)", jti).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check device existence: %w", err)
	}
	return exists, nil
}
