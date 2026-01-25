package controllers

import (
	"kartcis-backend/models"
	"kartcis-backend/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func TestEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		email = "ali25developer@gmail.com"
	}

	// Create a dummy order and ticket for testing
	now := time.Now()
	order := models.Order{
		OrderNumber:  "TEST-ORDER-123",
		CustomerName: "Test User",
		CreatedAt:    now,
	}

	event := models.Event{
		Title:     "Konser Musik Digital",
		EventDate: now.AddDate(0, 0, 7),
		EventTime: "19:00",
		Venue:     "Istora Senayan",
		City:      "Jakarta",
		Image:     "https://images.unsplash.com/photo-1501281668745-f7f57925c3b4?q=80&w=2070&auto=format&fit=crop",
	}

	ticketType := models.TicketType{
		Name: "VIP Gold",
	}

	ticket := models.Ticket{
		TicketCode:           "KRT-TEST-999",
		AttendeeName:         "Ahmat Ali",
		AttendeeEmail:        email,
		Event:                event,
		TicketType:           ticketType,
		CustomFieldResponses: `{"No. Identitas": "1234567890", "Ukuran Kaos": "XL"}`,
	}

	// Trigger Email
	utils.SendTicketEmail(order, []models.Ticket{ticket})

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Test email has been queued to " + email,
		"details": "Check your terminal logs for delivery status.",
	})
}

func TestPaymentEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		email = "ali25developer@gmail.com"
	}

	now := time.Now()
	order := models.Order{
		OrderNumber:          "TEST-PAY-456",
		CustomerName:         "Test Buyer",
		CustomerEmail:        email,
		TotalAmount:          150500,
		PaymentMethod:        "BCA VA",
		VirtualAccountNumber: "7001299999123456",
		CreatedAt:            now,
	}

	// Trigger Payment Email
	utils.SendPaymentInstructionEmail(order)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Test payment instruction email has been queued to " + email,
	})
}
