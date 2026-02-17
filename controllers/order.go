package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"kartcis-backend/config"
	"kartcis-backend/models"
	"kartcis-backend/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type CheckoutRequest struct {
	Items []struct {
		TicketTypeID uint `json:"ticket_type_id"`
		Quantity     int  `json:"quantity"`
		Attendees    []struct {
			Name                 string      `json:"name"`
			Email                string      `json:"email"`
			Phone                string      `json:"phone"`
			CustomFieldResponses interface{} `json:"custom_field_responses"` // Allow object or string
		} `json:"attendees"`
	} `json:"items"`
	PaymentMethod string `json:"payment_method"`
	// Guest Info (Optional if logged in)
	CustomerInfo struct {
		Name  string `json:"name"`
		Email string `json:"email"`
		Phone string `json:"phone"`
	} `json:"customer_info"`
}

func CreateOrder(c *gin.Context) {
	usrID, exists := c.Get("userID") // Using OptionalAuthMiddleware

	var req CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	tx := config.DB.Begin()

	var totalAmount float64
	var orderItems []models.Ticket

	// Determine Customer Info
	var customerName, customerEmail, customerPhone string
	var userID *uint

	if exists {
		// Logged in user
		id := usrID.(uint)
		userID = &id
		var user models.User
		if err := tx.First(&user, id).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "User not found"})
			return
		}
		customerName = user.Name
		customerEmail = user.Email
		customerPhone = user.Phone
	} else {
		// Guest
		if req.CustomerInfo.Name == "" || req.CustomerInfo.Email == "" || req.CustomerInfo.Phone == "" {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Customer info required for guest checkout"})
			return
		}
		customerName = req.CustomerInfo.Name
		customerEmail = req.CustomerInfo.Email
		customerPhone = req.CustomerInfo.Phone
		userID = nil
	}

	var totalAdminFee float64

	for _, item := range req.Items {
		var ticketType models.TicketType
		// Preload Event to get FeePercentage
		if err := tx.Preload("Event").First(&ticketType, item.TicketTypeID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid ticket type"})
			return
		}

		if ticketType.Event.Status == "cancelled" {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Maaf, event ini telah dibatalkan."})
			return
		}

		if ticketType.Event.Status == "completed" {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Maaf, event ini sudah selesai."})
			return
		}

		if ticketType.Event.Status == "sold_out" {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Maaf, tiket untuk event ini sudah habis terjual."})
			return
		}

		if ticketType.Event.Status != "published" {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Maaf, event ini tidak tersedia saat ini."})
			return
		}

		if item.Quantity <= 0 {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Quantity must be at least 1"})
			return
		}

		if ticketType.Available < item.Quantity {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": fmt.Sprintf("Not enough quota for %s", ticketType.Name)})
			return
		}

		// Deduct quota atomically
		if err := tx.Model(&ticketType).
			Where("available >= ?", item.Quantity).
			Update("available", gorm.Expr("available - ?", item.Quantity)).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to update quota (oversold check)"})
			return
		}

		// Refresh from DB to get the new 'available' value for the rest of the logic
		tx.First(&ticketType, ticketType.ID)

		itemSubtotal := ticketType.Price * float64(item.Quantity)
		totalAmount += itemSubtotal

		// Calculate Admin Fee for this item based on Event settings
		// Default 5.0 if 0 (though DB default handles new records, existing might need care)
		// But migrated schema has default 5.0.
		// Calculate fee: subtotal * (percentage / 100)
		fee := itemSubtotal * (ticketType.Event.FeePercentage / 100)
		totalAdminFee += fee

		// Create tickets
		for i := 0; i < item.Quantity; i++ {
			attendeeName := customerName
			attendeeEmail := customerEmail
			attendeePhone := customerPhone
			customResponses := ""

			// If specific attendee info provided for this ticket index
			if i < len(item.Attendees) {
				if item.Attendees[i].Name != "" {
					attendeeName = item.Attendees[i].Name
				}
				if item.Attendees[i].Email != "" {
					attendeeEmail = item.Attendees[i].Email
				}
				if item.Attendees[i].Phone != "" {
					attendeePhone = item.Attendees[i].Phone
				}

				// Handle CustomFieldResponses (could be string or object)
				if item.Attendees[i].CustomFieldResponses != nil {
					switch v := item.Attendees[i].CustomFieldResponses.(type) {
					case string:
						customResponses = v
					default:
						// Marshal object to JSON string
						b, _ := json.Marshal(v)
						customResponses = string(b)
					}
				}
			}

			orderItems = append(orderItems, models.Ticket{
				EventID:              ticketType.EventID,
				TicketTypeID:         ticketType.ID,
				TicketCode:           fmt.Sprintf("T-%d-%d-%d", time.Now().UnixNano(), ticketType.ID, i),
				AttendeeName:         attendeeName,
				AttendeeEmail:        attendeeEmail,
				AttendeePhone:        attendeePhone,
				CustomFieldResponses: customResponses,
				Status:               "active",
			})
		}
	}

	// Handle Unique Code for Manual Bank Transfer
	var uniqueCode int
	if strings.HasPrefix(req.PaymentMethod, "BANK_TRANSFER_") || req.PaymentMethod == "MANUAL_JAGO" {
		// Generate unique code and ensure TOTAL AMOUNT is unique for pending orders
		baseAmount := totalAmount + totalAdminFee

		// 1. Get ALL currently used codes for this amount
		var usedCodes []int
		// We query just the UniqueCode column for pending orders where (TotalAmount - UniqueCode) is approx BaseAmount
		// To be safe and DB agnostic about float precision logic, we can query by range or just select unique_code
		// where status='pending' AND ABS(total_amount - unique_code - baseAmount) < 1.0

		tx.Model(&models.Order{}).
			Where("status = ? AND total_amount >= ? AND total_amount <= ?", "pending", baseAmount+101, baseAmount+999).
			Pluck("unique_code", &usedCodes)

		// 2. Check Capacity
		if len(usedCodes) >= 899 {
			tx.Rollback()
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"message": "Maaf, sistem pembayaran untuk nominal ini sedang sangat penuh. Mohon coba 15-30 menit lagi.",
			})
			return
		}

		// 3. Find First Available Slot (Linear Search)
		// Map for O(1) lookup
		usedMap := make(map[int]bool)
		for _, code := range usedCodes {
			usedMap[code] = true
		}

		// Search 101 to 999
		for code := 101; code <= 999; code++ {
			if !usedMap[code] {
				uniqueCode = code
				break
			}
		}
	}

	// Create Order
	order := models.Order{
		UserID:        userID,
		OrderNumber:   fmt.Sprintf("ORD-%d", time.Now().Unix()),
		CustomerName:  customerName,
		CustomerEmail: customerEmail,
		CustomerPhone: customerPhone,
		TotalAmount:   totalAmount + totalAdminFee + float64(uniqueCode), // Add fee and unique code to total
		AdminFee:      totalAdminFee,
		UniqueCode:    uniqueCode,
		Status:        "pending",
		PaymentMethod: req.PaymentMethod,
		CreatedAt:     time.Now(),
	}

	// Process Payment (Simulate Gateway)
	// This function prepares the order for payment (generating VA, URLs, etc)
	processPaymentGateway(&order, req.PaymentMethod, userID)

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to create order"})
		return
	}

	// Save tickets linked to order
	for i := range orderItems {
		orderItems[i].OrderID = &order.ID
		if err := tx.Create(&orderItems[i]).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to generate tickets"})
			return
		}
	}

	tx.Commit()

	// Record history
	config.DB.Create(&models.OrderStatusHistory{
		OrderID:   order.ID,
		Status:    "pending",
		Notes:     "Order created",
		CreatedAt: time.Now(),
	})

	// Send Payment Instruction Email
	utils.SendPaymentInstructionEmail(order)

	// Record history: Email Sent
	config.DB.Create(&models.OrderStatusHistory{
		OrderID:   order.ID,
		Status:    "pending",
		Notes:     "Payment instruction email sent to customer",
		CreatedAt: time.Now(),
	})

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    order,
	})
}

