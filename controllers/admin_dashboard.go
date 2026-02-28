package controllers

import (
	"kartcis-backend/config"
	"kartcis-backend/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Admin Dashboard - Stats
// Admin Dashboard - Stats
func AdminGetStats(c *gin.Context) {
	var totalUsers int64
	var totalEvents int64
	var totalOrders int64
	var paidOrders int64
	var pendingOrders int64
	var totalRevenue float64

	role, _ := c.Get("userRole")
	userID, _ := c.Get("userID")
	eventID := c.Query("event_id")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	startStr, endStr := "", ""
	if startDate != "" {
		startStr = startDate + " 00:00:00"
	}
	if endDate != "" {
		endStr = endDate + " 23:59:59"
	}

	if role == "organizer" {
		// Scoped Stats
		// 1. Total Customers (Users who bought tickets)
		queryUsers := config.DB.Table("users").
			Joins("JOIN orders ON orders.user_id = users.id").
			Joins("JOIN tickets ON tickets.order_id = orders.id").
			Joins("JOIN events ON events.id = tickets.event_id").
			Where("events.organizer_id = ?", userID)
		if eventID != "" {
			queryUsers = queryUsers.Where("events.id = ?", eventID)
		}
		if startStr != "" {
			queryUsers = queryUsers.Where("orders.created_at >= ?", startStr)
		}
		if endStr != "" {
			queryUsers = queryUsers.Where("orders.created_at <= ?", endStr)
		}
		queryUsers.Distinct("users.id").Count(&totalUsers)

		// 2. Events Owned
		queryEvents := config.DB.Model(&models.Event{}).Where("organizer_id = ?", userID)
		if eventID != "" {
			queryEvents = queryEvents.Where("id = ?", eventID)
		}
		if startStr != "" {
			queryEvents = queryEvents.Where("created_at >= ?", startStr)
		}
		if endStr != "" {
			queryEvents = queryEvents.Where("created_at <= ?", endStr)
		}
		queryEvents.Count(&totalEvents)

		// 3. Transactions (Orders involving their events)
		// Note: Orders might duplicate if joined, use distinct order ID counting
		queryOrders := config.DB.Table("orders").
			Joins("JOIN tickets ON tickets.order_id = orders.id").
			Joins("JOIN events ON events.id = tickets.event_id").
			Where("events.organizer_id = ?", userID)
		if eventID != "" {
			queryOrders = queryOrders.Where("events.id = ?", eventID)
		}
		if startStr != "" {
			queryOrders = queryOrders.Where("orders.created_at >= ?", startStr)
		}
		if endStr != "" {
			queryOrders = queryOrders.Where("orders.created_at <= ?", endStr)
		}
		queryOrders.Distinct("orders.id").Count(&totalOrders)

		queryPaid := config.DB.Table("orders").
			Joins("JOIN tickets ON tickets.order_id = orders.id").
			Joins("JOIN events ON events.id = tickets.event_id").
			Where("events.organizer_id = ? AND orders.status = ?", userID, "paid")
		if eventID != "" {
			queryPaid = queryPaid.Where("events.id = ?", eventID)
		}
		if startStr != "" {
			queryPaid = queryPaid.Where("orders.created_at >= ?", startStr)
		}
		if endStr != "" {
			queryPaid = queryPaid.Where("orders.created_at <= ?", endStr)
		}
		queryPaid.Distinct("orders.id").Count(&paidOrders)

		queryPending := config.DB.Table("orders").
			Joins("JOIN tickets ON tickets.order_id = orders.id").
			Joins("JOIN events ON events.id = tickets.event_id").
			Where("events.organizer_id = ? AND orders.status = ?", userID, "pending")
		if eventID != "" {
			queryPending = queryPending.Where("events.id = ?", eventID)
		}
		if startStr != "" {
			queryPending = queryPending.Where("orders.created_at >= ?", startStr)
		}
		if endStr != "" {
			queryPending = queryPending.Where("orders.created_at <= ?", endStr)
		}
		queryPending.Distinct("orders.id").Count(&pendingOrders)

		// 4. Net Revenue (Sum of THEIR ticket prices MINUS Platform Fee in paid orders)
		type Result struct {
			Total float64
		}
		var result Result
		queryRevenue := config.DB.Table("tickets").
			Joins("JOIN orders ON orders.id = tickets.order_id").
			Joins("JOIN ticket_types ON ticket_types.id = tickets.ticket_type_id").
			Joins("JOIN events ON events.id = tickets.event_id").
			Where("orders.status = ? AND events.organizer_id = ?", "paid", userID)
		if eventID != "" {
			queryRevenue = queryRevenue.Where("events.id = ?", eventID)
		}
		if startStr != "" {
			queryRevenue = queryRevenue.Where("orders.created_at >= ?", startStr)
		}
		if endStr != "" {
			queryRevenue = queryRevenue.Where("orders.created_at <= ?", endStr)
		}
		queryRevenue.Select("COALESCE(SUM(ticket_types.price - (ticket_types.price * events.fee_percentage / 100)), 0) as total").
			Scan(&result)
		totalRevenue = result.Total

	} else {
		// Admin (Global)
		if eventID != "" {
			queryUsers := config.DB.Table("users").
				Joins("JOIN orders ON orders.user_id = users.id").
				Joins("JOIN tickets ON tickets.order_id = orders.id").
				Where("tickets.event_id = ?", eventID)
			if startStr != "" {
				queryUsers = queryUsers.Where("orders.created_at >= ?", startStr)
			}
			if endStr != "" {
				queryUsers = queryUsers.Where("orders.created_at <= ?", endStr)
			}
			queryUsers.Distinct("users.id").Count(&totalUsers)

			queryEvents := config.DB.Model(&models.Event{}).Where("id = ?", eventID)
			if startStr != "" {
				queryEvents = queryEvents.Where("created_at >= ?", startStr)
			}
			if endStr != "" {
				queryEvents = queryEvents.Where("created_at <= ?", endStr)
			}
			queryEvents.Count(&totalEvents)

			queryOrders := config.DB.Table("orders").
				Joins("JOIN tickets ON tickets.order_id = orders.id").
				Where("tickets.event_id = ?", eventID)
			if startStr != "" {
				queryOrders = queryOrders.Where("orders.created_at >= ?", startStr)
			}
			if endStr != "" {
				queryOrders = queryOrders.Where("orders.created_at <= ?", endStr)
			}
			queryOrders.Distinct("orders.id").Count(&totalOrders)

			queryPaid := config.DB.Table("orders").
				Joins("JOIN tickets ON tickets.order_id = orders.id").
				Where("tickets.event_id = ? AND orders.status = ?", eventID, "paid")
			if startStr != "" {
				queryPaid = queryPaid.Where("orders.created_at >= ?", startStr)
			}
			if endStr != "" {
				queryPaid = queryPaid.Where("orders.created_at <= ?", endStr)
			}
			queryPaid.Distinct("orders.id").Count(&paidOrders)

			queryPending := config.DB.Table("orders").
				Joins("JOIN tickets ON tickets.order_id = orders.id").
				Where("tickets.event_id = ? AND orders.status = ?", eventID, "pending")
			if startStr != "" {
				queryPending = queryPending.Where("orders.created_at >= ?", startStr)
			}
			if endStr != "" {
				queryPending = queryPending.Where("orders.created_at <= ?", endStr)
			}
			queryPending.Distinct("orders.id").Count(&pendingOrders)

			type Result struct {
				Total float64
			}
			var result Result
			queryRevenue := config.DB.Table("tickets").
				Joins("JOIN orders ON orders.id = tickets.order_id").
				Joins("JOIN ticket_types ON ticket_types.id = tickets.ticket_type_id").
				Where("orders.status = ? AND tickets.event_id = ?", "paid", eventID)
			if startStr != "" {
				queryRevenue = queryRevenue.Where("orders.created_at >= ?", startStr)
			}
			if endStr != "" {
				queryRevenue = queryRevenue.Where("orders.created_at <= ?", endStr)
			}
			queryRevenue.Select("COALESCE(SUM(ticket_types.price), 0) as total").
				Scan(&result)
			totalRevenue = result.Total
		} else {
			queryUsers := config.DB.Model(&models.User{})
			if startStr != "" {
				queryUsers = queryUsers.Where("created_at >= ?", startStr)
			}
			if endStr != "" {
				queryUsers = queryUsers.Where("created_at <= ?", endStr)
			}
			queryUsers.Count(&totalUsers)

			queryEvents := config.DB.Model(&models.Event{})
			if startStr != "" {
				queryEvents = queryEvents.Where("created_at >= ?", startStr)
			}
			if endStr != "" {
				queryEvents = queryEvents.Where("created_at <= ?", endStr)
			}
			queryEvents.Count(&totalEvents)

			queryOrders := config.DB.Model(&models.Order{})
			if startStr != "" {
				queryOrders = queryOrders.Where("created_at >= ?", startStr)
			}
			if endStr != "" {
				queryOrders = queryOrders.Where("created_at <= ?", endStr)
			}
			queryOrders.Count(&totalOrders)

			queryPaid := config.DB.Model(&models.Order{}).Where("status = ?", "paid")
			if startStr != "" {
				queryPaid = queryPaid.Where("created_at >= ?", startStr)
			}
			if endStr != "" {
				queryPaid = queryPaid.Where("created_at <= ?", endStr)
			}
			queryPaid.Count(&paidOrders)

			queryPending := config.DB.Model(&models.Order{}).Where("status = ?", "pending")
			if startStr != "" {
				queryPending = queryPending.Where("created_at >= ?", startStr)
			}
			if endStr != "" {
				queryPending = queryPending.Where("created_at <= ?", endStr)
			}
			queryPending.Count(&pendingOrders)

			// Calculate revenue (sum of paid orders)
			type Result struct {
				Total float64
			}
			var result Result
			queryRevenue := config.DB.Model(&models.Order{}).Where("status = ?", "paid")
			if startStr != "" {
				queryRevenue = queryRevenue.Where("created_at >= ?", startStr)
			}
			if endStr != "" {
				queryRevenue = queryRevenue.Where("created_at <= ?", endStr)
			}
			queryRevenue.Select("COALESCE(SUM(total_amount), 0) as total").Scan(&result)
			totalRevenue = result.Total
		}
	}

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
// Admin Get Revenue Chart Data
func AdminGetRevenueChart(c *gin.Context) {
	// Group by day for last 30 days
	type DailyRevenue struct {
		Date  string  `json:"date"`
		Total float64 `json:"total"`
	}

	var revenues []DailyRevenue
	revenueMap := make(map[string]float64)

	last30Days := time.Now().AddDate(0, 0, -30)

	// Init map
	for i := 0; i < 30; i++ {
		dateStr := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		revenueMap[dateStr] = 0
	}

	role, _ := c.Get("userRole")
	userID, _ := c.Get("userID")

	if role == "organizer" {
		// Organizer: Sum Ticket Prices
		type TicketResult struct {
			Price     float64
			CreatedAt time.Time
		}
		var ticketResults []TicketResult

		config.DB.Table("tickets").
			Joins("JOIN orders ON orders.id = tickets.order_id").
			Joins("JOIN ticket_types ON ticket_types.id = tickets.ticket_type_id").
			Joins("JOIN events ON events.id = tickets.event_id").
			Where("orders.status = ? AND events.organizer_id = ? AND orders.created_at >= ?", "paid", userID, last30Days).
			Select("ticket_types.price - (ticket_types.price * events.fee_percentage / 100) as price, orders.created_at").
			Scan(&ticketResults)

		for _, t := range ticketResults {
			dateStr := t.CreatedAt.Format("2006-01-02")
			// Only add if within map keys (optimization, though query already filters)
			if _, ok := revenueMap[dateStr]; ok {
				revenueMap[dateStr] += t.Price
			}
		}

	} else {
		// Admin: Sum Order Amounts
		var orders []models.Order
		config.DB.Where("status = ? AND created_at >= ?", "paid", last30Days).Find(&orders)

		for _, o := range orders {
			dateStr := o.CreatedAt.Format("2006-01-02")
			if _, ok := revenueMap[dateStr]; ok {
				revenueMap[dateStr] += o.TotalAmount
			}
		}
	}

	// specific format for chart (array)
	// Map iteration order is random, need to sort? Chart usually expects sorted?
	// The prompt didn't ask for sorting but frontend probably needs it.
	// Let's iterate 0..29 backwards or forwards to build array.
	// Actually, let's keep it simple as before, but the previous code iterated the MAP.
	// `for date, total := range revenueMap` -> random order.
	// Better to iterate dates.

	// The User didn't complain about order before, but let's be nice.
	// But let's stick to previous behavior strict replacement if possible to minimize diff,
	// EXCEPT previous behavior was iterating map which is random.
	// I will just iterate map to match previous implementation structure,
	// or iterate dates to be better. Let's iterate dates.

	for i := 29; i >= 0; i-- {
		dateStr := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		revenues = append(revenues, DailyRevenue{Date: dateStr, Total: revenueMap[dateStr]})
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
