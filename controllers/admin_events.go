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
func generateSlug(title string) string {
	slug := strings.ToLower(title)
	re := regexp.MustCompile(`[^a-z0-9]+`)
	slug = re.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug
}

// AdminGetEvents retrieves events with admin-specific filters
func AdminGetEvents(c *gin.Context) {
	events := []models.Event{}
	var totalItems int64

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	query := config.DB.Model(&models.Event{})
	status := c.Query("status")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	search := c.Query("search")
	if search != "" {
		query = query.Where("title ILIKE ? OR venue ILIKE ? OR city ILIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// Count Total
	query.Count(&totalItems)

	// Fetch Data
	query.Preload("Category").Preload("TicketTypes").Order("updated_at desc").Limit(limit).Offset(offset).Find(&events)

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

type EventRequest struct {
	Title               string               `json:"title"`
	Slug                string               `json:"slug"`
	Description         string               `json:"description"`
	DetailedDescription string               `json:"detailed_description"`
	EventDate           string               `json:"event_date"` // Change to string for flexible parsing
	EventTime           string               `json:"event_time"`
	Venue               string               `json:"venue"`
	City                string               `json:"city"`
	Organizer           string               `json:"organizer"`
	Image               string               `json:"image"`
	Quota               int                  `json:"quota"`
	IsFeatured          *bool                `json:"is_featured"` // Use pointer for boolean
	Status              string               `json:"status"`
	CategoryID          uint                 `json:"category_id"`
	MinPrice            float64              `json:"min_price"`
	MaxPrice            float64              `json:"max_price"`
	FeePercentage       float64              `json:"fee_percentage"`
	CustomFields        string               `json:"custom_fields"`
	TicketTypes         *[]models.TicketType `json:"ticket_types"` // Use pointer to distinguish between nil (omitted) and [] (empty)
}

func parseEventDate(dateStr string) (time.Time, error) {
	// Try ISO 8601 first
	t, err := time.Parse(time.RFC3339, dateStr)
	if err == nil {
		return t, nil
	}
	// Try YYYY-MM-DD
	t, err = time.Parse("2006-01-02", dateStr)
	if err == nil {
		return t, nil
	}
	return time.Time{}, err
}

func CreateEvent(c *gin.Context) {
	var req EventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid input", "error": err.Error()})
		return
	}

	// Parsing Date
	eventDate, err := parseEventDate(req.EventDate)
	if err != nil && req.EventDate != "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid date format. Use YYYY-MM-DD or ISO8601"})
		return
	}

	// Validation
	if req.Title == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"success": false, "message": "Title is required"})
		return
	}

	if req.CategoryID == 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"success": false, "message": "Category ID is required"})
		return
	}

	// Check if category exists
	var category models.Category
	if err := config.DB.First(&category, req.CategoryID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid Category ID"})
		return
	}

	isFeatured := false
	if req.IsFeatured != nil {
		isFeatured = *req.IsFeatured
	}

	ticketTypes := []models.TicketType{}
	if req.TicketTypes != nil {
		ticketTypes = *req.TicketTypes
	}

	// Calculate automatic fields from ticket types if available
	var minPrice, maxPrice float64
	var totalQuota int
	if len(ticketTypes) > 0 {
		minPrice = ticketTypes[0].Price
		maxPrice = ticketTypes[0].Price
		for _, tt := range ticketTypes {
			if tt.Price < minPrice {
				minPrice = tt.Price
			}
			if tt.Price > maxPrice {
				maxPrice = tt.Price
			}
			totalQuota += tt.Quota
		}
	} else {
		// Use manual values if no tickets provided
		minPrice = req.MinPrice
		maxPrice = req.MaxPrice
		totalQuota = req.Quota
	}

	// Map Request to Model
	input := models.Event{
		Title:               req.Title,
		Slug:                req.Slug,
		Description:         req.Description,
		DetailedDescription: req.DetailedDescription,
		EventDate:           eventDate,
		EventTime:           req.EventTime,
		Venue:               req.Venue,
		City:                req.City,
		Organizer:           req.Organizer,
		Image:               req.Image,
		Quota:               totalQuota, // Auto calculated
		IsFeatured:          isFeatured,
		Status:              req.Status,
		CategoryID:          req.CategoryID,
		MinPrice:            minPrice, // Auto calculated
		MaxPrice:            maxPrice, // Auto calculated
		FeePercentage:       req.FeePercentage,
		CustomFields:        req.CustomFields,
		TicketTypes:         ticketTypes,
	}

	// Basic defaults
	if input.Slug == "" {
		input.Slug = generateSlug(input.Title)
	}

	if input.Status == "" {
		input.Status = "draft"
	}

	// Default Fee Percentage Logic
	if input.FeePercentage == 0 {
		input.FeePercentage = 5.0
	}

	tx := config.DB.Begin()

	// 1. Create Event
	if err := tx.Create(&input).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to create event", "error": err.Error()})
		return
	}

	// 2. Handle Ticket Types if provided
	for i := range input.TicketTypes {
		input.TicketTypes[i].EventID = input.ID
		if input.TicketTypes[i].Available == 0 && input.TicketTypes[i].Quota > 0 {
			input.TicketTypes[i].Available = input.TicketTypes[i].Quota
		}
	}

	if len(input.TicketTypes) > 0 {
		if err := tx.Save(&input.TicketTypes).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to create ticket types"})
			return
		}
	}

	tx.Commit()

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": input})
}

