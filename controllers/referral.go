package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"

	"kartcis-backend/config"
	"kartcis-backend/models"

	"github.com/gin-gonic/gin"
)

// ─────────────────────────────────────────────
// Helper
// ─────────────────────────────────────────────

func generateReferralCode(prefix string) string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	suffix := strings.ToUpper(hex.EncodeToString(bytes))
	if prefix != "" {
		return strings.ToUpper(prefix) + "-" + suffix
	}
	return "REF-" + suffix
}

// ─────────────────────────────────────────────
// ADMIN: CRUD Referral Codes
// ─────────────────────────────────────────────

// GET /admin/referrals
func AdminGetReferralCodes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	var total int64
	config.DB.Model(&models.ReferralCode{}).Count(&total)

	var codes []models.ReferralCode
	config.DB.Preload("User").Preload("Event").Order("created_at DESC").Limit(limit).Offset(offset).Find(&codes)

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"referral_codes": codes,
			"pagination": gin.H{
				"current_page": page,
				"total_pages":  totalPages,
				"total_items":  total,
				"per_page":     limit,
			},
		},
	})
}

// POST /admin/referrals
func CreateReferralCode(c *gin.Context) {
	var input struct {
		PartnerName   string     `json:"partner_name" binding:"required"`
		UserID        *uint      `json:"user_id"`
		EventID       *uint      `json:"event_id"`
		CustomCode    string     `json:"custom_code"`
		Prefix        string     `json:"prefix"`
		DiscountType  string     `json:"discount_type"` // none, percent, fixed
		DiscountValue float64    `json:"discount_value"`
		RewardType    string     `json:"reward_type"` // none, percent, fixed
		RewardValue   float64    `json:"reward_value"`
		MaxUses       int        `json:"max_uses"`
		ExpiresAt     *time.Time `json:"expires_at"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	if input.DiscountType == "" {
		input.DiscountType = "none"
	}
	if input.RewardType == "" {
		input.RewardType = "none"
	}

	code := input.CustomCode
	if code == "" {
		code = generateReferralCode(input.Prefix)
	} else {
		code = strings.ToUpper(strings.TrimSpace(code))
	}

	referral := models.ReferralCode{
		Code:          code,
		PartnerName:   input.PartnerName,
		UserID:        input.UserID,
		EventID:       input.EventID,
		DiscountType:  input.DiscountType,
		DiscountValue: input.DiscountValue,
		RewardType:    input.RewardType,
		RewardValue:   input.RewardValue,
		MaxUses:       input.MaxUses,
		ExpiresAt:     input.ExpiresAt,
		IsActive:      true,
	}

	if err := config.DB.Create(&referral).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Gagal membuat referral code: " + err.Error()})
		return
	}

	config.DB.Preload("User").Preload("Event").First(&referral, referral.ID)
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": referral})
}

// GET /admin/referrals/:id
func GetReferralCodeDetail(c *gin.Context) {
	id := c.Param("id")

	var code models.ReferralCode
	if err := config.DB.Preload("User").Preload("Event").First(&code, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Referral code not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": code})
}

// PUT /admin/referrals/:id
func UpdateReferralCode(c *gin.Context) {
	id := c.Param("id")

	var code models.ReferralCode
	if err := config.DB.First(&code, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Referral code not found"})
		return
	}

	var input struct {
		PartnerName   string     `json:"partner_name"`
		EventID       *uint      `json:"event_id"`
		DiscountType  string     `json:"discount_type"`
		DiscountValue *float64   `json:"discount_value"`
		RewardType    string     `json:"reward_type"`
		RewardValue   *float64   `json:"reward_value"`
		MaxUses       *int       `json:"max_uses"`
		ExpiresAt     *time.Time `json:"expires_at"`
		IsActive      *bool      `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if input.PartnerName != "" {
		updates["partner_name"] = input.PartnerName
	}
	if input.EventID != nil {
		updates["event_id"] = input.EventID
	}
	if input.DiscountType != "" {
		updates["discount_type"] = input.DiscountType
	}
	if input.DiscountValue != nil {
		updates["discount_value"] = *input.DiscountValue
	}
	if input.RewardType != "" {
		updates["reward_type"] = input.RewardType
	}
	if input.RewardValue != nil {
		updates["reward_value"] = *input.RewardValue
	}
	if input.MaxUses != nil {
		updates["max_uses"] = *input.MaxUses
	}
	if input.ExpiresAt != nil {
		updates["expires_at"] = input.ExpiresAt
	}
	if input.IsActive != nil {
		updates["is_active"] = *input.IsActive
	}

	config.DB.Model(&code).Updates(updates)
	config.DB.Preload("User").Preload("Event").First(&code, code.ID)

	c.JSON(http.StatusOK, gin.H{"success": true, "data": code})
}

// DELETE /admin/referrals/:id
func DeleteReferralCode(c *gin.Context) {
	id := c.Param("id")

	var code models.ReferralCode
	if err := config.DB.First(&code, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Referral code not found"})
		return
	}

	config.DB.Delete(&code)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Referral code deleted"})
}

// PATCH /admin/referrals/:id/status
func UpdateReferralCodeStatus(c *gin.Context) {
	id := c.Param("id")

	var input struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	var code models.ReferralCode
	if err := config.DB.First(&code, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Referral code not found"})
		return
	}

	config.DB.Model(&code).Update("is_active", input.IsActive)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": code})
}

// ─────────────────────────────────────────────
// ADMIN: Statistik per Referral Code
// ─────────────────────────────────────────────

// GET /admin/referrals/:id/stats
func GetReferralStats(c *gin.Context) {
	id := c.Param("id")

	var code models.ReferralCode
	if err := config.DB.Preload("User").Preload("Event").First(&code, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Referral code not found"})
		return
	}

	// Status yang dihitung sebagai sales (kecuali cancelled)
	validStatuses := []string{"pending", "paid", "success", "SUCCESSFUL"}

	// 1. Total orders
	var totalOrders int64
	config.DB.Model(&models.Order{}).Where("referral_code = ? AND status IN ?", code.Code, validStatuses).Count(&totalOrders)

	// 2. Total tiket terjual
	var totalTickets int64
	config.DB.Model(&models.Ticket{}).
		Joins("JOIN orders ON orders.id = tickets.order_id").
		Where("orders.referral_code = ? AND orders.status IN ?", code.Code, validStatuses).
		Count(&totalTickets)

	// 3. Total revenue (Ticket Sales Gross)
	var totalRevenue float64
	config.DB.Model(&models.Order{}).
		Select("COALESCE(SUM(total_amount - admin_fee - unique_code), 0)").
		Where("referral_code = ? AND status IN ?", code.Code, validStatuses).
		Scan(&totalRevenue)

	// 4. Total discount given
	var totalDiscount float64
	config.DB.Model(&models.Order{}).
		Select("COALESCE(SUM(discount_amount), 0)").
		Where("referral_code = ? AND status IN ?", code.Code, validStatuses).
		Scan(&totalDiscount)

	// 5. Total Pendapatan Mitra (MARKETER REVENUE)
	// Kalkulasi berdasarkan RewardType & RewardValue
	var totalEarnings float64
	if code.RewardType == "fixed" {
		totalEarnings = float64(totalOrders) * code.RewardValue
	} else if code.RewardType == "percent" {
		totalEarnings = totalRevenue * (code.RewardValue / 100)
	}

	var recentOrders []models.Order
	config.DB.Where("referral_code = ?", code.Code).
		Order("created_at DESC").
		Limit(10).
		Find(&recentOrders)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"referral_code": code,
			"stats": gin.H{
				"total_orders":   totalOrders,
				"total_tickets":  totalTickets,
				"total_revenue":  totalRevenue,  // Omzet Penjualan
				"total_discount": totalDiscount, // Diskon yang dikeluarkan
				"total_earnings": totalEarnings, // Komisi untuk mitra (Revenue mereka)
				"commission_info": gin.H{
					"type":  code.RewardType,
					"value": code.RewardValue,
				},
			},
			"recent_orders": recentOrders,
		},
	})
}

