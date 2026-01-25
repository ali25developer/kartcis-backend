package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"kartcis-backend/config"
	"kartcis-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Authorization header is required"})
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			secret = "secret123"
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Invalid token"})
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Check expiry
			if float64(time.Now().Unix()) > claims["exp"].(float64) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Token expired"})
				return
			}

			userID := uint(claims["sub"].(float64))

			// Optional: Check if user still exists in DB
			var user models.User
			if result := config.DB.First(&user, userID); result.Error != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "message": "User not found"})
				return
			}

			c.Set("userID", userID)
			c.Set("userRole", user.Role)

			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Invalid token claims"})
		}
	}
}

// OptionalAuthMiddleware attempts to parse token but continues if missing/invalid
func OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			secret = "secret123"
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			// Ignore invalid token, treat as guest
			c.Next()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Check expiry
			if float64(time.Now().Unix()) > claims["exp"].(float64) {
				// Expired token -> guest
				c.Next()
				return
			}

			userID := uint(claims["sub"].(float64))

			// Optional: Check if user exists
			var user models.User
			if result := config.DB.First(&user, userID); result.Error == nil {
				c.Set("userID", userID)
				c.Set("userRole", user.Role)
			}
		}

		c.Next()
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
