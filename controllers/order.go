package controllers

import (
	"encoding/json"
	"fmt"
	"log"
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
	VoucherCode   string `json:"voucher_code"` // Added for discount
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
	ticketPrices := make(map[uint]float64)

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

		// --- CHECK MAX PURCHASE PER USER ---
		if ticketType.MaxPurchasePerUser > 0 {
			var alreadyPurchased int64
			// Join with orders to check user's previous non-cancelled purchases
			tx.Model(&models.Ticket{}).
				Joins("JOIN orders ON orders.id = tickets.order_id").
				Where("tickets.ticket_type_id = ? AND orders.status != ?", ticketType.ID, "cancelled").
				Where("(orders.user_id = ? OR orders.customer_email = ?)", userID, customerEmail).
				Count(&alreadyPurchased)

			if int(alreadyPurchased)+item.Quantity > ticketType.MaxPurchasePerUser {
				tx.Rollback()
				remaining := ticketType.MaxPurchasePerUser - int(alreadyPurchased)
				if remaining <= 0 {
					c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": fmt.Sprintf("Maaf, Anda sudah mencapai batas maksimal pembelian untuk tiket '%s'.", ticketType.Name)})
				} else {
					c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": fmt.Sprintf("Maaf, Anda hanya bisa membeli %d tiket lagi untuk '%s'.", remaining, ticketType.Name)})
				}
				return
			}
		}

		// --- FLASH SALE MODULE ---
		// Determine active price: default to normal ticket price
		activePrice := ticketType.Price
		isFlashSaleContext := false

		var activeFlashSales []models.FlashSale
		var flashSale *models.FlashSale

		// Load Timezone WIB (Asia/Jakarta) agar sinkron dengan database
		loc, _ := time.LoadLocation("Asia/Jakarta")
		now := time.Now().In(loc)

		errFlash := tx.Where("ticket_type_id = ? AND is_active = true", ticketType.ID).Find(&activeFlashSales).Error
		if errFlash == nil && len(activeFlashSales) > 0 {
			ny, nm, nd := now.Date()
			currentTimeStr := now.Format("15:04")

			// Temukan 1 Flash Sale yang sedang aktif secara waktu persis saat ini
			// Temukan Harga Termurah dari seluruh Flash Sale yang sedang aktif saat ini
			for i := range activeFlashSales {
				fs := activeFlashSales[i]
				if fs.FlashDate != nil {
					sy, sm, sd := fs.FlashDate.Date()
					if sy == ny && sm == nm && sd == nd {
						if fs.StartTime != "" && fs.EndTime != "" {
							if currentTimeStr >= fs.StartTime && currentTimeStr < fs.EndTime {
								// Pilih yang harganya paling murah jika ada jadwal bentrok
								if flashSale == nil || fs.FlashPrice < flashSale.FlashPrice {
									flashSale = &activeFlashSales[i]
								}
							}
						}
					}
				}
			}
		}

		if flashSale != nil {
			// Flash sale is active, check quota
			availableFlashQuota := flashSale.Quota - flashSale.Sold
			if availableFlashQuota >= item.Quantity {
				isFlashSaleContext = true
				activePrice = flashSale.FlashPrice
			} else if availableFlashQuota > 0 {
				// Edge case: Partially available flash sale
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": fmt.Sprintf("Kuota Flash Sale sisa %d, mengurangi pesanan Anda.", availableFlashQuota)})
				return
			}
		}

		if !isFlashSaleContext {
			// Normal Ticketing Validation (Non-Flash Sale)
			if ticketType.Available < item.Quantity {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": fmt.Sprintf("Not enough quota for %s", ticketType.Name)})
				return
			}

			// Deduct normal quota atomically
			res := tx.Model(&ticketType).
				Where("available >= ?", item.Quantity).
				Update("available", gorm.Expr("available - ?", item.Quantity))

			if err := res.Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to update quota (oversold check)"})
				return
			}
			if res.RowsAffected == 0 {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": fmt.Sprintf("Mohon maaf, tiket '%s' baru saja habis terjual atau kuota tidak cukup.", ticketType.Name)})
				return
			}
		} else {
			// Deduct flash sale quota
			resFlash := tx.Model(flashSale).
				Where("quota - sold >= ?", item.Quantity). // Extra safe check
				Update("sold", gorm.Expr("sold + ?", item.Quantity))

			if err := resFlash.Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to update flash sale quota"})
				return
			}
			if resFlash.RowsAffected == 0 {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Mohon maaf, kuota Flash Sale sudah habis atau tidak mencukupi."})
				return
			}

			// Optionally: still deduct event overall quota if needed?
			// Usually Flash Sale Quota is a subset of Total Quota. Let's deduct both.
			resMaster := tx.Model(&ticketType).
				Where("available >= ?", item.Quantity).
				Update("available", gorm.Expr("available - ?", item.Quantity))

			if err := resMaster.Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to update master quota"})
				return
			}
			if resMaster.RowsAffected == 0 {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": fmt.Sprintf("Mohon maaf, tiket utama '%s' sudah habis.", ticketType.Name)})
				return
			}
		}

		// Refresh from DB to get the new 'available' value for the rest of the logic
		tx.First(&ticketType, ticketType.ID)

		ticketPrices[ticketType.ID] = activePrice
		itemSubtotal := activePrice * float64(item.Quantity)
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

			var flashID *uint
			if isFlashSaleContext && flashSale != nil {
				id := flashSale.ID
				flashID = &id
			}

			orderItems = append(orderItems, models.Ticket{
				EventID:              ticketType.EventID,
				TicketTypeID:         ticketType.ID,
				TicketCode:           fmt.Sprintf("T-%d-%d-%d", time.Now().UnixNano(), ticketType.ID, i),
				AttendeeName:         attendeeName,
				AttendeeEmail:        attendeeEmail,
				AttendeePhone:        attendeePhone,
				PurchasedPrice:       activePrice,
				FlashSaleID:          flashID,
				CustomFieldResponses: customResponses,
				Status:               "active",
			})
		}
	}

	// Voucher Processing
	var discountAmount float64
	var appliedVoucherCode string

	if req.VoucherCode != "" {
		var voucher models.Voucher
		// Check validity: is_active, expires_at tracking
		if err := tx.Where("code = ? AND is_active = ?", req.VoucherCode, true).First(&voucher).Error; err == nil {
			valid := true
			if voucher.ExpiresAt != nil && voucher.ExpiresAt.Before(time.Now()) {
				valid = false
			}
			if voucher.MaxUses > 0 && voucher.UsedCount >= voucher.MaxUses {
				valid = false
			}

			// NEW: Prevent multiple usage by same user/email
			if valid {
				var count int64
				uq := tx.Model(&models.Order{}).Where("voucher_code = ? AND status != ?", req.VoucherCode, "cancelled")
				if userID != nil {
					uq = uq.Where("(user_id = ? OR customer_email = ?)", *userID, customerEmail)
				} else {
					uq = uq.Where("customer_email = ?", customerEmail)
				}
				uq.Count(&count)
				if count > 0 {
					valid = false
					tx.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Anda sudah pernah menggunakan voucher ini sebelumnya"})
					return
				}
			}

			// If event_id is specified, make sure it matches the current items
			// For simplicity we check if the voucher applies to AT LEAST ONE item
			if valid && (voucher.EventID != nil || voucher.TicketTypeID != nil) {
				eventMatch := false
				ticketTypeMatch := false
				for _, item := range orderItems {
					if voucher.EventID != nil && item.EventID == *voucher.EventID {
						eventMatch = true
					}
					if voucher.TicketTypeID != nil && item.TicketTypeID == *voucher.TicketTypeID {
						ticketTypeMatch = true
					}
				}

				if voucher.EventID != nil && !eventMatch {
					valid = false
				}
				if voucher.TicketTypeID != nil && !ticketTypeMatch {
					valid = false
				}
			}

			if valid {
				// Calculate eligible amount
				var eligibleAmount float64
				for _, item := range orderItems {
					isEligible := true
					if voucher.EventID != nil && item.EventID != *voucher.EventID {
						isEligible = false
					}
					if voucher.TicketTypeID != nil && item.TicketTypeID != *voucher.TicketTypeID {
						isEligible = false
					}
					if isEligible {
						eligibleAmount += ticketPrices[item.TicketTypeID]
					}
				}

				// Calculate discount against eligible amount only
				if voucher.DiscountType == "percent" {
					calc := eligibleAmount * (voucher.DiscountValue / 100)
					if voucher.MaxDiscountAmount != nil && calc > *voucher.MaxDiscountAmount {
						discountAmount = *voucher.MaxDiscountAmount
					} else {
						discountAmount = calc
					}
				} else if voucher.DiscountType == "fixed" {
					discountAmount = voucher.DiscountValue // fixed discount applies globally but bounded by eligibleAmount? Or per eligibleAmount? Let's just limit by eligibleAmount
				}

				if discountAmount > eligibleAmount {
					discountAmount = eligibleAmount // Never discount more than eligible price
				}

				appliedVoucherCode = voucher.Code

				// Increment used_count
				tx.Model(&voucher).Update("used_count", gorm.Expr("used_count + ?", 1))
			} else {
				// Don't fail the whole order, but ideally return error so user knows.
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Voucher tidak valid atau sudah kadaluarsa"})
				return
			}
		} else {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Kode voucher tidak ditemukan"})
			return
		}
	}

	// Unique code generation removed (System now uses Flip)
	var uniqueCode int

	// Create Order
	order := models.Order{
		UserID:         userID,
		OrderNumber:    fmt.Sprintf("ORD-%d", time.Now().Unix()),
		CustomerName:   customerName,
		CustomerEmail:  customerEmail,
		CustomerPhone:  customerPhone,
		TotalAmount:    totalAmount + totalAdminFee - discountAmount + float64(uniqueCode), // Add fee, unique code, subtract discount
		AdminFee:       totalAdminFee,
		DiscountAmount: discountAmount,
		VoucherCode:    appliedVoucherCode,
		UniqueCode:     uniqueCode,
		Status:         "pending",
		PaymentMethod:  req.PaymentMethod,
		CreatedAt:      time.Now(),
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
		} else {
			if order.UserID != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Please login to view this order"})
				return
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
		} else {
			if order.UserID != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Please login to view these tickets"})
				return
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
	// Flip Callback detect
	// Flip sends 'data' as string/JSON or 'data' field in JSON
	// Also check X-Callback-Token header
	callbackToken := os.Getenv("FLIP_WEBHOOK_TOKEN")
	flipToken := c.GetHeader("X-Callback-Token")

	// 1. Try Flip Payload
	var flipData struct {
		Data  string `json:"data"`
		Token string `json:"token"`
	}

	if err := c.ShouldBindJSON(&flipData); err == nil && (flipData.Token == callbackToken || flipToken == callbackToken) {
		// Proceed as Flip Call
		var bill struct {
			ExternalID string `json:"external_id"`
			Status     string `json:"status"` // SUCCESSFUL, CANCELLED
		}
		if err := json.Unmarshal([]byte(flipData.Data), &bill); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid flip data"})
			return
		}

		status := "pending"
		if bill.Status == "SUCCESSFUL" {
			status = "success"
		} else if bill.Status == "CANCELLED" {
			status = "failed"
		}

		processOrderPayment(bill.ExternalID, status, c)
		return
	}

	// 2. Original Mock Callback (Fallback)
	var input struct {
		OrderNumber string `json:"order_number"`
		Status      string `json:"status"` // success, failed
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid input"})
		return
	}
	processOrderPayment(input.OrderNumber, input.Status, c)
}

