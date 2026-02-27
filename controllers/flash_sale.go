package controllers

import (
	"kartcis-backend/config"
	"kartcis-backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type FlashSaleRequest struct {
	EventID      uint    `json:"event_id" binding:"required"`
	TicketTypeID uint    `json:"ticket_type_id" binding:"required"`
	FlashPrice   float64 `json:"flash_price" binding:"required"`
	Quota        int     `json:"quota" binding:"required"`
	FlashDate    string  `json:"flash_date"` // YYYY-MM-DD
	StartTime    string  `json:"start_time"` // HH:MM
	EndTime      string  `json:"end_time"`   // HH:MM
	IsActive     *bool   `json:"is_active"`
}

func CreateFlashSale(c *gin.Context) {
	var req FlashSaleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid input"})
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	flashSale := models.FlashSale{
		EventID:      req.EventID,
		TicketTypeID: req.TicketTypeID,
		FlashPrice:   req.FlashPrice,
		Quota:        req.Quota,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		IsActive:     isActive,
	}

	if req.FlashDate != "" {
		fd, err := parseEventDate(req.FlashDate)
		if err == nil {
			flashSale.FlashDate = &fd
		}
	}

	if err := config.DB.Create(&flashSale).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to create flash sale: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": flashSale})
}

func GetFlashSales(c *gin.Context) {
	eventId := c.Query("event_id")
	var sales []models.FlashSale

	query := config.DB.Preload("TicketType")
	if eventId != "" {
		query = query.Where("event_id = ?", eventId)
	}

	query.Find(&sales)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": sales})
}

func UpdateFlashSale(c *gin.Context) {
	id := c.Param("id")
	var flashSale models.FlashSale

	if err := config.DB.First(&flashSale, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Flash sale not found"})
		return
	}

	var req FlashSaleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid input"})
		return
	}

	updates := map[string]interface{}{
		"flash_price": req.FlashPrice,
		"quota":       req.Quota,
		"start_time":  req.StartTime,
		"end_time":    req.EndTime,
	}

	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.FlashDate != "" {
		fd, _ := parseEventDate(req.FlashDate)
		updates["flash_date"] = fd
	}

	config.DB.Model(&flashSale).Updates(updates)

	c.JSON(http.StatusOK, gin.H{"success": true, "data": flashSale})
}

func DeleteFlashSale(c *gin.Context) {
	id := c.Param("id")
	config.DB.Delete(&models.FlashSale{}, id)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Flash sale deleted"})
}
