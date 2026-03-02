package controllers

import (
	"kartcis-backend/config"
	"kartcis-backend/models"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ValidateVoucher checks if a voucher code is valid for a given event and amount
func ValidateVoucher(c *gin.Context) {
	code := c.Query("code")
	// Use QueryArray to support multiple IDs (?event_id=1&event_id=2)
	eventIDs := c.QueryArray("event_id")
	ticketTypeIDs := c.QueryArray("ticket_type_id")

	// Fallback to comma-separated string check if QueryArray is empty but param exists
	if len(eventIDs) == 0 && c.Query("event_id") != "" {
		eventIDs = strings.Split(c.Query("event_id"), ",")
	}
	if len(ticketTypeIDs) == 0 && c.Query("ticket_type_id") != "" {
		ticketTypeIDs = strings.Split(c.Query("ticket_type_id"), ",")
	}

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Kode voucher tidak boleh kosong"})
		return
	}

	var voucher models.Voucher
	if err := config.DB.Where("code = ? AND is_active = ?", code, true).First(&voucher).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Voucher tidak ditemukan atau tidak aktif"})
		return
	}

	if voucher.ExpiresAt != nil && voucher.ExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Voucher telah kadaluarsa"})
		return
	}

	if voucher.MaxUses > 0 && voucher.UsedCount >= voucher.MaxUses {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Batas penggunaan voucher telah habis"})
		return
	}

	// NEW: Check if this user/email has already used this voucher
	email := c.Query("email")
	userIDStr := c.Query("user_id")

	query := config.DB.Where("voucher_code = ? AND status != ?", code, "cancelled")

	if userIDStr != "" {
		userID, _ := strconv.Atoi(userIDStr)
		query = query.Where("(user_id = ? OR customer_email = ?)", userID, email)
	} else if email != "" {
		query = query.Where("customer_email = ?", email)
	}

	if email != "" || userIDStr != "" {
		var count int64
		query.Model(&models.Order{}).Count(&count)
		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Anda sudah pernah menggunakan voucher ini sebelumnya"})
			return
		}
	}

	// Check Event Compatibility
	if len(eventIDs) > 0 && voucher.EventID != nil {
		match := false
		for _, idStr := range eventIDs {
			id, _ := strconv.Atoi(idStr)
			if int(*voucher.EventID) == id {
				match = true
				break
			}
		}
		if !match {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Voucher tidak berlaku untuk event ini"})
			return
		}
	}

	// Check Ticket Type Compatibility
	if len(ticketTypeIDs) > 0 && voucher.TicketTypeID != nil {
		match := false
		for _, idStr := range ticketTypeIDs {
			id, _ := strconv.Atoi(idStr)
			if int(*voucher.TicketTypeID) == id {
				match = true
				break
			}
		}
		if !match {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Voucher tidak berlaku untuk tipe tiket ini"})
			return
		}
	}

	// Calculate affected items for frontend UI feedback
	var affectedTicketTypeIDs []int
	isGlobal := voucher.EventID == nil && voucher.TicketTypeID == nil

	if isGlobal {
		// If global, all provided ticket types are affected
		for _, idStr := range ticketTypeIDs {
			id, _ := strconv.Atoi(idStr)
			affectedTicketTypeIDs = append(affectedTicketTypeIDs, id)
		}
	} else if voucher.TicketTypeID != nil {
		// If specific ticket type, only that one is affected
		affectedTicketTypeIDs = append(affectedTicketTypeIDs, int(*voucher.TicketTypeID))
	} else if voucher.EventID != nil {
		// If event-wide, we need to know which ticket IDs belong to this event.
		// Since we don't want to query the DB for all ticket types here for performance,
		// we assume the frontend sends the correct event_id for each ticket_type_id.
		// We'll return the voucher's event_id so the frontend can filter its cart.
		// For now, we'll just return the event_id and let the frontend handle filtering.
		// If the frontend sends ticket_type_ids, we can assume they are for the specified event.
		// So, all provided ticket_type_ids are affected if they belong to the voucher's event.
		// However, without querying the DB for ticket types, we can't verify this here.
		// A simpler approach is to return the event_id and let the frontend filter its cart items.
		// Or, if ticketTypeIDs are provided, assume they are valid for the event and return them.
		for _, idStr := range ticketTypeIDs {
			id, _ := strconv.Atoi(idStr)
			affectedTicketTypeIDs = append(affectedTicketTypeIDs, id)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Voucher valid",
		"data": gin.H{
			"code":                     voucher.Code,
			"discount_type":            voucher.DiscountType,
			"discount_value":           voucher.DiscountValue,
			"max_discount_amount":      voucher.MaxDiscountAmount,
			"eligible_event_id":        voucher.EventID,
			"eligible_ticket_type_id":  voucher.TicketTypeID,
			"affected_ticket_type_ids": affectedTicketTypeIDs,
			"is_global":                isGlobal,
		},
	})
}

