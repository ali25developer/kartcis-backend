package routes

import (
	// Added
	"kartcis-backend/controllers"
	"kartcis-backend/middleware"
	"net/http"
	"os"

	// Added
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Middleware
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.SmartLogger()) // Added Smart Logging

	// Root Handler
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":    "Welcome to Kartcis API",
			"docs":       "See /api-spec for documentation", // Conceptual
			"version":    "1.0.0",
			"hot_reload": "active",
		})
	})

	apiPrefix := os.Getenv("API_PREFIX")
	if apiPrefix == "" {
		apiPrefix = "/api/v1"
	}
	v1 := r.Group(apiPrefix)

	// Static Files (Images)
	v1.Static("/uploads", "./uploads")

	// Auth
	auth := v1.Group("/auth")
	{
		auth.POST("/register", controllers.Register)
		auth.POST("/login", controllers.Login)
		auth.POST("/logout", middleware.AuthMiddleware(), controllers.Logout)
		auth.GET("/me", middleware.AuthMiddleware(), controllers.GetMe)
		auth.PUT("/profile", middleware.AuthMiddleware(), controllers.UpdateProfile) // Added
		auth.POST("/forgot-password", controllers.ForgotPassword)
		auth.POST("/reset-password", controllers.ResetPassword) // Added
	}

	v1.GET("/categories", controllers.GetCategories)
	v1.GET("/categories/:slug", controllers.GetCategoryDetail)

	// Public Events
	v1.GET("/events", controllers.GetEvents)
	v1.GET("/events/popular", controllers.GetPopularEvents)
	v1.GET("/events/upcoming", controllers.GetUpcomingEvents)
	v1.GET("/events/featured", controllers.GetFeaturedEvents)
	v1.GET("/events/:slug", controllers.GetEventDetail) // Renamed :id to :slug

	v1.GET("/cities", controllers.GetCities)

	// Tickets (Mixed Auth/Public)
	tickets := v1.Group("/tickets")
	{
		tickets.GET("/my-tickets", middleware.AuthMiddleware(), controllers.GetMyTickets)
		tickets.GET("/:code", controllers.GetTicketDetail)
		tickets.GET("/:code/verify", controllers.VerifyTicket)
		tickets.GET("/:code/download", controllers.DownloadTicketPDF)

		// Check-in (Admin/Scanner)
		// Spec says 👑 Admin Only or Scanner
		tickets.POST("/check-in", middleware.AuthMiddleware(), requireAdmin(), controllers.CheckInTicket)
	}

	// Orders (User)
	userOrders := v1.Group("/orders")
	v1.POST("/orders/payment-callback", controllers.PaymentCallback)
	// Let's check SimulatePayment in order.go: `orderNumber := c.Param("order_number")`. Yes.
	// So if I rename route param to :order_number, I must update controller too.

	// Orders (Guest or Auth)
	v1.POST("/orders", middleware.OptionalAuthMiddleware(), controllers.CreateOrder)
	v1.GET("/orders/:order_number", middleware.OptionalAuthMiddleware(), controllers.GetOrderDetail)
	v1.GET("/orders/:order_number/tickets", middleware.OptionalAuthMiddleware(), controllers.GetOrderTickets)
	v1.POST("/orders/:order_number/cancel", controllers.UserCancelOrder)

	userOrders.Use(middleware.AuthMiddleware())
	{
		userOrders.GET("", controllers.GetUserOrders)
		userOrders.POST("/:order_number/pay", controllers.PayOrder)
		// order.go PayOrder: `id := c.Param("id")`. Queries `Where("id = ? AND user_id = ?", id, userID)`.
		// If we want consistency, maybe keep :id here if it's internal ID.
		// Spec says GET /orders/{order_number}.
		// Does PayOrder exist in spec? `POST /api/v1/orders/{id}/pay` is not in the provided Cheatsheet (only SimulatePayment).
		// But in routes.go it was `userOrders.POST("/:id/pay", controllers.PayOrder)`.
		// Let's assume PayOrder uses ID.
	}

	// Admin & Organizer Shared Routes
	admin := v1.Group("/admin")
	admin.Use(middleware.AuthMiddleware(), requireAdminOrOrganizer())
	{
		// Events (Scoped)
		admin.GET("/events", controllers.AdminGetEvents)
		admin.POST("/events", controllers.CreateEvent)
		admin.GET("/events/:id", controllers.GetEventDetail)
		admin.PUT("/events/:id", controllers.UpdateEvent)
		admin.DELETE("/events/:id", controllers.DeleteEvent)
		admin.PATCH("/events/:id/status", controllers.UpdateEventStatus)
		admin.GET("/events/:id/analytics", controllers.GetEventAnalytics)

		// Ticket Types (Scoped)
		admin.GET("/ticket-types", controllers.AdminGetTicketTypes)
		admin.POST("/ticket-types", controllers.CreateTicketType)
		admin.GET("/ticket-types/:id", controllers.GetTicketTypeDetail)
		admin.PUT("/ticket-types/:id", controllers.UpdateTicketType)
		admin.DELETE("/ticket-types/:id", controllers.DeleteTicketType)
		admin.PATCH("/ticket-types/:id/status", controllers.UpdateTicketTypeStatus)

		// Dashboard (Scoped)
		admin.GET("/stats", controllers.AdminGetStats)
		admin.GET("/dashboard/revenue", controllers.AdminGetRevenueChart)
		admin.GET("/dashboard/transactions-overview", controllers.GetTransactionsOverview)
		admin.GET("/dashboard/events-overview", controllers.GetEventsOverview)
		admin.GET("/dashboard/users-overview", controllers.GetUsersOverview) // Organizer probably shouldn't see ALL users, but maybe their customers? Let's leave for now or restricted?
		// Wait, Users Overview might leak? Let's assume stats are fine if scoped.
		// Actually, let's keep Users Overview for Admin only.

		// Transactions (Scoped)
		admin.GET("/transactions", controllers.AdminGetTransactions)
		admin.GET("/transactions/export", controllers.ExportTransactions)
		admin.GET("/transactions/revenue-summary", controllers.GetRevenueSummary)
		admin.GET("/transactions/:id", controllers.AdminGetTransactionDetail)
		admin.POST("/transactions/:id/resend-email", controllers.ResendTicketEmail)
		admin.POST("/transactions/:id/cancel", controllers.CancelTransaction)
		admin.POST("/transactions/:id/mark-paid", controllers.MarkTransactionPaid)
		admin.PUT("/transactions/:id/status", controllers.UpdateTransactionStatus)
		admin.GET("/transactions/:id/timeline", controllers.GetTransactionTimeline)
		// admin.POST("/transactions/trigger-scraping", controllers.AdminTriggerScraping) // Scraping is system level

		// Upload
		admin.POST("/upload", controllers.UploadFile)
	}

	// Super Admin Only Routes
	superAdmin := v1.Group("/admin")
	superAdmin.Use(middleware.AuthMiddleware(), requireAdmin())
	{
		// System Actions
		superAdmin.POST("/transactions/trigger-scraping", controllers.AdminTriggerScraping)

		// Categories
		superAdmin.GET("/categories", controllers.AdminGetCategories)
		superAdmin.POST("/categories", controllers.CreateCategory)
		superAdmin.GET("/categories/:id", controllers.AdminGetCategoryDetail)
		superAdmin.PUT("/categories/:id", controllers.UpdateCategory)
		superAdmin.DELETE("/categories/:id", controllers.DeleteCategory)
		superAdmin.PATCH("/categories/:id/status", controllers.UpdateCategoryStatus)
		superAdmin.PUT("/categories/reorder", controllers.ReorderCategories)

		// Users
		superAdmin.GET("/users", controllers.AdminGetUsers)
		superAdmin.POST("/users", controllers.AdminCreateUser)
		superAdmin.GET("/users/:id", controllers.AdminGetUserDetail)
		superAdmin.PUT("/users/:id", controllers.AdminUpdateUser)
		superAdmin.DELETE("/users/:id", controllers.AdminDeleteUser)
		superAdmin.PATCH("/users/:id/role", controllers.AdminUpdateUserRole)
		superAdmin.PATCH("/users/:id/status", controllers.AdminUpdateUserStatus)
		superAdmin.GET("/users/:id/activity", controllers.AdminGetUserActivity)
		superAdmin.GET("/users/:id/transactions", controllers.AdminGetUserTransactions)

		// Reports (Global) - Unless scoped later
		superAdmin.GET("/reports/sales", controllers.AdminGetSalesReport)
		superAdmin.GET("/reports/events", controllers.AdminGetEventsReport)
		superAdmin.GET("/reports/users", controllers.GetUsersReport)
		superAdmin.GET("/reports/top-events", controllers.GetTopEventsReport)
		superAdmin.GET("/reports/export", controllers.AdminExportReport)

		// Site Settings
		superAdmin.PUT("/settings", controllers.UpdateSettings)
	}

	// Public Settings (Already outside)
	v1.GET("/settings", controllers.GetSettings)

	return r
}

func requireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("userRole")
		if !exists || role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "message": "Admin access required"})
			return
		}
		c.Next()
	}
}

func requireAdminOrOrganizer() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("userRole")
		if !exists || (role != "admin" && role != "organizer") {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"success": false, "message": "Admin or Organizer access required"})
			return
		}
		c.Next()
	}
}
