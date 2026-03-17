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
	VoucherCode   string `json:"voucher_code"`  // Added for voucher discount
	ReferralCode  string `json:"referral_code"` // Added for referral/affiliate
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

	// Referral Code Processing
	var referralDiscount float64
	var appliedReferralCode string

	if req.ReferralCode != "" {
		rCode := strings.ToUpper(strings.TrimSpace(req.ReferralCode))
		var referral models.ReferralCode
		if err := tx.Where("code = ? AND is_active = true", rCode).First(&referral).Error; err == nil {
			rValid := true
			if referral.ExpiresAt != nil && referral.ExpiresAt.Before(time.Now()) {
				rValid = false
			}
			if referral.MaxUses > 0 && referral.UsedCount >= referral.MaxUses {
				rValid = false
			}
			// Event scope: at least one ticket must belong to referral's event
			if rValid && referral.EventID != nil {
				eventFound := false
				for _, t := range orderItems {
					if t.EventID == *referral.EventID {
						eventFound = true
						break
					}
				}
				if !eventFound {
					rValid = false
				}
			}

			if rValid {
				// Discount opsional — hanya jika discount_type bukan "none"
				if referral.DiscountType == "percent" && referral.DiscountValue > 0 {
					referralDiscount = totalAmount * (referral.DiscountValue / 100)
				} else if referral.DiscountType == "fixed" && referral.DiscountValue > 0 {
					referralDiscount = referral.DiscountValue
					if referralDiscount > totalAmount {
						referralDiscount = totalAmount
					}
				}
				appliedReferralCode = referral.Code
				// Increment used_count untuk tracking
				tx.Model(&referral).Update("used_count", gorm.Expr("used_count + ?", 1))
			}
		}
	}

	// Handle Unique Code for Manual Bank Transfer
	var uniqueCode int
	if strings.HasPrefix(req.PaymentMethod, "BANK_TRANSFER_") || req.PaymentMethod == "MANUAL_JAGO" {
		// Generate unique code and ensure TOTAL AMOUNT is unique for pending orders
		baseAmount := totalAmount + totalAdminFee - discountAmount - referralDiscount

		// 1. Get ALL codes used in the last 3 hours
		var usedCodes []int
		threeHoursAgo := time.Now().Add(-3 * time.Hour)
		tx.Model(&models.Order{}).
			Where("created_at >= ? AND total_amount >= ? AND total_amount <= ?", threeHoursAgo, baseAmount+101, baseAmount+999).
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
		usedMap := make(map[int]bool)
		for _, code := range usedCodes {
			usedMap[code] = true
		}
		for code := 101; code <= 999; code++ {
			if !usedMap[code] {
				uniqueCode = code
				break
			}
		}
	}

	// Create Order
	order := models.Order{
		UserID:         userID,
		OrderNumber:    fmt.Sprintf("ORD-%d", time.Now().Unix()),
		CustomerName:   customerName,
		CustomerEmail:  customerEmail,
		CustomerPhone:  customerPhone,
		TotalAmount:    totalAmount + totalAdminFee - discountAmount - referralDiscount + float64(uniqueCode),
		AdminFee:       totalAdminFee,
		DiscountAmount: discountAmount + referralDiscount, // Combined discount
		VoucherCode:    appliedVoucherCode,
		ReferralCode:   appliedReferralCode,
		UniqueCode:     uniqueCode,
		Status:         "pending",
		PaymentMethod:  req.PaymentMethod,
		CreatedAt:      time.Now(),
	}

	// Process Payment (Generate VA, URLs, payment instructions)
	log.Printf("[Order] Processing payment gateway for method: %s", req.PaymentMethod)
	if err := processPaymentGateway(&order, req.PaymentMethod, userID); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Flip API Error: " + err.Error()})
		return
	}

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

