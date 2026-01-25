package controllers

import (
	"kartcis-backend/config"
	"kartcis-backend/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Admin Dashboard - Stats
func AdminGetStats(c *gin.Context) {
	var totalUsers int64
	var totalEvents int64
	var totalOrders int64
	var paidOrders int64
	var pendingOrders int64
	var totalRevenue float64

	config.DB.Model(&models.User{}).Count(&totalUsers)
	config.DB.Model(&models.Event{}).Count(&totalEvents)
	config.DB.Model(&models.Order{}).Count(&totalOrders)
	config.DB.Model(&models.Order{}).Where("status = ?", "paid").Count(&paidOrders)
	config.DB.Model(&models.Order{}).Where("status = ?", "pending").Count(&pendingOrders)

	// Calculate revenue (sum of paid orders)
	type Result struct {
		Total float64
	}
	var result Result
	config.DB.Model(&models.Order{}).Where("status = ?", "paid").Select("sum(total_amount) as total").Scan(&result)
	totalRevenue = result.Total

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"total_users":          totalUsers,
			"total_events":         totalEvents,
			"total_transactions":   totalOrders,
			"paid_transactions":    paidOrders,
			"pending_transactions": pendingOrders,
			"total_revenue":        totalRevenue,
		},
	})
}

// Admin Get Revenue Chart Data
func AdminGetRevenueChart(c *gin.Context) {
	// Group by day for last 30 days
	type DailyRevenue struct {
		Date  string  `json:"date"`
		Total float64 `json:"total"`
	}

	var revenues []DailyRevenue

	last30Days := time.Now().AddDate(0, 0, -30)
	var orders []models.Order
	config.DB.Where("status = ? AND created_at >= ?", "paid", last30Days).Find(&orders)

	revenueMap := make(map[string]float64)
	// Init map
	for i := 0; i < 30; i++ {
		dateStr := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		revenueMap[dateStr] = 0
	}

	for _, o := range orders {
		dateStr := o.CreatedAt.Format("2006-01-02")
		revenueMap[dateStr] += o.TotalAmount
	}

	// specific format for chart (array)
	for date, total := range revenueMap {
		revenues = append(revenues, DailyRevenue{Date: date, Total: total})
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": revenues})
}

// GetTransactionsOverview
func GetTransactionsOverview(c *gin.Context) {
	// Recent transactions
	var orders []models.Order
	config.DB.Order("created_at desc").Limit(5).Find(&orders)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": orders})
}

// GetEventsOverview
func GetEventsOverview(c *gin.Context) {
	// Recent events or top events
	var events []models.Event
	config.DB.Order("created_at desc").Limit(5).Find(&events)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": events})
}

// GetUsersOverview
func GetUsersOverview(c *gin.Context) {
	// Recent users
	var users []models.User
	config.DB.Order("created_at desc").Limit(5).Find(&users)
	c.JSON(http.StatusOK, gin.H{"success": true, "data": users})
}
