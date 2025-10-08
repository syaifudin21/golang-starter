package handler

import (
	"fmt"
	"net/http"
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

	// Generate a unique filename
	extension := filepath.Ext(file.Filename)
	newFileName := uuid.New().String() + extension

	filePath, err := h.fileService.SaveFile(file, newFileName, userID)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to save file: %v", err))
	}

	return utils.SuccessResponse(c, "File uploaded successfully", echo.Map{"filePath": filePath})
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

	offset := (page - 1) * pageSize

	files, totalCount, err := h.fileService.GetFilesByUserID(userID, pageSize, offset)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to retrieve files: %v", err))
	}

	totalPages := (totalCount + int64(pageSize) - 1) / int64(pageSize)

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