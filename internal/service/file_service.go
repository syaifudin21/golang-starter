package service

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jlaffaye/ftp"

	"exam/internal/model"
	"exam/internal/repository"
)

type FileService struct {
	uploadedFileRepository *repository.UploadedFileRepository
}

func NewFileService(uploadedFileRepository *repository.UploadedFileRepository) *FileService {
	return &FileService{uploadedFileRepository: uploadedFileRepository}
}

func (s *FileService) SaveFile(file *multipart.FileHeader, filename string, userID uint) (string, error) {
	// Create the uploads directory if it's not exist
	uploadsDir := "./uploads"
	if _, err := os.Stat(uploadsDir); os.IsNotExist(err) {
		err = os.MkdirAll(uploadsDir, os.ModePerm)
		if err != nil {
			return "", fmt.Errorf("failed to create uploads directory: %w", err)
		}
	}

	filePath := filepath.Join(uploadsDir, filename)

	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("failed to copy file content: %w", err)
	}

	var finalFilePath string
	var storedFilePath string // Declare storedFilePath here

	// --- FTP Upload ---
	ftpHost := os.Getenv("FTP_HOST")
	ftpPort := os.Getenv("FTP_PORT")
	ftpUser := os.Getenv("FTP_USER")
	ftpPass := os.Getenv("FTP_PASSWORD")
	ftpRemotePath := os.Getenv("FTP_REMOTE_PATH")
	ftpDomain := os.Getenv("FTP_DOMAIN")

	if ftpHost != "" && ftpPort != "" && ftpUser != "" && ftpPass != "" && ftpRemotePath != "" && ftpDomain != "" {
		port, err := strconv.Atoi(ftpPort)
		if err != nil {
			return "", fmt.Errorf("invalid FTP_PORT: %w", err)
		}

		conn, err := ftp.Dial(fmt.Sprintf("%s:%d", ftpHost, port), ftp.DialWithTimeout(5*time.Second))
		if err != nil {
			return "", fmt.Errorf("failed to connect to FTP server: %w", err)
		}
		defer conn.Quit()

		err = conn.Login(ftpUser, ftpPass)
		if err != nil {
			return "", fmt.Errorf("failed to login to FTP server: %w", err)
		}

		// Ensure the remote directory exists
		err = conn.ChangeDir(ftpRemotePath)
		if err != nil {
			// If directory doesn't exist, try to create it
			err = conn.MakeDir(ftpRemotePath)
			if err != nil {
				return "", fmt.Errorf("failed to change or create FTP remote directory: %w", err)
			}
			err = conn.ChangeDir(ftpRemotePath) // Change into the newly created directory
			if err != nil {
				return "", fmt.Errorf("failed to change into newly created FTP remote directory: %w", err)
			}
		}

		// Re-open the local file for FTP upload
		localFile, err := os.Open(filePath)
		if err != nil {
			return "", fmt.Errorf("failed to open local file for FTP upload: %w", err)
		}
		defer localFile.Close()

		ftpFileName := filepath.Base(filePath) // Use only the filename, not the full local path
		err = conn.Stor(ftpFileName, localFile)
		if err != nil {
			return "", fmt.Errorf("failed to upload file to FTP: %w", err)
		}

		// Delete the local file after successful FTP upload
		if err := os.Remove(filePath); err != nil {
			fmt.Printf("Warning: Failed to delete local file %s after FTP upload: %v\n", filePath, err)
		}

		finalFilePath = fmt.Sprintf("%s/%s/%s", ftpDomain, url.PathEscape(ftpRemotePath), url.PathEscape(ftpFileName))
		storedFilePath = fmt.Sprintf("/%s/%s", url.PathEscape(ftpRemotePath), url.PathEscape(ftpFileName)) // Store relative path
	} else {
		fmt.Println("FTP credentials not fully provided, skipping FTP upload. File saved locally.")
		finalFilePath = filePath // Return local path if FTP credentials are not set
		storedFilePath = filePath // Store local path
	}

	// Save file metadata to database
	uploadedFile := &model.UploadedFile{
		UUID:      uuid.New().String(),
		UserID:    userID,
		FileName:  file.Filename,
		FilePath:  storedFilePath, // Use storedFilePath here
		FileSize:  file.Size,
		MimeType:  file.Header.Get("Content-Type"),
	}

	if err := s.uploadedFileRepository.CreateUploadedFile(uploadedFile); err != nil {
		return "", fmt.Errorf("failed to save file metadata: %w", err)
	}

	return finalFilePath, nil
}

func (s *FileService) GetFilesByUserID(userID uint, limit, offset int) ([]model.UploadedFile, int64, error) {
	return s.uploadedFileRepository.GetUploadedFilesByUserID(userID, limit, offset)
}