// Payment Callback (Webhook dari Flip)
func PaymentCallback(c *gin.Context) {
	callbackToken := os.Getenv("FLIP_WEBHOOK_TOKEN")

	// Flip mengirim callback sebagai form-encoded:
	// data  = JSON string berisi detail bill
	// token = validation token (form field TERPISAH, bukan di dalam JSON)
	dataStr := c.PostForm("data")
	tokenStr := c.PostForm("token") // ← token di sini, bukan di dalam JSON data

	if dataStr == "" {
		// Log for debugging payload
		log.Printf("[Flip-Callback] Empty data received")
		c.Status(http.StatusBadRequest)
		return
	}

	// Validasi token (dari form field "token")
	if callbackToken != "" && tokenStr != callbackToken {
		log.Printf("[Flip-Callback] Invalid token received: %s", tokenStr)
		c.Status(http.StatusUnauthorized)
		return
	}

	var bill struct {
		BillLinkID int64  `json:"bill_link_id"` // int64! Format berubah jadi 19 digit per April 10, 2026
		Status     string `json:"status"`
	}
	if err := json.Unmarshal([]byte(dataStr), &bill); err != nil {
		log.Printf("[Flip-Callback] Failed to parse data: %v", err)
		c.Status(http.StatusBadRequest)
		return
	}

	log.Printf("[Flip-Callback] Received: bill_link_id=%d, status=%s", bill.BillLinkID, bill.Status)

	// Lookup order by payment_data yang menyimpan bill_link_id
	var order models.Order

	// Format terbaru ("146927")
	exactMatch := fmt.Sprintf("%d", bill.BillLinkID)

	if err := config.DB.Where("payment_data = ?", exactMatch).
		First(&order).Error; err != nil {
		// Return 200 agar Flip tidak retry terus
		log.Printf("[Flip-Callback] Order not found for bill_link_id: %d", bill.BillLinkID)
		c.Status(http.StatusOK)
		return
	}

	status := "pending"
	switch bill.Status {
	case "SUCCESSFUL":
		status = "paid"
	case "CANCELLED", "FAILED":
		status = "cancelled"
	}

	if status == "pending" {
		c.Status(http.StatusOK)
		return
	}

	processOrderPayment(order.OrderNumber, status, c)
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
	} else if status == "cancelled" || status == "failed" || status == "CANCELLED" || status == "expired" {
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

		// Restore used counts for vouchers/referrals and cancel commission
		if err := restoreVoucherAndReferral(tx, order); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to restore vouchers and referrals"})
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

	// Restore Vouchers and Referrals
	if err := restoreVoucherAndReferral(tx, order); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to restore vouchers and referrals"})
		return
	}

	// Record history
	tx.Create(&models.OrderStatusHistory{
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

// Helper to restore limits on cancellation
func restoreVoucherAndReferral(tx *gorm.DB, order models.Order) error {
	if order.VoucherCode != "" {
		if err := tx.Model(&models.Voucher{}).Where("code = ?", order.VoucherCode).
			Update("used_count", gorm.Expr("used_count - ?", 1)).Error; err != nil {
			return err
		}
	}
	if order.ReferralCode != "" {
		if err := tx.Model(&models.ReferralCode{}).Where("code = ?", order.ReferralCode).
			Update("used_count", gorm.Expr("used_count - ?", 1)).Error; err != nil {
			return err
		}
	}
	return nil
}

// processPaymentGateway abstracts the payment generation logic.
func processPaymentGateway(order *models.Order, paymentMethod string, userID *uint) error {
	pm := strings.ToUpper(strings.TrimSpace(paymentMethod))
	if !strings.HasPrefix(pm, "BANK_TRANSFER_") && pm != "MANUAL_JAGO" {
		if os.Getenv("FLIP_API_KEY") != "" {
			resp, err := utils.CreateFlipBill(order.OrderNumber, int(order.TotalAmount), order.CustomerName, order.CustomerEmail, order.CustomerPhone, os.Getenv("FRONTEND_URL")+"/orders/"+order.OrderNumber, nil)
			if err == nil {
				order.PaymentURL = resp.PaymentURL
				order.PaymentData = fmt.Sprintf("%d", resp.ID)
				return nil
			}
			return err
		}
	}

	/*
		// 1. Virtual Accounts (BCA, Mandiri, BNI, BRI, etc.)
		if strings.Contains(pm, "VA") {
			bankCode := "88888" // Default
			if strings.Contains(pm, "BCA") {
				bankCode = "70012"
			} else if strings.Contains(pm, "MANDIRI") {
				bankCode = "88888"
			} else if strings.Contains(pm, "BNI") {
				bankCode = "88881"
			} else if strings.Contains(pm, "BRI") {
				bankCode = "88882"
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
			return nil
		}

		// 2. E-Wallets / QRIS (OVO, Dana, ShopeePay, LinkAja)
		if strings.Contains(pm, "QRIS") || strings.Contains(pm, "OVO") || strings.Contains(pm, "DANA") {
			order.PaymentURL = fmt.Sprintf("https://simulator.kartcis.id/pay/%s", order.OrderNumber)
			return nil
		}

		// 3. Retail Outlet (Alfamart/Indomaret)
		if strings.Contains(pm, "ALFAMART") || strings.Contains(pm, "INDOMARET") {
			order.VirtualAccountNumber = fmt.Sprintf("ALFA-%d", time.Now().UnixNano()%100000000)
			return nil
		}

		// 4. Bank Transfer Manual (Jago) — diverifikasi otomatis via email scraping
		if strings.HasPrefix(pm, "BANK_TRANSFER_") || pm == "MANUAL_JAGO" {
			accNo := os.Getenv("JAGO_ACCOUNT_NUMBER")
			accName := os.Getenv("JAGO_ACCOUNT_NAME")
			if accNo == "" {
				accNo = "1010101020" // Default Demo
				accName = "Kartcis Demo Account"
			}
			order.VirtualAccountNumber = accNo
			order.PaymentData = accName
			order.PaymentInstructions = fmt.Sprintf(
				"Silakan transfer ke Bank Jago: %s a/n %s. Pastikan nominal sampai 3 digit terakhir (Rp %v) agar dapat diverifikasi otomatis.",
				accNo, accName, utils.FormatPrice(order.TotalAmount),
			)
			return nil
		}
	*/
	return nil
}
