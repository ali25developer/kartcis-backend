package controllers

import (
	"net/http"
	"strconv"
	"time"

	"kartcis-backend/config"
	"kartcis-backend/models"

	"github.com/gin-gonic/gin"
)

func GetEvents(c *gin.Context) {
	events := []models.Event{}
	var totalItems int64

	// Parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "12"))
	search := c.Query("search")
	category := c.Query("category")
	featured := c.Query("featured")

	offset := (page - 1) * limit

	query := config.DB.Model(&models.Event{}).Preload("Category").Preload("TicketTypes")

	// Filters
	query = query.Where("status = ?", "published") // Public API only shows published

	if search != "" {
		query = query.Where("title ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if category != "" {
		// Support filtering by ID if numeric
		if id, err := strconv.Atoi(category); err == nil {
			query = query.Where("category_id = ?", id)
		} else {
			// Filter by Slug or Name
			query = query.Joins("JOIN categories ON events.category_id = categories.id").
				Where("categories.slug = ? OR categories.name ILIKE ?", category, category)
		}
	}

	if featured == "true" {
		query = query.Where("is_featured = ?", true)
	}

	// Count Total
	query.Count(&totalItems)

	// Fetch Data
	query.Limit(limit).Offset(offset).Order("event_date ASC").Find(&events)

	// Pagination
	totalPages := int(totalItems) / limit
	if int(totalItems)%limit != 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"events": events,
			"pagination": gin.H{
				"current_page": page,
				"total_pages":  totalPages,
				"total_items":  totalItems,
				"per_page":     limit,
			},
		},
	})
}

func GetEventDetail(c *gin.Context) {
	// Try to get identifier from 'slug' or 'id' param name
	identifier := c.Param("slug")
	if identifier == "" {
		identifier = c.Param("id")
	}

	var event models.Event

	// 1. Try find by Slug
	if err := config.DB.Preload("Category").Preload("TicketTypes").Where("slug = ?", identifier).First(&event).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": event})
		return
	}

	// 2. Fallback: If identifier is numeric, try finding by ID
	if id, errConv := strconv.Atoi(identifier); errConv == nil {
		if err := config.DB.Preload("Category").Preload("TicketTypes").Where("id = ?", id).First(&event).Error; err == nil {
			c.JSON(http.StatusOK, gin.H{"success": true, "data": event})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Event not found"})
}

// GetUpcomingEvents
func GetUpcomingEvents(c *gin.Context) {
	events := []models.Event{}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if err := config.DB.Model(&models.Event{}).
		Preload("Category").Preload("TicketTypes").
		Where("status = ? AND event_date >= ?", "published", time.Now()).
		Order("event_date ASC").
		Limit(limit).
		Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to fetch upcoming events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": events})
}

// GetPopularEvents
func GetPopularEvents(c *gin.Context) {
	events := []models.Event{}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	// Logic: Most tickets sold (status 'active' or 'used')
	// Using left join to include events with 0 sales if needed (though popular implies sales)
	if err := config.DB.Table("events").
		Select("events.*, COUNT(tickets.id) as sales_count").
		Joins("LEFT JOIN tickets ON tickets.event_id = events.id AND tickets.status IN (?)", []string{"active", "used"}).
		Where("events.status = ?", "published").
		Group("events.id").
		Order("sales_count DESC").
		Limit(limit).
		Preload("Category").Preload("TicketTypes").
		Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to fetch popular events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": events})
}

// GetFeaturedEvents
func GetFeaturedEvents(c *gin.Context) {
	events := []models.Event{}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if err := config.DB.Model(&models.Event{}).
		Preload("Category").Preload("TicketTypes").
		Where("status = ? AND is_featured = ?", "published", true).
		Order("event_date ASC").
		Limit(limit).
		Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to fetch featured events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": events})
}

// GetCities
func GetCities(c *gin.Context) {
	cities := []string{}
	// Get distinct cities from published events
	if err := config.DB.Model(&models.Event{}).
		Where("status = ?", "published").
		Distinct("city").
		Pluck("city", &cities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to fetch cities"})
		return
	}

	// Format as list of objects with count could be nicer
	type CityResponse struct {
		Name  string `json:"name"`
		Count int64  `json:"count"`
	}
	response := []CityResponse{}
	for _, cityName := range cities {
		var count int64
		config.DB.Model(&models.Event{}).Where("status = ? AND city = ?", "published", cityName).Count(&count)
		response = append(response, CityResponse{Name: cityName, Count: count})
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": response})
}
