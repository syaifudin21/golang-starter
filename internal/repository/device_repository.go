package repository

import (
	"exam/internal/model"
	"time"

	"gorm.io/gorm"
)

// DeviceRepository defines methods for interacting with device data in the database.
type DeviceRepository interface {
	CreateDevice(device *model.Device) error
	UpdateDevice(device *model.Device) error
	GetDeviceByRefreshToken(refreshToken string) (*model.Device, error)
	GetDeviceByJTI(jti string) (*model.Device, error)
	ListUserDevices(userID uint) ([]model.Device, error)
	ForceDisconnectDevice(userID uint, jti string) (int64, error)
}

type deviceRepository struct {
	db *gorm.DB
}

func NewDeviceRepository(db *gorm.DB) DeviceRepository {
	return &deviceRepository{db: db}
}

func (r *deviceRepository) CreateDevice(device *model.Device) error {
	return r.db.Create(device).Error
}

func (r *deviceRepository) UpdateDevice(device *model.Device) error {
	return r.db.Save(device).Error
}

func (r *deviceRepository) GetDeviceByRefreshToken(refreshToken string) (*model.Device, error) {
	var device model.Device
	err := r.db.Where("refresh_token = ?", refreshToken).First(&device).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &device, nil
}

func (r *deviceRepository) GetDeviceByJTI(jti string) (*model.Device, error) {
	var device model.Device
	err := r.db.Where("jti = ?", jti).First(&device).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &device, nil
}

func (r *deviceRepository) ListUserDevices(userID uint) ([]model.Device, error) {
	var devices []model.Device
	err := r.db.Where("user_id = ? AND logout_at IS NULL", userID).Find(&devices).Error
	return devices, err
}

func (r *deviceRepository) ForceDisconnectDevice(userID uint, jti string) (int64, error) {
	now := time.Now()
	result := r.db.Model(&model.Device{}).Where("jti = ? AND user_id = ?", jti, userID).Updates(map[string]interface{}{
		"logout_at":     &now,
		"refresh_token": nil,
	})
	return result.RowsAffected, result.Error
}