// Extracted internal function to process payment status
func processOrderPayment(orderNumber string, status string, c *gin.Context) {
	var order models.Order
	if err := config.DB.Where("order_number = ?", orderNumber).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Order not found"})
		return
	}

	tx := config.DB.Begin()

	if status == "success" || status == "SUCCESSFUL" || status == "paid" {
		now := time.Now()
		if err := tx.Model(&order).Updates(models.Order{
			Status: "paid",
			PaidAt: &now,
		}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to update order status"})
			return
		}
	} else if status == "failed" || status == "CANCELLED" || status == "expired" {
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
		Status:    status,
		Notes:     "Callback received: " + status,
		CreatedAt: time.Now(),
	})

	tx.Commit()

	// If status is paid, send email (Triggered outside transaction for performance)
	if status == "success" || status == "SUCCESSFUL" || status == "paid" {
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
// processPaymentGateway abstracts the payment generation logic.
func processPaymentGateway(order *models.Order, paymentMethod string, userID *uint) {
	if os.Getenv("FLIP_API_KEY") != "" {
		// Use Flip Bill
		redirectURL := fmt.Sprintf("%s/payment/%s", os.Getenv("FRONTEND_URL"), order.OrderNumber)
		resp, err := utils.CreateFlipBill(order.OrderNumber, int(order.TotalAmount), order.CustomerName, order.CustomerEmail, order.CustomerPhone, redirectURL)
		if err == nil {
			order.PaymentURL = resp.PaymentURL
			order.PaymentData = fmt.Sprintf("Flip Link ID: %d, Bill ID: %d", resp.ID, resp.BillID)
			order.PaymentInstructions = "Silakan klik link pembayaran Flip untuk menyelesaikan transaksi."
			return
		}
		// Fallback to mock if API fails? Or log error?
		log.Printf("Flip API Error for order %s: %v", order.OrderNumber, err)
	}

	// 1. Virtual Accounts (Mock Fallback)
	if strings.Contains(paymentMethod, "VA") {
		// ... existing mock VA logic ... (shortened for clarity if it's the same)
		bankCode := "88888"
		if strings.Contains(paymentMethod, "BCA") {
			bankCode = "70012"
		} else if strings.Contains(paymentMethod, "Mandiri") {
			bankCode = "88888"
		} else if strings.Contains(paymentMethod, "BNI") {
			bankCode = "88881"
		} else if strings.Contains(paymentMethod, "BRI") {
			bankCode = "88882"
		}

		idForVA := 0
		if userID != nil {
			idForVA = int(*userID)
		} else {
			idForVA = 99999
		}

		userIDPadded := fmt.Sprintf("%05d", idForVA)
		timestamp := fmt.Sprintf("%06d", time.Now().Unix()%1000000)
		order.VirtualAccountNumber = fmt.Sprintf("%s%s%s", bankCode, userIDPadded, timestamp)
		return
	}

	// 2. QRIS/E-Wallet (Mock Fallback)
	if strings.Contains(paymentMethod, "QRIS") || strings.Contains(paymentMethod, "OVO") || strings.Contains(paymentMethod, "Dana") {
		order.PaymentURL = fmt.Sprintf("https://simulator.kartcis.id/pay/%s", order.OrderNumber)
		return
	}

	// 4. Retail Outlet (Mock Fallback)
	if strings.Contains(paymentMethod, "Alfamart") || strings.Contains(paymentMethod, "Indomaret") {
		order.VirtualAccountNumber = fmt.Sprintf("ALFA-%d", time.Now().UnixNano()%100000000)
		return
	}
}