func UpdateEvent(c *gin.Context) {
	id := c.Param("id")
	var event models.Event

	if err := config.DB.First(&event, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Event not found"})
		return
	}

	var req EventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid input", "error": err.Error()})
		return
	}

	// Use transaction for consistency
	tx := config.DB.Begin()

	// Map Request to Updates (ONLY if not empty/zero)
	updates := make(map[string]interface{})
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Slug != "" {
		updates["slug"] = req.Slug
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.DetailedDescription != "" {
		updates["detailed_description"] = req.DetailedDescription
	}
	if req.EventTime != "" {
		updates["event_time"] = req.EventTime
	}
	if req.Venue != "" {
		updates["venue"] = req.Venue
	}
	if req.City != "" {
		updates["city"] = req.City
	}
	if req.Organizer != "" {
		updates["organizer"] = req.Organizer
	}
	if req.Image != "" {
		updates["image"] = req.Image
	}
	if req.CategoryID != 0 {
		updates["category_id"] = req.CategoryID
	}
	if req.FeePercentage != 0 {
		updates["fee_percentage"] = req.FeePercentage
	}
	if req.CustomFields != "" {
		updates["custom_fields"] = req.CustomFields
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.IsFeatured != nil {
		updates["is_featured"] = *req.IsFeatured
	}

	// Parsing Date if provided
	if req.EventDate != "" {
		eventDate, err := parseEventDate(req.EventDate)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid date format"})
			return
		}
		updates["event_date"] = eventDate
	}

	// Update basic fields
	if err := tx.Model(&event).Updates(updates).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to update event"})
		return
	}

	// Handle Ticket Types Sync if provided
	if req.TicketTypes != nil {
		newTTs := *req.TicketTypes

		// AUTO CALCULATE from Ticket Types
		var minPrice, maxPrice float64
		var totalQuota int
		if len(newTTs) > 0 {
			minPrice = newTTs[0].Price
			maxPrice = newTTs[0].Price
			for _, tt := range newTTs {
				if tt.Price < minPrice {
					minPrice = tt.Price
				}
				if tt.Price > maxPrice {
					maxPrice = tt.Price
				}
				totalQuota += tt.Quota
			}
			updates["min_price"] = minPrice
			updates["max_price"] = maxPrice
			updates["quota"] = totalQuota
		}

		providedIDs := make(map[uint]bool)
		for _, tt := range newTTs {
			if tt.ID != 0 {
				providedIDs[tt.ID] = true
			}
		}

		// 1. Delete ticket types NOT in the provided list
		var existingTTs []models.TicketType
		tx.Where("event_id = ?", event.ID).Find(&existingTTs)

		for _, exTT := range existingTTs {
			if !providedIDs[exTT.ID] {
				// Safety check: is it sold?
				var count int64
				tx.Model(&models.Ticket{}).Where("ticket_type_id = ?", exTT.ID).Count(&count)
				if count > 0 {
					tx.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"success": false,
						"message": "Cannot delete ticket type '" + exTT.Name + "' because it has already been sold.",
					})
					return
				}
				if err := tx.Delete(&exTT).Error; err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to delete old ticket type"})
					return
				}
			}
		}

		// 2. Update existing or Create new
		for _, tt := range newTTs {
			tt.EventID = event.ID
			if tt.ID == 0 {
				// New
				tt.Available = tt.Quota
				if err := tx.Create(&tt).Error; err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to create new ticket type"})
					return
				}
			} else {
				// Update existing: Adjust 'Available' based on change in 'Quota'
				var oldTT models.TicketType
				if err := tx.First(&oldTT, tt.ID).Error; err == nil {
					diff := tt.Quota - oldTT.Quota
					if diff != 0 {
						tt.Available = oldTT.Available + diff
						if tt.Available < 0 {
							tt.Available = 0
						}
					} else {
						// Keep current availability if quota didn't change
						tt.Available = oldTT.Available
					}

					// Safety Check: Never exceed Quota
					if tt.Available > tt.Quota {
						tt.Available = tt.Quota
					}
				}

				if err := tx.Model(&models.TicketType{}).Where("id = ? AND event_id = ?", tt.ID, event.ID).
					Select("Name", "Description", "Price", "OriginalPrice", "Quota", "Available").
					Updates(tt).Error; err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to update ticket type"})
					return
				}
			}
		}

		// 3. Final Sync for Event Totals (MinPrice, MaxPrice, Total Quota)
		if err := tx.Model(&event).Updates(map[string]interface{}{
			"min_price": updates["min_price"],
			"max_price": updates["max_price"],
			"quota":     updates["quota"],
		}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to update event totals"})
			return
		}
	}

	tx.Commit()

	// Load updated event with associations for response
	config.DB.Preload("Category").Preload("TicketTypes").First(&event, id)

	c.JSON(http.StatusOK, gin.H{"success": true, "data": event})
}