func GetUserOrders(c *gin.Context) {
	userID, _ := c.Get("userID")
	orders := []models.Order{}
	var totalItems int64

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	query := config.DB.Where("user_id = ?", userID)

	// Count Total
	query.Model(&models.Order{}).Count(&totalItems)

	// Fetch Data
	query.Preload("Tickets").Order("created_at desc").Limit(limit).Offset(offset).Find(&orders)

	totalPages := int(totalItems) / limit
	if int(totalItems)%limit != 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"orders": orders,
			"pagination": gin.H{
				"current_page": page,
				"total_pages":  totalPages,
				"total_items":  totalItems,
				"per_page":     limit,
			},
		},
	})
}

func GetOrderDetail(c *gin.Context) {
	param := c.Param("order_number")
	userID, loggedIn := c.Get("userID")
	userRole, _ := c.Get("userRole")

	var order models.Order

	// 1. Try find by order_number (SAFE for Guest)
	if err := config.DB.Preload("Tickets.Event").Preload("Tickets.TicketType").Where("order_number = ?", param).First(&order).Error; err == nil {
		if loggedIn {
			isAdmin := userRole == "admin"
			isOwner := order.UserID != nil && *order.UserID == userID.(uint)
			if !isAdmin && !isOwner {
				if order.UserID != nil {
					c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "You are not authorized to view this order"})
					return
				}
			}
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "data": order})
		return
	}

	// 2. Try find by ID (RESTRICTED - Auth Only)
	if id, err := strconv.Atoi(param); err == nil {
		if !loggedIn {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Login required to access order by ID"})
			return
		}

		query := config.DB.Preload("Tickets.Event").Preload("Tickets.TicketType").Where("id = ?", id)
		if userRole != "admin" {
			query = query.Where("user_id = ?", userID)
		}

		if err := query.First(&order).Error; err == nil {
			c.JSON(http.StatusOK, gin.H{"success": true, "data": order})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Order not found"})
}

func GetOrderTickets(c *gin.Context) {
	param := c.Param("order_number")
	userID, loggedIn := c.Get("userID")
	userRole, _ := c.Get("userRole")

	var order models.Order

	// 1. Try find by order_number (SAFE for Guest)
	// We only need tickets here
	if err := config.DB.Preload("Tickets.Event").Preload("Tickets.TicketType").Where("order_number = ?", param).First(&order).Error; err == nil {
		if loggedIn {
			isAdmin := userRole == "admin"
			isOwner := order.UserID != nil && *order.UserID == userID.(uint)
			if !isAdmin && !isOwner {
				if order.UserID != nil {
					c.JSON(http.StatusForbidden, gin.H{"success": false, "message": "You are not authorized to view these tickets"})
					return
				}
			}
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "data": order.Tickets})
		return
	}

	// 2. Try find by ID (RESTRICTED - Auth Only)
	if id, err := strconv.Atoi(param); err == nil {
		if !loggedIn {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Login required to access tickets by ID"})
			return
		}

		query := config.DB.Preload("Tickets.Event").Preload("Tickets.TicketType").Where("id = ?", id)
		if userRole != "admin" {
			query = query.Where("user_id = ?", userID)
		}

		if err := query.First(&order).Error; err == nil {
			c.JSON(http.StatusOK, gin.H{"success": true, "data": order.Tickets}) // Return TICKETS only
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Order not found"})
}

func PayOrder(c *gin.Context) {
	// Manual payment confirmation endpoint.
	// In a real app, this might be called by an admin or triggered by manual transfer confirmation.
	param := c.Param("order_number")
	userID, _ := c.Get("userID")

	var order models.Order
	// Try find by order_number or ID
	query := config.DB.Where("user_id = ?", userID)
	if id, err := strconv.Atoi(param); err == nil {
		query = query.Where("id = ? OR order_number = ?", id, param)
	} else {
		query = query.Where("order_number = ?", param)
	}

	if err := query.First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Order not found"})
		return
	}

	// Simulate Success
	now := time.Now()
	config.DB.Model(&order).Updates(models.Order{
		Status: "paid",
		PaidAt: &now,
	})

	// Also Generate Tickets QR/Code if NOT generated at checkout?
	// Current logic generated them at checkout as "active". If payment fails they should probably be "pending" or cancelled.
	// Spec says: "Create order (checkout)" -> "Simulate payment".
	// Better logic: Tickets status = inactive/pending until paid.
	// For now, keeping MVP.

	c.JSON(http.StatusOK, gin.H{"success": true, "data": order})
}

// Payment Callback (Webhook)
func PaymentCallback(c *gin.Context) {
	// In real world, verify signature from Payment Gateway (Midtrans, Xendit, etc)
	var input struct {
		OrderNumber string `json:"order_number"`
		Status      string `json:"status"` // success, failed
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid input"})
		return
	}

	var order models.Order
	if err := config.DB.Where("order_number = ?", input.OrderNumber).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Order not found"})
		return
	}

	tx := config.DB.Begin()

	if input.Status == "success" {
		now := time.Now()
		if err := tx.Model(&order).Updates(models.Order{
			Status: "paid",
			PaidAt: &now,
		}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to update order status"})
			return
		}
	} else if input.Status == "failed" {
		if err := tx.Model(&order).Updates(models.Order{
			Status: "cancelled",
		}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to update order status"})
			return
		}
		if err := utils.RestoreQuota(tx, order.ID); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to restore quota"})
			return
		}
	}

	// Record history
	tx.Create(&models.OrderStatusHistory{
		OrderID:   order.ID,
		Status:    input.Status,
		Notes:     "Callback received: " + input.Status,
		CreatedAt: time.Now(),
	})

	tx.Commit()

	// If status is paid, send email (Triggered outside transaction for performance)
	if input.Status == "success" {
		var tickets []models.Ticket
		config.DB.Preload("Event").Preload("TicketType").Where("order_id = ?", order.ID).Find(&tickets)
		utils.SendTicketEmail(order, tickets)

		// Record history: E-Ticket Sent
		config.DB.Create(&models.OrderStatusHistory{
			OrderID:   order.ID,
			Status:    "paid",
			Notes:     "E-Ticket email sent to customer",
			CreatedAt: time.Now(),
		})
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Callback processed"})
}

