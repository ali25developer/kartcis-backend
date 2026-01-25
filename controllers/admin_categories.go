package controllers

import (
	"kartcis-backend/config"
	"kartcis-backend/models"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Helper for slug generation
func generateCategorySlug(name string) string {
	slug := strings.ToLower(name)
	re := regexp.MustCompile(`[^a-z0-9]+`)
	slug = re.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug
}

// AdminGetCategories
func AdminGetCategories(c *gin.Context) {
	categories := []models.Category{}
	var totalItems int64

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	query := config.DB.Model(&models.Category{})

	includeInactive := c.Query("include_inactive")
	if includeInactive != "true" {
		query = query.Where("is_active = ?", true)
	}

	search := c.Query("search")
	if search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Count Total
	query.Count(&totalItems)

	// Fetch Data
	if err := query.Order("display_order asc").Limit(limit).Offset(offset).Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to fetch categories"})
		return
	}

	totalPages := int(totalItems) / limit
	if int(totalItems)%limit != 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"categories": categories,
			"pagination": gin.H{
				"current_page": page,
				"total_pages":  totalPages,
				"total_items":  totalItems,
				"per_page":     limit,
			},
		},
	})
}

// CreateCategory
func CreateCategory(c *gin.Context) {
	var input models.Category
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid input", "errors": err.Error()})
		return
	}

	// Basic validation
	if input.Name == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"success": false, "message": "Validation failed", "errors": gin.H{"name": "Name is required"}})
		return
	}

	// Auto-generate slug if missing
	if input.Slug == "" {
		input.Slug = generateCategorySlug(input.Name)
	}

	if err := config.DB.Create(&input).Error; err != nil {
		// Handle duplicate slug error
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Failed to create category (slug might be taken)", "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Category created successfully", "data": input})
}

// AdminGetCategoryDetail
func AdminGetCategoryDetail(c *gin.Context) {
	id := c.Param("id")
	var category models.Category

	if err := config.DB.First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Category not found"})
		return
	}

	// Should include stats like total_revenue, etc. per spec
	// Should include stats like total_revenue, etc. per spec
	// Calculating event count from database
	var eventCount int64
	config.DB.Model(&models.Event{}).Where("category_id = ?", category.ID).Count(&eventCount)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"id":            category.ID,
			"name":          category.Name,
			"slug":          category.Slug,
			"description":   category.Description,
			"icon":          category.Icon,
			"image":         category.Image,
			"is_active":     category.IsActive,
			"display_order": category.DisplayOrder,
			"event_count":   eventCount,
			"created_at":    category.CreatedAt,
			"updated_at":    category.UpdatedAt,
			// "events": ... list of events, can implement later
		},
	})
}

// UpdateCategory
func UpdateCategory(c *gin.Context) {
	id := c.Param("id")
	var category models.Category

	if err := config.DB.First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Category not found"})
		return
	}

	var input models.Category
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid input"})
		return
	}

	// Update fields
	category.Name = input.Name
	category.Slug = input.Slug
	category.Description = input.Description
	category.Icon = input.Icon
	category.Image = input.Image
	category.DisplayOrder = input.DisplayOrder
	category.IsActive = input.IsActive // Explicitly update IsActive
	category.UpdatedAt = time.Now()

	if err := config.DB.Save(&category).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Failed to update category", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Category updated successfully", "data": category})
}

// DeleteCategory
func DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	var category models.Category

	if err := config.DB.First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Category not found"})
		return
	}

	// Check if used in events
	var count int64
	config.DB.Model(&models.Event{}).Where("category_id = ?", id).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Cannot delete category linked to events. Deactivate it instead."})
		return
	}

	config.DB.Delete(&category)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Category deleted successfully"})
}

// UpdateCategoryStatus
func UpdateCategoryStatus(c *gin.Context) {
	id := c.Param("id")
	var category models.Category

	if err := config.DB.First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Category not found"})
		return
	}

	type StatusInput struct {
		IsActive bool `json:"is_active"`
	}
	var input StatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid input"})
		return
	}

	category.IsActive = input.IsActive
	category.UpdatedAt = time.Now()
	config.DB.Save(&category)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Category status updated", "data": category})
}

// ReorderCategories
func ReorderCategories(c *gin.Context) {
	type OrderItem struct {
		ID           int `json:"id"`
		DisplayOrder int `json:"display_order"`
	}
	type ReorderInput struct {
		Categories []OrderItem `json:"categories"`
	}

	var input ReorderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid input"})
		return
	}

	tx := config.DB.Begin()
	for _, item := range input.Categories {
		if err := tx.Model(&models.Category{}).Where("id = ?", item.ID).Update("display_order", item.DisplayOrder).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to Update Order"})
			return
		}
	}
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Categories reordered successfully"})
}
