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

	// Ensure directory exists
	if err := os.MkdirAll("uploads", 0755); err != nil {
		fmt.Printf("Warning: failed to create uploads directory: %v\n", err)
	}

	if err := c.SaveUploadedFile(file, dst); err != nil {
		fmt.Printf("Upload fails on SaveUploadedFile: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": fmt.Sprintf("Failed to save file: %v", err)})
		return
	}

	// Ensure the file is readable by others (Nginx, etc.) in production
	if err := os.Chmod(dst, 0644); err != nil {
		fmt.Printf("Warning: failed to chmod file %s: %v\n", dst, err)
	}

	// Generate URL
	// Assuming API_BASE_URL env or constructed from host
	// For simple setup: /api/v1/uploads/filename if static route is set
	scheme := "http"
	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, c.Request.Host)

	// If there's an explicit API_URL in the .env, use it as the base
	if envAPIURL := os.Getenv("API_URL"); envAPIURL != "" {
		baseURL = envAPIURL
	}

	// Check route prefix
	apiPrefix := os.Getenv("API_PREFIX")
	if apiPrefix == "" {
		apiPrefix = "/api/v1"
	}

	// static route is defined as v1.Static("/uploads", ...)
	// So URL is baseURL + apiPrefix + "/uploads/" + filename
	// Example: http://kartcis.id/api/v1/uploads/filename
	fileURL := fmt.Sprintf("%s%s/uploads/%s", baseURL, apiPrefix, filename)

	// If the constructed base URL already ends with /api/v1, adjust it so we don't duplicate
	if strings.HasSuffix(baseURL, apiPrefix) {
		fileURL = fmt.Sprintf("%s/uploads/%s", baseURL, filename)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "File uploaded successfully",
		"data": gin.H{
			"url":      fileURL,
			"filename": filename,
		},
	})
}