func UserCancelOrder(c *gin.Context) {
	orderNumber := c.Param("order_number")
	var order models.Order

	if err := config.DB.Preload("Tickets").Where("order_number = ?", orderNumber).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Order not found"})
		return
	}

	// Policy: User can only cancel if status is "pending"
	if order.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": fmt.Sprintf("Cannot cancel order because it is already %s", order.Status),
		})
		return
	}

	tx := config.DB.Begin()

	if err := tx.Model(&order).Update("status", "cancelled").Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to cancel order", "error": err.Error()})
		return
	}

	// Restore Quota
	if err := utils.RestoreQuota(tx, order.ID); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to restore quota"})
		return
	}

	// Record history
	config.DB.Create(&models.OrderStatusHistory{
		OrderID:   order.ID,
		Status:    "cancelled",
		Notes:     "Cancelled by user",
		CreatedAt: time.Now(),
	})

	tx.Commit()

	// Send Cancellation Email
	utils.SendOrderCancelledEmail(order, "Dibatalkan oleh pengguna")

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Order cancelled successfully", "data": order})
}

// processPaymentGateway abstracts the payment generation logic.
// In the future, replace the body of this function with actual API calls to Flip/Midtrans.
func processPaymentGateway(order *models.Order, paymentMethod string, userID *uint) {
	// 1. Virtual Accounts (BCA, Mandiri, BNI, BRI, etc.)
	if strings.Contains(paymentMethod, "VA") {
		// MOCK: Generate local VA Number
		bankCode := "88888" // Default
		if strings.Contains(paymentMethod, "BCA") {
			bankCode = "70012"
		} else if strings.Contains(paymentMethod, "Mandiri") {
			bankCode = "88888"
		} else if strings.Contains(paymentMethod, "BNI") {
			bankCode = "88881"
		} else if strings.Contains(paymentMethod, "BRI") {
			bankCode = "88882" // Example
		}

		idForVA := 0
		if userID != nil {
			idForVA = int(*userID)
		} else {
			idForVA = 99999 // Guest
		}

		userIDPadded := fmt.Sprintf("%05d", idForVA)
		timestamp := fmt.Sprintf("%06d", time.Now().Unix()%1000000)

		order.VirtualAccountNumber = fmt.Sprintf("%s%s%s", bankCode, userIDPadded, timestamp)

		// In Real Gateway:
		// resp, _ := flip.CreateBill(amount, bankCode, ...)
		// order.VirtualAccountNumber = resp.VANumber
		// order.PaymentData = resp.RawJSON
		return
	}

	// 2. E-Wallets / QRIS (OVO, Dana, ShopeePay, LinkAja)
	if strings.Contains(paymentMethod, "QRIS") || strings.Contains(paymentMethod, "OVO") || strings.Contains(paymentMethod, "Dana") {
		// MOCK: Generate a Deep Link or QR String
		// For simulation, we point to a dummy payment page or return a static QR string

		// order.PaymentURL = "https://app.sandbox.midtrans.com/..."
		// order.PaymentURL = "https://flip.id/pw/..."

		// Let's just create a dummy link for now
		order.PaymentURL = fmt.Sprintf("https://simulator.kartcis.id/pay/%s", order.OrderNumber)
		return
	}
	// 3. Retail Outlet (Alfamart/Indomaret)
	if strings.Contains(paymentMethod, "Alfamart") || strings.Contains(paymentMethod, "Indomaret") {
		// MOCK: Generate Payment Code
		order.VirtualAccountNumber = fmt.Sprintf("ALFA-%d", time.Now().UnixNano()%100000000)
		return
	}

	// 4. Bank Transfer (Jago)
	if strings.Contains(paymentMethod, "BANK_TRANSFER_JAGO") || paymentMethod == "MANUAL_JAGO" {
		// Set Payment Instruction
		accNo := os.Getenv("JAGO_ACCOUNT_NUMBER")
		accName := os.Getenv("JAGO_ACCOUNT_NAME")
		if accNo == "" {
			accNo = "1010101020" // Default Demo
			accName = "Kartcis Demo Account"
		}

		order.VirtualAccountNumber = accNo
		order.PaymentData = accName
		order.PaymentInstructions = fmt.Sprintf("Silakan transfer ke Bank Jago: %s a/n %s. Pastikan nominal sampai 3 digit terakhir (Rp %v) agar dapat diverifikasi otomatis.", accNo, accName, utils.FormatPrice(order.TotalAmount))
		return
	}
}
