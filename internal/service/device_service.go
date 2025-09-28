package service

import (
	"exam/internal/model"
	"exam/internal/repository"
	"fmt"
)

type DeviceService struct {
	deviceRepo repository.DeviceRepository
}

func NewDeviceService(deviceRepo repository.DeviceRepository) *DeviceService {
	return &DeviceService{deviceRepo: deviceRepo}
}

func (s *DeviceService) ListUserDevices(userID uint) ([]model.Device, error) {
	devices, err := s.deviceRepo.ListUserDevices(userID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
	}
	return devices, nil
}

func (s *DeviceService) ForceDisconnectDevice(userID uint, jti string) error {
	rowsAffected, err := s.deviceRepo.ForceDisconnectDevice(userID, jti)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	if rowsAffected == 0 {
		// With GORM, if the record is not found, RowsAffected is 0 and there's no separate error.
		// We can infer that the device was either not found or not owned by the user.
		// For simplicity, we can group these. A more detailed check could be done if necessary.
		return ErrDeviceNotOwned // This covers both not found and not owned cases now.
	}

	return nil
}
