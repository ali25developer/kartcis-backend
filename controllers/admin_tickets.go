package controllers

import (
	"kartcis-backend/config"
	"kartcis-backend/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// CheckInTicket
func CheckInTicket(c *gin.Context) {
	type CheckInInput struct {
		TicketCode string `json:"ticket_code" binding:"required"`
	}

	var input CheckInInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Ticket code is required"})
		return
	}

	var ticket models.Ticket
	// Preload details for response
	if err := config.DB.Where("ticket_code = ?", input.TicketCode).Preload("Event").Preload("TicketType").First(&ticket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Ticket not found"})
		return
	}

	if ticket.Status == "used" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Ticket already used",
			"data": gin.H{
				"check_in_at": ticket.CheckInAt,
			},
		})
		return
	}

	// Mark as used
	now := time.Now()
	ticket.Status = "used"
	ticket.CheckInAt = &now
	ticket.UpdatedAt = now

	config.DB.Save(&ticket)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Check-in successful",
		"data": gin.H{
			"ticket_code":   ticket.TicketCode,
			"attendee_name": ticket.AttendeeName,
			"event_title":   ticket.Event.Title,
			"ticket_type":   ticket.TicketType.Name,
			"check_in_at":   ticket.CheckInAt,
		},
	})
}