// --- Admin CRUD Vouchers ---

func AdminGetVouchers(c *gin.Context) {
	var vouchers []models.Voucher
	var totalItems int64

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	query := config.DB.Model(&models.Voucher{})

	// Search
	search := c.Query("search")
	if search != "" {
		query = query.Where("code ILIKE ?", "%"+search+"%")
	}

	// Filter by event
	eventID := c.Query("event_id")
	if eventID != "" {
		query = query.Where("event_id = ?", eventID)
	}

	// Count Total
	query.Count(&totalItems)

	// Fetch Data
	if err := query.Preload("Event").Order("created_at desc").Limit(limit).Offset(offset).Find(&vouchers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to fetch vouchers"})
		return
	}

	totalPages := int(totalItems) / limit
	if int(totalItems)%limit != 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"vouchers": vouchers,
			"pagination": gin.H{
				"current_page": page,
				"total_pages":  totalPages,
				"total_items":  totalItems,
				"per_page":     limit,
			},
		},
	})
}

func CreateVoucher(c *gin.Context) {
	var input models.Voucher
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid input"})
		return
	}

	if err := config.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Gagal membuat voucher (Kode mungkin sudah ada)", "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Voucher created successfully", "data": input})
}

func GetVoucherDetail(c *gin.Context) {
	id := c.Param("id")
	var voucher models.Voucher

	if err := config.DB.Preload("Event").First(&voucher, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Voucher not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": voucher})
}

func UpdateVoucher(c *gin.Context) {
	id := c.Param("id")
	var voucher models.Voucher

	if err := config.DB.First(&voucher, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Voucher not found"})
		return
	}

	var input models.Voucher
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid input"})
		return
	}

	// Overwrite fields
	voucher.Code = input.Code
	voucher.DiscountType = input.DiscountType
	voucher.DiscountValue = input.DiscountValue
	voucher.MaxDiscountAmount = input.MaxDiscountAmount
	voucher.MaxUses = input.MaxUses
	voucher.EventID = input.EventID
	voucher.TicketTypeID = input.TicketTypeID
	voucher.ExpiresAt = input.ExpiresAt
	voucher.IsActive = input.IsActive
	voucher.UpdatedAt = time.Now()

	if err := config.DB.Save(&voucher).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Failed to update voucher", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Voucher updated successfully", "data": voucher})
}

func DeleteVoucher(c *gin.Context) {
	id := c.Param("id")
	var voucher models.Voucher

	if err := config.DB.First(&voucher, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Voucher not found"})
		return
	}

	config.DB.Delete(&voucher)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Voucher deleted successfully"})
}

func UpdateVoucherStatus(c *gin.Context) {
	id := c.Param("id")
	var voucher models.Voucher

	if err := config.DB.First(&voucher, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Voucher not found"})
		return
	}

	type StatusInput struct {
		IsActive bool `json:"is_active"`
	}
	var input StatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Invalid input"})
		return
	}

	voucher.IsActive = input.IsActive
	voucher.UpdatedAt = time.Now()
	config.DB.Save(&voucher)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Voucher status updated", "data": voucher})
}
