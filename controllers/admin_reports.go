package controllers

import (
	"fmt"
	"kartcis-backend/config"
	"kartcis-backend/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Sales Report
func AdminGetSalesReport(c *gin.Context) {
	// Basic aggregation: Sales by day/month
	// For MVP, return simple daily stats for last 30 days
	last30Days := time.Now().AddDate(0, 0, -30)

	var sales []struct {
		Date  string  `json:"date"`
		Total float64 `json:"total"`
		Count int64   `json:"count"`
	}

	// GORM/SQL raw query might be easiest for aggregation
	// Assuming Postgres
	err := config.DB.Model(&models.Order{}).
		Select("to_char(created_at, 'YYYY-MM-DD') as date, sum(total_amount) as total, count(*) as count").
		Where("status = ? AND created_at >= ?", "paid", last30Days).
		Group("to_char(created_at, 'YYYY-MM-DD')").
		Order("date ASC").
		Scan(&sales).Error

	if err != nil {
		// Fallback for non-postgres or error
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to generate report", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": sales})
}

// Events Report
func AdminGetEventsReport(c *gin.Context) {
	// Ticket sales per event
	var report []struct {
		EventTitle  string  `json:"event_title"`
		SoldTickets int64   `json:"sold_tickets"`
		Revenue     float64 `json:"revenue"`
	}

	config.DB.Table("orders").
		Joins("JOIN tickets ON tickets.order_id = orders.id").
		Joins("JOIN events ON tickets.event_id = events.id").
		Joins("JOIN ticket_types ON tickets.ticket_type_id = ticket_types.id").
		Where("orders.status = ?", "paid").
		Select("events.title as event_title, count(tickets.id) as sold_tickets, sum(ticket_types.price) as revenue").
		Group("events.id, events.title").
		Order("revenue DESC").
		Limit(10).
		Scan(&report)

	c.JSON(http.StatusOK, gin.H{"success": true, "data": report})
}

// Download/Export Report
func AdminExportReport(c *gin.Context) {
	// Re-run sales report query for export
	last30Days := time.Now().AddDate(0, 0, -30)
	var sales []struct {
		Date  string
		Total float64
		Count int64
	}

	config.DB.Model(&models.Order{}).
		Select("to_char(created_at, 'YYYY-MM-DD') as date, sum(total_amount) as total, count(*) as count").
		Where("status = ? AND created_at >= ?", "paid", last30Days).
		Group("to_char(created_at, 'YYYY-MM-DD')").
		Order("date ASC").
		Scan(&sales)

	// Generate CSV
	csvContent := "Date,TotalSales,TransactionCount\n"
	for _, s := range sales {
		csvContent += fmt.Sprintf("%s,%.2f,%d\n", s.Date, s.Total, s.Count)
	}

	c.Header("Content-Disposition", "attachment; filename=sales_report.csv")
	c.Data(http.StatusOK, "text/csv", []byte(csvContent))
}

// GetUsersReport
func GetUsersReport(c *gin.Context) {
	// New user registration stats
	last30Days := time.Now().AddDate(0, 0, -30)
	var stats []struct {
		Date  string
		Count int64
	}
	// Warning: to_char is Postgres specific. If using SQLite/MySQL, syntax differs.
	config.DB.Model(&models.User{}).
		Select("to_char(created_at, 'YYYY-MM-DD') as date, count(*) as count").
		Where("created_at >= ?", last30Days).
		Group("date").
		Scan(&stats)

	c.JSON(http.StatusOK, gin.H{"success": true, "data": stats})
}

// GetTopEventsReport
func GetTopEventsReport(c *gin.Context) {
	// Similar to EventsReport but focused on top 5
	AdminGetEventsReport(c) // Reusing logic
}
