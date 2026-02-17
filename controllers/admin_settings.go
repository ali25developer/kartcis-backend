package controllers

import (
	"kartcis-backend/config"
	"kartcis-backend/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GetSettings (Public) - Returns key-value JSON
func GetSettings(c *gin.Context) {
	var settings []models.SiteSetting
	if err := config.DB.Find(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to fetch settings"})
		return
	}

	// Transform to Key-Value map
	settingsMap := make(map[string]string)
	for _, s := range settings {
		settingsMap[s.Key] = s.Value
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    settingsMap,
	})
}

// UpdateSettings (Admin) - Updates specific keys
func UpdateSettings(c *gin.Context) {
	var input map[string]string
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	tx := config.DB.Begin()

	for key, value := range input {
		var setting models.SiteSetting

		// UPSERT Logic: Update if exists, Create if not
		if err := tx.Where("key = ?", key).First(&setting).Error; err != nil {
			// Create
			newSetting := models.SiteSetting{
				Key:   key,
				Value: value,
			}
			if err := tx.Create(&newSetting).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to create setting: " + key})
				return
			}
		} else {
			// Update
			setting.Value = value
			setting.UpdatedAt = time.Now()
			if err := tx.Save(&setting).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to update setting: " + key})
				return
			}
		}
	}

	tx.Commit()

	// Log Activity
	if userID, exists := c.Get("userID"); exists {
		config.DB.Create(&models.ActivityLog{
			UserID:    userID.(uint),
			Action:    "update_settings",
			IPAddress: c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
			Details:   "Updated site settings",
			CreatedAt: time.Now(),
		})
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Settings updated successfully"})
}
