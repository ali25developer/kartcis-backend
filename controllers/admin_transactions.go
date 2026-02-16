package controllers

import (
	"fmt"
	"kartcis-backend/config"
	"kartcis-backend/jobs"
	"kartcis-backend/models"
	"kartcis-backend/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func AdminGetTransactions(c *gin.Context) {
	orders := []models.Order{}
	var totalItems int64

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	query := config.DB.Model(&models.Order{})
	status := c.Query("status")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Search filter
	search := c.Query("search")
	if search != "" {
		query = query.Where("order_number ILIKE ? OR customer_email ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	// Count Total
	query.Count(&totalItems)

	// Fetch Data
	query.Preload("Tickets.Event").Preload("Tickets.TicketType").Order("created_at desc").Limit(limit).Offset(offset).Find(&orders)

	totalPages := int(totalItems) / limit
	if int(totalItems)%limit != 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"transactions": orders,
			"pagination": gin.H{
				"current_page": page,
				"total_pages":  totalPages,
				"total_items":  totalItems,
				"per_page":     limit,
			},
			"stats": gin.H{
				"total": totalItems,
			},
		},
	})
}

func AdminGetTransactionDetail(c *gin.Context) {
	id := c.Param("id")
	var order models.Order

	if err := config.DB.Preload("Tickets.Event").Preload("Tickets.TicketType").First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": order})
}

func CancelTransaction(c *gin.Context) {
	id := c.Param("id")
	var order models.Order
	if err := config.DB.Preload("Tickets").First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Order not found"})
		return
	}

	// Final State Protection
	if order.Status == "paid" || order.Status == "cancelled" || order.Status == "expired" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": fmt.Sprintf("Cannot cancel because the transaction is already %s", order.Status),
		})
		return
	}

	tx := config.DB.Begin()

	order.Status = "cancelled"
	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to cancel transaction"})
		return
	}

	// Restore Quota
	if err := utils.RestoreQuota(tx, order.ID); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to restore quota"})
		return
	}

	// Record history
	tx.Create(&models.OrderStatusHistory{
		OrderID:   order.ID,
		Status:    "cancelled",
		Notes:     "Cancelled by Admin",
		CreatedAt: time.Now(),
	})

	tx.Commit()

	// Send Cancellation Email
	utils.SendOrderCancelledEmail(order, "Dibatalkan oleh Admin")

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Transaction cancelled and quota restored"})
}

func MarkTransactionPaid(c *gin.Context) {
	id := c.Param("id")
	var order models.Order
	if err := config.DB.First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Order not found"})
		return
	}

	// Final State Protection
	if order.Status == "paid" || order.Status == "cancelled" || order.Status == "expired" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": fmt.Sprintf("Cannot mark as paid because the transaction is already %s", order.Status),
		})
		return
	}

	order.Status = "paid"
	config.DB.Save(&order)

	// Record history
	config.DB.Create(&models.OrderStatusHistory{
		OrderID:   order.ID,
		Status:    "paid",
		Notes:     "Marked as paid by Admin",
		CreatedAt: time.Now(),
	})

	// Send Confirmation Email
	var tickets []models.Ticket
	config.DB.Preload("Event").Preload("TicketType").Where("order_id = ?", order.ID).Find(&tickets)
	utils.SendTicketEmail(order, tickets)

	// Record history: Email Sent
	config.DB.Create(&models.OrderStatusHistory{
		OrderID:   order.ID,
		Status:    "paid",
		Notes:     "E-Ticket email sent to customer",
		CreatedAt: time.Now(),
	})

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Transaction marked as paid and email sent"})
}

func ResendTicketEmail(c *gin.Context) {
	id := c.Param("id")
	// In a real system, we would trigger the mailer service here.
	// Since we don't have SMTP setup, we will log this action to ActivityLog to show "System attempted send".
	// This makes it a "real" database action rather than a static return.

	// Create activity log (assuming middleware sets userID, but admin might be implicit if we don't have context)
	// We'll skip UserID linkage if complex, or just log to a system log/audit table.
	// For now, let's just return success but ensure we check if order exists.
	var order models.Order
	if err := config.DB.First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Order not found"})
		return
	}

	// Fetch tickets and send email
	var tickets []models.Ticket
	if err := config.DB.Preload("Event").Preload("TicketType").Where("order_id = ?", order.ID).Find(&tickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to fetch tickets"})
		return
	}

	utils.SendTicketEmail(order, tickets)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Email resend request for Order %s has been processed", order.OrderNumber),
	})
}