// ─────────────────────────────────────────────
// PUBLIC: Validate Referral Code
// ─────────────────────────────────────────────

// GET /referrals/validate?code=REF-XXXX&event_id=1
func ValidateReferralCode(c *gin.Context) {
	code := strings.ToUpper(strings.TrimSpace(c.Query("code")))
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Referral code is required"})
		return
	}

	var referral models.ReferralCode
	if err := config.DB.Where("code = ? AND is_active = true", code).First(&referral).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Referral code tidak ditemukan atau tidak aktif"})
		return
	}

	if referral.ExpiresAt != nil && referral.ExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Referral code sudah kadaluarsa"})
		return
	}

	if referral.MaxUses > 0 && referral.UsedCount >= referral.MaxUses {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Referral code sudah mencapai batas penggunaan"})
		return
	}

	if referral.EventID != nil {
		eventIDStr := c.Query("event_id")
		if eventIDStr != "" {
			eventID, _ := strconv.Atoi(eventIDStr)
			if uint(eventID) != *referral.EventID {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Referral code tidak berlaku untuk event ini"})
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"code":           referral.Code,
			"partner_name":   referral.PartnerName,
			"discount_type":  referral.DiscountType,
			"discount_value": referral.DiscountValue,
			"event_id":       referral.EventID,
		},
	})
}
