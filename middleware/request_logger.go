package middleware

import (
	"kartcis-backend/config"
	"kartcis-backend/models"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func SmartLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process Request
		c.Next()

		// AFTER REQUEST
		end := time.Now()
		latency := end.Sub(start).Milliseconds()
		path := c.Request.URL.Path
		method := c.Request.Method
		status := c.Writer.Status()

		// --- SMART FILTER LOGIC ---

		// 1. Skip non-API routes (static files, health checks)
		if !strings.HasPrefix(path, "/api") {
			return
		}

		// 2. Skip OPTIONS requests
		if method == "OPTIONS" {
			return
		}

		// 3. Skip GET requests for public API (unless it's an error)
		//    We want to log ALL Admin actions (starts with /api/v1/admin)
		isAdminRoute := strings.Contains(path, "/admin/")
		isError := status >= 400
		isMutation := method == "POST" || method == "PUT" || method == "DELETE" || method == "PATCH"

		shouldLog := isAdminRoute || isError || isMutation

		if !shouldLog {
			return
		}

		// 4. Capture User ID if available
		var userID *uint
		if val, exists := c.Get("userID"); exists {
			if id, ok := val.(uint); ok {
				userID = &id
			}
		}

		// 5. Async Log to DB (Don't block response)
		go func(uid *uint, m, p, ip, ua string, s int, l int64) {
			config.DB.Create(&models.RequestLog{
				UserID:    uid,
				Method:    m,
				Path:      p,
				IPAddress: ip,
				UserAgent: ua,
				Status:    s,
				Latency:   l,
				CreatedAt: time.Now(),
			})
		}(userID, method, path, c.ClientIP(), c.Request.UserAgent(), status, latency)
	}
}