func ExportTransactions(c *gin.Context) {
	var orders []models.Order
	config.DB.Find(&orders)

	// Simple CSV generation
	csvContent := "ID,OrderNumber,CustomerName,Amount,Status,Date\n"
	for _, o := range orders {
		csvContent += fmt.Sprintf("%d,%s,%s,%.2f,%s,%s\n", o.ID, o.OrderNumber, o.CustomerName, o.TotalAmount, o.Status, o.CreatedAt.Format("2006-01-02"))
	}

	c.Header("Content-Disposition", "attachment; filename=transactions.csv")
	c.Data(http.StatusOK, "text/csv", []byte(csvContent))
}

func GetTransactionTimeline(c *gin.Context) {
	id := c.Param("id")
	var order models.Order
	if err := config.DB.First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Order not found"})
		return
	}

	var histories []models.OrderStatusHistory
	config.DB.Where("order_id = ?", id).Order("created_at asc").Find(&histories)

	c.JSON(http.StatusOK, gin.H{"success": true, "data": histories})
}

func GetRevenueSummary(c *gin.Context) {
	// Summary of revenue by status or total
	var total float64
	type Result struct {
		Total float64
	}
	var result Result
	config.DB.Model(&models.Order{}).Where("status = ?", "paid").Select("sum(total_amount) as total").Scan(&result)
	total = result.Total

	// Yesterday vs Today
	today := time.Now().Truncate(24 * time.Hour)
	var todayRev, yesterdayRev float64
	config.DB.Model(&models.Order{}).Where("status = ? AND created_at >= ?", "paid", today).Select("sum(total_amount) as total").Scan(&result)
	todayRev = result.Total

	yesterday := today.AddDate(0, 0, -1)
	config.DB.Model(&models.Order{}).Where("status = ? AND created_at >= ? AND created_at < ?", "paid", yesterday, today).Select("sum(total_amount) as total").Scan(&result)
	yesterdayRev = result.Total

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total_revenue":     total,
			"today_revenue":     todayRev,
			"yesterday_revenue": yesterdayRev,
		},
	})
}
func UpdateTransactionStatus(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": "Status is required"})
		return
	}

	var order models.Order
	if err := config.DB.First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "Order not found"})
		return
	}

	// Final State Protection: Once paid, cancelled or expired, do not allow further status changes
	if order.Status == "paid" || order.Status == "cancelled" || order.Status == "expired" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": fmt.Sprintf("Cannot change status because the transaction is already in a final state: %s", order.Status),
		})
		return
	}

	tx := config.DB.Begin()

	oldStatus := order.Status
	order.Status = input.Status
	// If marked as paid, update paid_at
	if input.Status == "paid" && order.PaidAt == nil {
		now := time.Now()
		order.PaidAt = &now
	}

	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to update status"})
		return
	}

	// If status changed to cancelled/expired, restore quota
	if (input.Status == "cancelled" || input.Status == "expired") && (oldStatus != "cancelled" && oldStatus != "expired") {
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
		Notes:     "Status updated by Admin",
		CreatedAt: time.Now(),
	})

	tx.Commit()

	// If status changed to cancelled/expired, send email
	if (input.Status == "cancelled" || input.Status == "expired") && (oldStatus != "cancelled" && oldStatus != "expired") {
		reason := "Pesanan telah dibatalkan"
		if input.Status == "expired" {
			reason = "Waktu pembayaran telah habis (Expired)"
		}
		utils.SendOrderCancelledEmail(order, reason)
	}

	// If status changed to paid, send email
	if input.Status == "paid" && oldStatus != "paid" {
		var tickets []models.Ticket
		config.DB.Preload("Event").Preload("TicketType").Where("order_id = ?", order.ID).Find(&tickets)
		utils.SendTicketEmail(order, tickets)

		// Record history: Email Sent
		config.DB.Create(&models.OrderStatusHistory{
			OrderID:   order.ID,
			Status:    "paid",
			Notes:     "E-Ticket email sent to customer (Status update)",
			CreatedAt: time.Now(),
		})
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Transaction status updated", "data": order})
}

func AdminTriggerScraping(c *gin.Context) {
	// Trigger the background job logic immediately
	go jobs.CheckBankJagoEmails()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Manual email scraping triggered. Check transaction logs in a few moments.",
	})
}
