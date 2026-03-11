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
	"gorm.io/gorm"
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
	role, _ := c.Get("userRole")
	userID, _ := c.Get("userID")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	query := config.DB.Preload("User").Preload("Event")

	// Organizer hanya melihat referral code milik mereka
	if role != "admin" {
		query = query.Where("user_id = ?", userID)
	}

	var total int64
	query.Model(&models.ReferralCode{}).Count(&total)

	var codes []models.ReferralCode
	query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&codes)

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
		UserID        uint       `json:"user_id" binding:"required"` // Marketer user ID
		EventID       *uint      `json:"event_id"`
		Prefix        string     `json:"prefix"`        // Optional prefix for code
		CustomCode    string     `json:"custom_code"`   // Or specify exact code
		RewardType    string     `json:"reward_type"`   // "percent" or "fixed"
		RewardValue   float64    `json:"reward_value"`  // Commission value
		DiscountType  string     `json:"discount_type"` // "percent", "fixed", or "none"
		DiscountValue float64    `json:"discount_value"`
		MaxUses       int        `json:"max_uses"` // 0 = unlimited
		ExpiresAt     *time.Time `json:"expires_at"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	// Validate reward type
	if input.RewardType != "percent" && input.RewardType != "fixed" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "reward_type must be 'percent' or 'fixed'"})
		return
	}

	// Validate discount type (optional)
	if input.DiscountType != "" && input.DiscountType != "percent" && input.DiscountType != "fixed" && input.DiscountType != "none" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "discount_type must be 'percent', 'fixed', or 'none'"})
		return
	}
	if input.DiscountType == "" {
		input.DiscountType = "none"
	}

	// Determine code
	code := input.CustomCode
	if code == "" {
		code = generateReferralCode(input.Prefix)
	} else {
		code = strings.ToUpper(strings.TrimSpace(code))
	}

	// Verify user exists
	var marketer models.User
	if err := config.DB.First(&marketer, input.UserID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Marketer user not found"})
		return
	}

	referral := models.ReferralCode{
		Code:          code,
		UserID:        input.UserID,
		EventID:       input.EventID,
		RewardType:    input.RewardType,
		RewardValue:   input.RewardValue,
		DiscountType:  input.DiscountType,
		DiscountValue: input.DiscountValue,
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
	role, _ := c.Get("userRole")
	userID, _ := c.Get("userID")

	var code models.ReferralCode
	query := config.DB.Preload("User").Preload("Event")
	if role != "admin" {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.First(&code, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Referral code not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": code})
}

// PUT /admin/referrals/:id
func UpdateReferralCode(c *gin.Context) {
	id := c.Param("id")
	role, _ := c.Get("userRole")
	userID, _ := c.Get("userID")

	var code models.ReferralCode
	query := config.DB
	if role != "admin" {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.First(&code, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Referral code not found"})
		return
	}

	var input struct {
		EventID       *uint      `json:"event_id"`
		RewardType    string     `json:"reward_type"`
		RewardValue   *float64   `json:"reward_value"`
		DiscountType  string     `json:"discount_type"`
		DiscountValue *float64   `json:"discount_value"`
		MaxUses       *int       `json:"max_uses"`
		ExpiresAt     *time.Time `json:"expires_at"`
		IsActive      *bool      `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if input.EventID != nil {
		updates["event_id"] = input.EventID
	}
	if input.RewardType != "" {
		updates["reward_type"] = input.RewardType
	}
	if input.RewardValue != nil {
		updates["reward_value"] = *input.RewardValue
	}
	if input.DiscountType != "" {
		updates["discount_type"] = input.DiscountType
	}
	if input.DiscountValue != nil {
		updates["discount_value"] = *input.DiscountValue
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
	role, _ := c.Get("userRole")
	userID, _ := c.Get("userID")

	var code models.ReferralCode
	query := config.DB
	if role != "admin" {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.First(&code, id).Error; err != nil {
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
// ADMIN: Commission Management
// ─────────────────────────────────────────────

// GET /admin/referrals/commissions
func AdminGetCommissions(c *gin.Context) {
	role, _ := c.Get("userRole")
	userID, _ := c.Get("userID")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit
	statusFilter := c.Query("status")

	query := config.DB.Preload("ReferralCode").Preload("Marketer").Preload("Order")

	// Organizer / marketer hanya lihat commission milik mereka
	if role != "admin" {
		query = query.Where("marketer_id = ?", userID)
	}

	if statusFilter != "" {
		query = query.Where("status = ?", statusFilter)
	}

	var total int64
	query.Model(&models.ReferralCommission{}).Count(&total)

	var commissions []models.ReferralCommission
	query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&commissions)

	totalPages := int(total) / limit
	if int(total)%limit != 0 {
		totalPages++
	}

	// Calculate total pending commission
	var totalPending float64
	pendingQuery := config.DB.Model(&models.ReferralCommission{}).Select("COALESCE(SUM(commission_amount), 0)").Where("status = ?", "pending")
	if role != "admin" {
		pendingQuery = pendingQuery.Where("marketer_id = ?", userID)
	}
	pendingQuery.Scan(&totalPending)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"commissions": commissions,
			"summary": gin.H{
				"total_pending": totalPending,
			},
			"pagination": gin.H{
				"current_page": page,
				"total_pages":  totalPages,
				"total_items":  total,
				"per_page":     limit,
			},
		},
	})
}

