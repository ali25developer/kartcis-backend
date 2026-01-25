package controllers

import (
	"kartcis-backend/config"
	"kartcis-backend/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetCategories godoc
// @Summary Get all categories
// @Description Get list of active categories
// @Tags categories
// @Produce json
// @Success 200 {object} object
// @Router /categories [get]
func GetCategories(c *gin.Context) {
	categories := []models.Category{}
	var totalItems int64

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "12"))
	offset := (page - 1) * limit

	query := config.DB.Model(&models.Category{}).Where("is_active = ?", true)

	// Count Total
	query.Count(&totalItems)

	// Fetch Data
	if err := query.Order("display_order asc").Limit(limit).Offset(offset).Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to fetch categories", "error": err.Error()})
		return
	}

	type CategoryResponse struct {
		models.Category
		EventCount int64 `json:"event_count"`
	}

	response := []CategoryResponse{}
	for _, cat := range categories {
		var count int64
		config.DB.Model(&models.Event{}).Where("category_id = ? AND status = ?", cat.ID, "published").Count(&count)
		response = append(response, CategoryResponse{
			Category:   cat,
			EventCount: count,
		})
	}

	totalPages := int(totalItems) / limit
	if int(totalItems)%limit != 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"categories": response,
			"pagination": gin.H{
				"current_page": page,
				"total_pages":  totalPages,
				"total_items":  totalItems,
				"per_page":     limit,
			},
		},
	})
}

// GetCategoryDetail godoc
// @Summary Get category detail
// @Description Get category details by slug
// @Tags categories
// @Produce json
// @Param slug path string true "Category Slug"
// @Success 200 {object} object
// @Router /categories/{slug} [get]
func GetCategoryDetail(c *gin.Context) {
	slug := c.Param("slug")
	var category models.Category

	// 1. Try find by Slug
	if err := config.DB.Where("slug = ? AND is_active = ?", slug, true).First(&category).Error; err != nil {
		// 2. Fallback: If slug is numeric, try finding by ID
		if id, errConv := strconv.Atoi(slug); errConv == nil {
			if err := config.DB.Where("id = ? AND is_active = ?", id, true).First(&category).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Category not found"})
				return
			}
		} else {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Category not found"})
			return
		}
	}

	var count int64
	config.DB.Model(&models.Event{}).Where("category_id = ? AND status = ?", category.ID, "published").Count(&count)

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
			"event_count":   count,
			"created_at":    category.CreatedAt,
			"updated_at":    category.UpdatedAt,
		},
	})
}
