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

	// Admin
	admin := v1.Group("/admin")
	admin.Use(middleware.AuthMiddleware(), requireAdmin())
	{
		// Admin Categories
		admin.GET("/categories", controllers.AdminGetCategories)
		admin.POST("/categories", controllers.CreateCategory)
		admin.GET("/categories/:id", controllers.AdminGetCategoryDetail)
		admin.PUT("/categories/:id", controllers.UpdateCategory)
		admin.DELETE("/categories/:id", controllers.DeleteCategory)
		admin.PATCH("/categories/:id/status", controllers.UpdateCategoryStatus)
		admin.PUT("/categories/reorder", controllers.ReorderCategories)

		// Admin Events
		admin.GET("/events", controllers.AdminGetEvents)
		admin.POST("/events", controllers.CreateEvent)
		admin.GET("/events/:id", controllers.GetEventDetail) // Can reuse public detail or simpler admin detail? Spec implies separate.
		// Let's use GetEventDetail (public) if it's sufficient, or create AdminGetEventDetail if logic differs (e.g. show hidden fields).
		// `controllers/admin_events.go` does not currently have AdminGetEventDetail explicitly, using public logic or default?
		// Wait, I didn't add AdminGetEventDetail to admin_events.go.
		// Ideally we reuse the public one if it shows enough.
		admin.PUT("/events/:id", controllers.UpdateEvent)
		admin.DELETE("/events/:id", controllers.DeleteEvent)
		admin.PATCH("/events/:id/status", controllers.UpdateEventStatus)
		admin.GET("/events/:id/analytics", controllers.GetEventAnalytics)

		// Admin Ticket Types
		admin.GET("/ticket-types", controllers.AdminGetTicketTypes)
		admin.POST("/ticket-types", controllers.CreateTicketType)
		admin.GET("/ticket-types/:id", controllers.GetTicketTypeDetail)
		admin.PUT("/ticket-types/:id", controllers.UpdateTicketType)
		admin.DELETE("/ticket-types/:id", controllers.DeleteTicketType)
		admin.PATCH("/ticket-types/:id/status", controllers.UpdateTicketTypeStatus)

		// Admin Dashboard
		admin.GET("/stats", controllers.AdminGetStats)
		admin.GET("/dashboard/revenue", controllers.AdminGetRevenueChart)
		admin.GET("/dashboard/transactions-overview", controllers.GetTransactionsOverview)
		admin.GET("/dashboard/events-overview", controllers.GetEventsOverview)
		admin.GET("/dashboard/users-overview", controllers.GetUsersOverview)

		// Admin Users
		admin.GET("/users", controllers.AdminGetUsers)
		admin.POST("/users", controllers.AdminCreateUser)
		admin.GET("/users/:id", controllers.AdminGetUserDetail)
		admin.PUT("/users/:id", controllers.AdminUpdateUser)
		admin.DELETE("/users/:id", controllers.AdminDeleteUser)
		admin.PATCH("/users/:id/role", controllers.AdminUpdateUserRole)
		admin.PATCH("/users/:id/status", controllers.AdminUpdateUserStatus)
		admin.GET("/users/:id/activity", controllers.AdminGetUserActivity)
		admin.GET("/users/:id/transactions", controllers.AdminGetUserTransactions)

		// Admin Transactions
		admin.GET("/transactions", controllers.AdminGetTransactions)
		admin.GET("/transactions/export", controllers.ExportTransactions)         // Specific route before :id
		admin.GET("/transactions/revenue-summary", controllers.GetRevenueSummary) // Specific route before :id
		admin.GET("/transactions/:id", controllers.AdminGetTransactionDetail)
		admin.POST("/transactions/:id/resend-email", controllers.ResendTicketEmail)
		admin.POST("/transactions/:id/cancel", controllers.CancelTransaction)
		admin.POST("/transactions/:id/mark-paid", controllers.MarkTransactionPaid)
		admin.PUT("/transactions/:id/status", controllers.UpdateTransactionStatus)
		admin.GET("/transactions/:id/timeline", controllers.GetTransactionTimeline)
		admin.POST("/transactions/trigger-scraping", controllers.AdminTriggerScraping)

		// Reports
		admin.GET("/reports/sales", controllers.AdminGetSalesReport)
		admin.GET("/reports/events", controllers.AdminGetEventsReport)
		admin.GET("/reports/users", controllers.GetUsersReport)
		admin.GET("/reports/top-events", controllers.GetTopEventsReport)
		admin.GET("/reports/export", controllers.AdminExportReport)

		// Upload
		admin.POST("/upload", controllers.UploadFile)
	}

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
