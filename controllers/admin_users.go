package controllers

import (
	"kartcis-backend/config"
	"kartcis-backend/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Get All Users
func AdminGetUsers(c *gin.Context) {
	users := []models.User{}
	var totalItems int64

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	query := config.DB.Model(&models.User{})

	// Filters
	search := c.Query("search")
	if search != "" {
		query = query.Where("name ILIKE ? OR email ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	role := c.Query("role")
	if role != "" {
		query = query.Where("role = ?", role)
	}

	status := c.Query("status")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Count Total
	query.Count(&totalItems)

	// Fetch Data
	query.Order("created_at desc").Limit(limit).Offset(offset).Find(&users)

	totalPages := int(totalItems) / limit
	if int(totalItems)%limit != 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"users": users,
			"pagination": gin.H{
				"current_page": page,
				"total_pages":  totalPages,
				"total_items":  totalItems,
				"per_page":     limit,
			},
		},
	})
}

// Get User Detail
func AdminGetUserDetail(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "User not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": user})
}

// Create User (Admin)
func AdminCreateUser(c *gin.Context) {
	var input models.User
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid input"})
		return
	}

	// Hash password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	input.Password = string(hashedPassword)

	if err := config.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": input})
}

// Update User (Admin)
func AdminUpdateUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "User not found"})
		return
	}

	var input models.User
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid input"})
		return
	}

	user.Name = input.Name
	user.Email = input.Email
	user.Role = input.Role
	// Skip password update here unless specific logic

	config.DB.Save(&user)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": user})
}

// Delete User
func AdminDeleteUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "User not found"})
		return
	}

	config.DB.Delete(&user)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "User deleted"})
}

// UpdateUserRole
func AdminUpdateUserRole(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Role is required"})
		return
	}

	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "User not found"})
		return
	}

	user.Role = input.Role
	config.DB.Save(&user)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "User role updated"})
}

// UpdateUserStatus
func AdminUpdateUserStatus(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Status is required"})
		return
	}

	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "User not found"})
		return
	}

	// User model now has Status field
	user.Status = input.Status
	config.DB.Save(&user)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "User status updated"})
}

// GetUserActivity
func AdminGetUserActivity(c *gin.Context) {
	id := c.Param("id")

	logs := []models.ActivityLog{}
	if err := config.DB.Where("user_id = ?", id).Order("created_at desc").Limit(50).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to fetch activity logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    logs,
	})
}

// GetUserTransactions
func AdminGetUserTransactions(c *gin.Context) {
	id := c.Param("id")
	orders := []models.Order{}
	if err := config.DB.Where("user_id = ?", id).Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to fetch transactions"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": orders})
}