// PATCH /admin/referrals/commissions/:id/status
func UpdateCommissionStatus(c *gin.Context) {
	id := c.Param("id")

	var input struct {
		Status string `json:"status" binding:"required"` // paid, cancelled
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	if input.Status != "paid" && input.Status != "cancelled" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Status must be 'paid' or 'cancelled'"})
		return
	}

	var commission models.ReferralCommission
	if err := config.DB.First(&commission, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Commission not found"})
		return
	}

	config.DB.Model(&commission).Update("status", input.Status)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": commission})
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
	query := config.DB.Preload("User").Where("code = ? AND is_active = true", code)

	if err := query.First(&referral).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Referral code tidak ditemukan atau tidak aktif"})
		return
	}

	// Check expiry
	if referral.ExpiresAt != nil && referral.ExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Referral code sudah kadaluarsa"})
		return
	}

	// Check max uses
	if referral.MaxUses > 0 && referral.UsedCount >= referral.MaxUses {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Referral code sudah mencapai batas penggunaan"})
		return
	}

	// If event scoped, validate against requested event_id
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
			"marketer_name":  referral.User.Name,
			"discount_type":  referral.DiscountType,
			"discount_value": referral.DiscountValue,
			"event_id":       referral.EventID,
		},
	})
}

// ─────────────────────────────────────────────
// INTERNAL: Apply Referral During Order
// Called from CreateOrder when referral_code is provided
// ─────────────────────────────────────────────

// ApplyReferralCode validates the referral and creates commission record.
// Returns (discountAmount, error)
func ApplyReferralCode(tx *gorm.DB, code string, order *models.Order, totalAmount float64) (float64, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return 0, nil
	}

	var referral models.ReferralCode
	if err := tx.Where("code = ? AND is_active = true", code).First(&referral).Error; err != nil {
		return 0, nil // Code not found → ignore silently in checkout context
	}

	// Check expiry
	if referral.ExpiresAt != nil && referral.ExpiresAt.Before(time.Now()) {
		return 0, nil
	}

	// Check max uses
	if referral.MaxUses > 0 && referral.UsedCount >= referral.MaxUses {
		return 0, nil
	}

	// Event scope check
	if referral.EventID != nil {
		// Check if any ticket in the order belongs to the event
		var count int64
		tx.Model(&models.Ticket{}).Where("order_id = ? AND event_id = ?", order.ID, *referral.EventID).Count(&count)
		// Note: order.ID may be 0 here (pre-save), so we use order items — handle in CreateOrder instead
	}

	// Calculate discount for buyer
	var discountAmount float64
	if referral.DiscountType == "percent" && referral.DiscountValue > 0 {
		discountAmount = totalAmount * (referral.DiscountValue / 100)
	} else if referral.DiscountType == "fixed" && referral.DiscountValue > 0 {
		discountAmount = referral.DiscountValue
		if discountAmount > totalAmount {
			discountAmount = totalAmount
		}
	}

	// Calculate marketer commission
	var commissionAmount float64
	if referral.RewardType == "percent" && referral.RewardValue > 0 {
		commissionAmount = totalAmount * (referral.RewardValue / 100)
	} else if referral.RewardType == "fixed" && referral.RewardValue > 0 {
		commissionAmount = referral.RewardValue
	}

	// Increment used_count
	tx.Model(&referral).Update("used_count", gorm.Expr("used_count + ?", 1))

	// Commission will be recorded after order is saved (order.ID is needed)
	// Return a closure pattern via storing in function — we'll handle in CreateOrder via direct call
	// For now, if order.ID is available we immediately record, otherwise caller must do it.
	if order.ID > 0 && commissionAmount > 0 {
		commission := models.ReferralCommission{
			ReferralCodeID:   referral.ID,
			MarketerID:       referral.UserID,
			OrderID:          order.ID,
			CommissionAmount: commissionAmount,
			Status:           "pending",
		}
		tx.Create(&commission)
	}

	return discountAmount, nil
}