func DeleteEvent(c *gin.Context) {
	id := c.Param("id")

	// Check if any tickets have been sold
	var count int64
	config.DB.Model(&models.Ticket{}).Where("event_id = ? AND status IN ?", id, []string{"active", "used"}).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Cannot delete event because tickets have already been sold. Please cancel the event instead.",
		})
		return
	}

	if err := config.DB.Delete(&models.Event{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to delete"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Event deleted"})
}

// UpdateEventStatus
func UpdateEventStatus(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Status is required"})
		return
	}

	var event models.Event
	if err := config.DB.First(&event, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Event not found"})
		return
	}

	event.Status = input.Status
	config.DB.Save(&event)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Event status updated"})
}

// GetEventAnalytics
func GetEventAnalytics(c *gin.Context) {
	id := c.Param("id")
	var event models.Event
	if err := config.DB.First(&event, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Event not found"})
		return
	}

	var totalTickets int64
	var soldTickets int64
	var revenue float64

	// Count tickets associated with this event
	config.DB.Model(&models.Ticket{}).Where("event_id = ?", id).Count(&totalTickets)

	// Sold tickets
	config.DB.Model(&models.Ticket{}).Where("event_id = ? AND status IN ?", id, []string{"active", "used"}).Count(&soldTickets)

	// Calculate revenue
	type Result struct {
		Total float64
	}
	var result Result
	config.DB.Table("tickets").
		Joins("JOIN ticket_types ON tickets.ticket_type_id = ticket_types.id").
		Where("tickets.event_id = ? AND tickets.status IN ?", id, []string{"active", "used"}).
		Select("SUM(ticket_types.price) as total").
		Scan(&result)

	revenue = result.Total

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"event_title": event.Title,
			"quota":       event.Quota,
			"sold":        soldTickets,
			"revenue":     revenue,
			"views":       0, // Analytics feature pending integration
		},
	})
}
