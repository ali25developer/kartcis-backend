package controllers

import (
	"fmt"
	"kartcis-backend/config"
	"kartcis-backend/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GetMyTickets
func GetMyTickets(c *gin.Context) {
	usrID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Unauthorized"})
		return
	}
	// Convert "userID" from context (usually float64 from JWT claims) to uint
	// Convert "userID" from context
	var userID uint
	if idUint, ok := usrID.(uint); ok {
		userID = idUint
	} else if idFloat, ok := usrID.(float64); ok {
		userID = uint(idFloat)
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Invalid user session"})
		return
	}

	allTickets := []models.Ticket{}
	// Preload necessary associations
	// Note: Check if "Order" table name is "orders". GORM default is plural.
	if err := config.DB.
		Joins("JOIN orders ON orders.id = tickets.order_id").
		Joins("JOIN events ON events.id = tickets.event_id").
		Where("orders.user_id = ?", userID).
		Order("orders.paid_at DESC NULLS LAST").
		Preload("Order").
		Preload("Event").
		Preload("TicketType").
		Find(&allTickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to fetch tickets", "error": err.Error()})
		return
	}

	upcoming := []models.Ticket{}
	past := []models.Ticket{}
	now := time.Now()

	for _, t := range allTickets {
		if t.Event.EventDate.After(now) || t.Event.EventDate.Equal(now) {
			upcoming = append(upcoming, t)
		} else {
			past = append(past, t)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"upcoming": upcoming,
			"past":     past,
		},
	})
}

// GetTicketDetail (Public/Guest)
func GetTicketDetail(c *gin.Context) {
	ticketCode := c.Param("code")
	var ticket models.Ticket

	if err := config.DB.Where("ticket_code = ?", ticketCode).Preload("Event").Preload("TicketType").Preload("Order").First(&ticket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Ticket not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": ticket})
}

// VerifyTicket
func VerifyTicket(c *gin.Context) {
	ticketCode := c.Param("code")
	var ticket models.Ticket

	if err := config.DB.Where("ticket_code = ?", ticketCode).Preload("Event").Preload("TicketType").First(&ticket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Ticket not found"})
		return
	}

	isValid := ticket.Status == "active"

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"ticket_code":   ticket.TicketCode,
			"status":        ticket.Status,
			"is_valid":      isValid,
			"attendee_name": ticket.AttendeeName,
			"event_title":   ticket.Event.Title, // Need to ensure Event is preloaded
			"event_date":    ticket.Event.EventDate,
			"ticket_type":   ticket.TicketType.Name,
			"check_in_at":   ticket.CheckInAt,
		},
	})
}

// DownloadTicketPDF
func DownloadTicketPDF(c *gin.Context) {
	ticketCode := c.Param("code")
	var ticket models.Ticket
	if err := config.DB.Where("ticket_code = ?", ticketCode).Preload("Event").Preload("TicketType").First(&ticket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Ticket not found"})
		return
	}

	// Generate HTML Receipt/Ticket
	htmlContent := fmt.Sprintf(`
		<html>
		<head><title>Ticket %s</title></head>
		<body style="font-family: Arial, sans-serif; padding: 20px; border: 2px dashed #333; max-width: 600px; margin: 0 auto;">
			<h1 style="text-align: center;">%s</h1>
			<p style="text-align: center;">%s</p>
			<hr/>
			<div style="margin: 20px 0;">
				<p><strong>Ticket Code:</strong> %s</p>
				<p><strong>Attendee:</strong> %s</p>
				<p><strong>Type:</strong> %s</p>
				<p><strong>Date:</strong> %s</p>
			</div>
			<div style="text-align: center; margin-top: 30px;">
				<p><em>Scan this QR Code at the venue</em></p>
				<div style="background: #eee; width: 150px; height: 150px; margin: 0 auto; display: flex; align-items: center; justify-content: center;">
					[QR CODE PLACEHOLDER]
				</div>
			</div>
		</body>
		</html>
	`, ticket.TicketCode, ticket.Event.Title, ticket.Event.Venue, ticket.TicketCode, ticket.AttendeeName, ticket.TicketType.Name, ticket.Event.EventDate.Format("Mon, 02 Jan 2006"))

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=ticket-%s.html", ticketCode))
	c.Data(http.StatusOK, "text/html", []byte(htmlContent))
}
