package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UploadFile handles image upload
func UploadFile(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "No image file provided"})
		return
	}

	// Validate extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid file type. Only jpg, jpeg, png, webp allowed"})
		return
	}

	// Generate filename
	filename := fmt.Sprintf("%d-%s%s", time.Now().Unix(), uuid.New().String(), ext)
	dst := filepath.Join("uploads", filename)

	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to save file"})
		return
	}

	// Generate URL
	// Assuming API_BASE_URL env or constructed from host
	// For simple setup: /api/v1/uploads/filename if static route is set
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, c.Request.Host)

	// Check route prefix
	apiPrefix := os.Getenv("API_PREFIX")
	if apiPrefix == "" {
		apiPrefix = "/api/v1"
	}

	// static route is defined as v1.Static("/uploads", ...)
	// So URL is baseURL + apiPrefix + "/uploads/" + filename
	fileURL := fmt.Sprintf("%s%s/uploads/%s", baseURL, apiPrefix, filename)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "File uploaded successfully",
		"data": gin.H{
			"url":      fileURL,
			"filename": filename,
		},
	})
}
