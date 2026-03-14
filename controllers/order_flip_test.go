package controllers

// NOTE: Tests for Flip callback integration are temporarily disabled.
// Uncomment when Flip Production API is ready and PaymentCallback handler is re-enabled.

// import (
// 	"bytes"
// 	"encoding/json"
// 	"kartcis-backend/config"
// 	"kartcis-backend/models"
// 	"net/http"
// 	"net/http/httptest"
// 	"os"
// 	"testing"
// 	"time"
//
// 	"github.com/gin-gonic/gin"
// 	"github.com/stretchr/testify/assert"
// 	"gorm.io/driver/sqlite"
// 	"gorm.io/gorm"
// )
//
// func setupTestDB() {
// 	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
// 	config.DB = db
// 	db.AutoMigrate(
// 		&models.Order{},
// 		&models.Ticket{},
// 		&models.TicketType{},
// 		&models.Event{},
// 		&models.OrderStatusHistory{},
// 	)
// }
//
// func TestFlipCallback_Scenarios(t *testing.T) {
// 	setupTestDB()
// 	gin.SetMode(gin.TestMode)
//
// 	os.Setenv("FLIP_WEBHOOK_TOKEN", "test-token-123")
// 	defer os.Unsetenv("FLIP_WEBHOOK_TOKEN")
//
// 	// Buat dummy order dengan payment_data = bill_link_id (sesuai cara lookup di PaymentCallback)
// 	order := models.Order{
// 		OrderNumber:   "ORD-FLIP-TEST",
// 		TotalAmount:   50000,
// 		Status:        "pending",
// 		PaymentMethod: "FLIP",
// 		PaymentData:   "99001", // Ini adalah bill_link_id yang akan dikirim Flip
// 		CreatedAt:     time.Now(),
// 	}
// 	config.DB.Create(&order)
//
// 	r := gin.Default()
// 	r.POST("/callback", PaymentCallback)
//
// 	// Helper: buat request form-encoded seperti Flip sungguhan
// 	makeFlipCallbackRequest := func(billLinkID int64, status, token string) *http.Request {
// 		billData, _ := json.Marshal(map[string]interface{}{
// 			"bill_link_id": billLinkID,
// 			"amount":       "50000",
// 			"status":       status,
// 		})
// 		// Flip mengirim sebagai application/x-www-form-urlencoded
// 		body := "data=" + string(billData) + "&token=" + token
// 		req, _ := http.NewRequest("POST", "/callback", bytes.NewBufferString(body))
// 		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
// 		return req
// 	}
//
// 	t.Run("Scenario 1: SUCCESSFUL Payment", func(t *testing.T) {
// 		req := makeFlipCallbackRequest(99001, "SUCCESSFUL", "test-token-123")
// 		w := httptest.NewRecorder()
// 		r.ServeHTTP(w, req)
//
// 		assert.Equal(t, http.StatusOK, w.Code)
//
// 		var updatedOrder models.Order
// 		config.DB.Where("order_number = ?", "ORD-FLIP-TEST").First(&updatedOrder)
// 		assert.Equal(t, "paid", updatedOrder.Status)
// 		assert.NotNil(t, updatedOrder.PaidAt)
// 	})
//
// 	t.Run("Scenario 2: CANCELLED Payment", func(t *testing.T) {
// 		// Reset status ke pending untuk test berikutnya
// 		config.DB.Model(&order).Update("status", "pending")
//
// 		req := makeFlipCallbackRequest(99001, "CANCELLED", "test-token-123")
// 		w := httptest.NewRecorder()
// 		r.ServeHTTP(w, req)
//
// 		assert.Equal(t, http.StatusOK, w.Code)
//
// 		var updatedOrder models.Order
// 		config.DB.Where("order_number = ?", "ORD-FLIP-TEST").First(&updatedOrder)
// 		assert.Equal(t, "cancelled", updatedOrder.Status)
// 	})
//
// 	t.Run("Scenario 3: Invalid Token (Security Test)", func(t *testing.T) {
// 		payload := map[string]interface{}{
// 			"data":  "{}",
// 			"token": "WRONG-TOKEN",
// 		}
// 		body, _ := json.Marshal(payload)
//
// 		req, _ := http.NewRequest("POST", "/callback", bytes.NewBuffer(body))
// 		req.Header.Set("X-Callback-Token", "WRONG-TOKEN")
//
// 		w := httptest.NewRecorder()
// 		r.ServeHTTP(w, req)
//
// 		// It should fail or return 400
// 		assert.Equal(t, http.StatusBadRequest, w.Code)
// 	})
// }
