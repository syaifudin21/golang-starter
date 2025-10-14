package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"exam/internal/service"
	"exam/internal/utils"
)

type FileHandler struct {
	fileService *service.FileService
}

func NewFileHandler(fileService *service.FileService) *FileHandler {
	return &FileHandler{fileService: fileService}
}

// UploadFile handles file uploads
// @Summary Upload a file
// @Description Uploads a single file to the server
// @Tags File
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "File to upload"
// @Success 200 {object} utils.Response "File uploaded successfully"
// @Failure 400 {object} utils.Response "Bad request"
// @Failure 500 {object} utils.Response "Internal server error"
// @Router /upload [post]
func (h *FileHandler) UploadFile(c echo.Context) error {
	userID, ok := c.Get("userID").(uint)
	if !ok {
		return utils.ErrorResponse(c, http.StatusUnauthorized, "User ID not found in context")
	}

	file, err := c.FormFile("file")
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Failed to get file from form: %v", err))
	}

	// Open the file to check its content type
	src, err := file.Open()
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to open file: %v", err))
	}
	defer src.Close()

	// Read the first 512 bytes to determine the content type
	buffer := make([]byte, 512)
	_, err = src.Read(buffer)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to read file: %v", err))
	}

	// Reset the file read pointer
	_, err = src.Seek(0, 0)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to reset file read pointer: %v", err))
	}

	// Get the content type
	contentType := http.DetectContentType(buffer)

	// Get file extension
	extension := filepath.Ext(file.Filename)

	// If content type is generic, try to determine from extension
	if contentType == "application/octet-stream" {
		switch extension {
		case ".mp3":
			contentType = "audio/mpeg"
		case ".wav":
			contentType = "audio/wav"
		case ".ogg":
			contentType = "audio/ogg"
		case ".mp4":
			contentType = "audio/mp4"
		}
	}

	// Check if the content type is allowed
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
		"audio/mpeg": true,
		"audio/wav":  true,
		"audio/ogg":  true,
		"audio/mp4":  true,
	}

	if !allowedTypes[contentType] {
		return utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("File type not allowed: %s", contentType))
	}

	// Generate a unique filename for storage
	uniqueFileName := uuid.New().String() + extension

	// Get custom file name from form data
	customName := c.FormValue("name")

	filePath, err := h.fileService.SaveFile(src, file, uniqueFileName, customName, userID)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to save file: %v", err))
	}

	return utils.SuccessResponse(c, "File uploaded successfully", echo.Map{"filePath": filePath, "name": customName})
}

// GetMyFiles retrieves all files uploaded by the authenticated user with pagination
// @Summary Get all uploaded files for the current user with pagination
// @Description Retrieves a paginated list of all files uploaded by the user whose token is provided in the Authorization header.
// @Tags File
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Number of items per page" default(10)
// @Success 200 {object} utils.Response{data=[]model.UploadedFile} "Paginated list of uploaded files"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Failure 500 {object} utils.Response "Internal server error"
// @Router /files [get]
func (h *FileHandler) GetMyFiles(c echo.Context) error {
	userID, ok := c.Get("userID").(uint)
	if !ok {
		return utils.ErrorResponse(c, http.StatusUnauthorized, "User ID not found in context")
	}

	// Get pagination parameters
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page <= 0 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(c.QueryParam("pageSize"))
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10 // Default page size
	}

	mimeType := c.QueryParam("mime")

	offset := (page - 1) * pageSize

	files, totalCount, err := h.fileService.GetFilesByUserID(userID, mimeType, pageSize, offset)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve files: %v", err))
	}

	totalPages := (totalCount + int64(pageSize) - 1) / int64(pageSize)

	ftpDomain := os.Getenv("FTP_DOMAIN")
	if ftpDomain == "" {
		fmt.Println("Warning: FTP_DOMAIN environment variable is not set. File paths might be incomplete.")
	}

	// Prepend FTP_DOMAIN to FilePath for each file
	for i := range files {
		files[i].FilePath = fmt.Sprintf("%s%s", ftpDomain, files[i].FilePath)
	}

	return c.JSON(http.StatusOK, echo.Map{
		"message": "Files retrieved successfully",
		"data": files,
		"pagination": echo.Map{
			"totalCount": totalCount,
			"totalPages": totalPages,
			"currentPage": page,
			"pageSize": pageSize,
		},
	})
}

// DeleteFile handles the deletion of an uploaded file
// @Summary Delete an uploaded file
// @Description Deletes a specific file uploaded by the authenticated user.
// @Tags File
// @Produce json
// @Param uuid path string true "File UUID"
// @Success 200 {object} utils.Response "File deleted successfully"
// @Failure 400 {object} utils.Response "Bad request"
// @Failure 401 {object} utils.Response "Unauthorized"
// @Failure 404 {object} utils.Response "File not found"
// @Failure 500 {object} utils.Response "Internal server error"
// @Router /files/{uuid} [delete]
func (h *FileHandler) DeleteFile(c echo.Context) error {
	userID, ok := c.Get("userID").(uint)
	if !ok {
		return utils.ErrorResponse(c, http.StatusUnauthorized, "User ID not found in context")
	}

	fileUUID := c.Param("uuid")
	if fileUUID == "" {
		return utils.ErrorResponse(c, http.StatusBadRequest, "File UUID is required")
	}

	if err := h.fileService.DeleteFile(fileUUID, userID); err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to delete file: %v", err))
	}

	return utils.SuccessResponse(c, "File deleted successfully", nil)
}