package repository

import (
	"exam/internal/model"

	"gorm.io/gorm"
)

type UploadedFileRepository struct {
	db *gorm.DB
}

func NewUploadedFileRepository(db *gorm.DB) *UploadedFileRepository {
	return &UploadedFileRepository{db: db}
}

func (r *UploadedFileRepository) CreateUploadedFile(uploadedFile *model.UploadedFile) error {
	return r.db.Create(uploadedFile).Error
}

func (r *UploadedFileRepository) GetUploadedFileByUUID(uuid string) (*model.UploadedFile, error) {
	var uploadedFile model.UploadedFile
	if err := r.db.Where("uuid = ?", uuid).First(&uploadedFile).Error; err != nil {
		return nil, err
	}
	return &uploadedFile, nil
}

func (r *UploadedFileRepository) GetUploadedFilesByUserID(userID uint, limit, offset int) ([]model.UploadedFile, int64, error) {
	var uploadedFiles []model.UploadedFile
	var totalCount int64

	// Get total count
	if err := r.db.Model(&model.UploadedFile{}).Where("user_id = ?", userID).Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated files
	if err := r.db.Where("user_id = ?", userID).Limit(limit).Offset(offset).Find(&uploadedFiles).Error; err != nil {
		return nil, 0, err
	}

	return uploadedFiles, totalCount, nil
}
