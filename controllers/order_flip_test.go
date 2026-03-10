package controllers

import (
	"bytes"
	"encoding/json"
	"kartcis-backend/config"
	"kartcis-backend/models"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() {
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	config.DB = db
	db.AutoMigrate(
		&models.Order{},
		&models.Ticket{},
		&models.TicketType{},
		&models.Event{},
		&models.OrderStatusHistory{},
	)
}

func TestFlipCallback_Scenarios(t *testing.T) {
	setupTestDB()
	gin.SetMode(gin.TestMode)

	// Set environment variables for testing
	os.Setenv("FLIP_WEBHOOK_TOKEN", "test-token-123")

	// 1. Create a dummy order
	order := models.Order{
		OrderNumber:   "ORD-FLIP-TEST",
		TotalAmount:   50000,
		Status:        "pending",
		PaymentMethod: "FLIP",
		CreatedAt:     time.Now(),
	}
	config.DB.Create(&order)

	r := gin.Default()
	r.POST("/callback", PaymentCallback)

	t.Run("Scenario 1: SUCCESSFUL Payment", func(t *testing.T) {
		billData := map[string]interface{}{
			"external_id": "ORD-FLIP-TEST",
			"status":      "SUCCESSFUL",
		}
		billJSON, _ := json.Marshal(billData)
		payload := map[string]interface{}{
			"data":  string(billJSON),
			"token": "test-token-123",
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/callback", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Callback-Token", "test-token-123")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var updatedOrder models.Order
		config.DB.Where("order_number = ?", "ORD-FLIP-TEST").First(&updatedOrder)
		assert.Equal(t, "paid", updatedOrder.Status)
		assert.NotNil(t, updatedOrder.PaidAt)

		// Check History
		var history models.OrderStatusHistory
		config.DB.Where("order_id = ? AND status = ?", updatedOrder.ID, "success").First(&history)
		assert.NotNil(t, history.ID)
	})

	t.Run("Scenario 2: CANCELLED Payment", func(t *testing.T) {
		// Reset status to pending for next test
		config.DB.Model(&order).Update("status", "pending")

		billData := map[string]interface{}{
			"external_id": "ORD-FLIP-TEST",
			"status":      "CANCELLED",
		}
		billJSON, _ := json.Marshal(billData)
		payload := map[string]interface{}{
			"data":  string(billJSON),
			"token": "test-token-123",
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/callback", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var updatedOrder models.Order
		config.DB.Where("order_number = ?", "ORD-FLIP-TEST").First(&updatedOrder)
		assert.Equal(t, "cancelled", updatedOrder.Status)
	})

	t.Run("Scenario 3: Invalid Token (Security Test)", func(t *testing.T) {
		payload := map[string]interface{}{
			"data":  "{}",
			"token": "WRONG-TOKEN",
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", "/callback", bytes.NewBuffer(body))
		req.Header.Set("X-Callback-Token", "WRONG-TOKEN")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// It should fail or return 400
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
