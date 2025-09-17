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

func (s *DeviceService) ListUserDevices(userID int) ([]model.Device, error) {
	devices, err := s.deviceRepo.ListUserDevices(userID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabase, err)
	}
	return devices, nil
}

func (s *DeviceService) ForceDisconnectDevice(userID int, jti string) error {
	rowsAffected, err := s.deviceRepo.ForceDisconnectDevice(userID, jti)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabase, err)
	}

	if rowsAffected == 0 {
		// Check if the device exists at all
		exists, err := s.deviceRepo.CheckDeviceExistence(jti)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrDatabase, err)
		}
		if !exists {
			return ErrDeviceNotFound
		}
		// If it exists but rowsAffected is 0, it means it's not owned by the user
		return ErrDeviceNotOwned
	}

	return nil
}