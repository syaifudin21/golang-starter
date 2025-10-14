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

func (s *FileService) SaveFile(file multipart.File, header *multipart.FileHeader, uniqueFilename string, customName string, userID uint) (string, error) {
	// Create the uploads directory if it's not exist
	uploadsDir := "./uploads"
	if _, err := os.Stat(uploadsDir); os.IsNotExist(err) {
		err = os.MkdirAll(uploadsDir, os.ModePerm)
		if err != nil {
			return "", fmt.Errorf("failed to create uploads directory: %w", err)
		}
	}

	filePath := filepath.Join(uploadsDir, uniqueFilename)

	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	if _, err = io.Copy(dst, file); err != nil {
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

		// Check if a temporary file from a previous failed upload exists and delete it
		tempFtpFileName := ".in." + ftpFileName
		if _, err := conn.FileSize(tempFtpFileName); err == nil {
			// File exists, attempt to delete it
			fmt.Printf("Warning: Temporary FTP file %s found, attempting to delete.\n", tempFtpFileName)
			if err := conn.Delete(tempFtpFileName); err != nil {
				fmt.Printf("Error: Failed to delete temporary FTP file %s: %v\n", tempFtpFileName, err)
				// Continue with upload, but log the error
			}
		}

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

	// Determine the name to save in the database
	dbFileName := header.Filename
	if customName != "" {
		dbFileName = customName
	}

	// Save file metadata to database
	uploadedFile := &model.UploadedFile{
		UUID:      uuid.New().String(),
		UserID:    userID,
		FileName:  dbFileName,
		FilePath:  storedFilePath, // Use storedFilePath here
		FileSize:  header.Size,
		MimeType:  header.Header.Get("Content-Type"),
	}

	if err := s.uploadedFileRepository.CreateUploadedFile(uploadedFile); err != nil {
		return "", fmt.Errorf("failed to save file metadata: %w", err)
	}

	return finalFilePath, nil
}

func (s *FileService) GetFilesByUserID(userID uint, mimeType string, limit, offset int) ([]model.UploadedFile, int64, error) {
	return s.uploadedFileRepository.GetUploadedFilesByUserID(userID, mimeType, limit, offset)
}

func (s *FileService) DeleteFile(fileUUID string, userID uint) error {
	// Get file metadata from database
	file, err := s.uploadedFileRepository.GetUploadedFileByUUIDAndUserID(fileUUID, userID)
	if err != nil {
		return fmt.Errorf("failed to get file metadata: %w", err)
	}
	if file == nil {
		return fmt.Errorf("file not found or not owned by user")
	}

	// Determine storage type and delete physical file
	ftpHost := os.Getenv("FTP_HOST")
	if ftpHost != "" { // Assuming FTP is configured if FTP_HOST is set
		// Delete from FTP
		ftpPort := os.Getenv("FTP_PORT")
		ftpUser := os.Getenv("FTP_USER")
		ftpPass := os.Getenv("FTP_PASSWORD")

		port, err := strconv.Atoi(ftpPort)
		if err != nil {
			return fmt.Errorf("invalid FTP_PORT: %w", err)
		}

		conn, err := ftp.Dial(fmt.Sprintf("%s:%d", ftpHost, port), ftp.DialWithTimeout(5*time.Second))
		if err != nil {
			return fmt.Errorf("failed to connect to FTP server: %w", err)
		}
		defer conn.Quit()

		err = conn.Login(ftpUser, ftpPass)
		if err != nil {
			return fmt.Errorf("failed to login to FTP server: %w", err)
		}

		// Use the full relative path from FilePath for FTP deletion
		err = conn.Delete(file.FilePath)
		if err != nil {
			return fmt.Errorf("failed to delete file from FTP: %w", err)
		}
	} else {
		// Delete from local storage
		// For local storage, file.FilePath is already the full path relative to the project root
		if err := os.Remove(file.FilePath); err != nil {
			return fmt.Errorf("failed to delete local file: %w", err)
		}
	}

	// Delete metadata from database
	if err := s.uploadedFileRepository.DeleteUploadedFile(fileUUID, userID); err != nil {
		return fmt.Errorf("failed to delete file metadata: %w", err)
	}

	return nil
}
