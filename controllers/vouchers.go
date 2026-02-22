package controllers

import (
	"kartcis-backend/config"
	"kartcis-backend/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ValidateVoucher checks if a voucher code is valid for a given event and amount
func ValidateVoucher(c *gin.Context) {
	code := c.Query("code")
	eventIDStr := c.Query("event_id")
	ticketTypeIDStr := c.Query("ticket_type_id")

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

	if eventIDStr != "" && voucher.EventID != nil {
		eventID, _ := strconv.Atoi(eventIDStr)
		if int(*voucher.EventID) != eventID {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Voucher tidak berlaku untuk event ini"})
			return
		}
	}

	if ticketTypeIDStr != "" && voucher.TicketTypeID != nil {
		ticketTypeID, _ := strconv.Atoi(ticketTypeIDStr)
		if int(*voucher.TicketTypeID) != ticketTypeID {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Voucher tidak berlaku untuk tipe tiket ini"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Voucher valid",
		"data": gin.H{
			"code":                voucher.Code,
			"discount_type":       voucher.DiscountType,
			"discount_value":      voucher.DiscountValue,
			"max_discount_amount": voucher.MaxDiscountAmount,
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
