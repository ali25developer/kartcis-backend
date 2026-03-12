package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"kartcis-backend/config"
	"kartcis-backend/models"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestCreateOrder_WithUserPayload(t *testing.T) {
	// Setup Test DB
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}
	config.DB = db
	db.AutoMigrate(&models.Order{}, &models.Ticket{}, &models.TicketType{}, &models.Event{}, &models.OrderStatusHistory{})

	// Seed Data
	event := models.Event{ID: 12, Title: "event test", Status: "published"}
	config.DB.Create(&event)
	ticketType := models.TicketType{ID: 14, EventID: 12, Name: "baru", Price: 1000, Available: 10}
	config.DB.Create(&ticketType)

	// Mock Flip Server
	flipServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"link_id":123,"bill_id":456,"link_url":"https://flip.id/mock"}`)
	}))
	defer flipServer.Close()

	os.Setenv("FLIP_API_KEY", "test_key")
	os.Setenv("FLIP_BASE_URL", flipServer.URL)
	os.Setenv("FRONTEND_URL", "http://localhost:3000")
	os.Setenv("SMTP_HOST", "") // Ensure it's empty so mailer skips

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/orders", CreateOrder)

	// User Payload
	payload := `{
		"items": [
			{
				"ticket_type_id": 14,
				"quantity": 1,
				"attendees": [
					{
						"name": "Ali Rohmansyah",
						"email": "ali25developer@gmail.com",
						"phone": "081234567891"
					}
				]
			}
		],
		"payment_method": "FLIP",
		"voucher_code": "",
		"customer_info": {
			"name": "Ali Rohmansyah",
			"email": "ali25developer@gmail.com",
			"phone": "081234567891"
		}
	}`

	req, _ := http.NewRequest("POST", "/orders", bytes.NewBuffer([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.True(t, response["success"].(bool))

	data := response["data"].(map[string]interface{})
	assert.Equal(t, "https://flip.id/mock", data["payment_url"])
	assert.Equal(t, "123", data["payment_data"]) // payment_data menyimpan link_id sebagai string
}

func TestCreateOrder_FlipFailure(t *testing.T) {
	// Setup Test DB
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	config.DB = db
	db.AutoMigrate(&models.Order{}, &models.Ticket{}, &models.TicketType{}, &models.Event{}, &models.OrderStatusHistory{})

	// Seed Data
	event := models.Event{ID: 15, Title: "Failure Test", Status: "published"}
	config.DB.Create(&event)
	ticketType := models.TicketType{ID: 16, EventID: 15, Name: "fail", Price: 1000, Available: 10}
	config.DB.Create(&ticketType)

	// Mock Flip Server returning 422
	flipServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, `{"code":"VALIDATION_ERROR","errors":[{"attribute":"step","code":1079,"message":"Param step is invalid"}]}`)
	}))
	defer flipServer.Close()

	os.Setenv("FLIP_API_KEY", "test_key")
	os.Setenv("FLIP_BASE_URL", flipServer.URL)
	os.Setenv("FRONTEND_URL", "http://localhost:3000")
	os.Setenv("SMTP_HOST", "")

	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/orders", CreateOrder)

	payload := `{
		"items": [{"ticket_type_id": 16, "quantity": 1}],
		"payment_method": "FLIP",
		"customer_info": {"name": "Fail User", "email": "fail@test.com", "phone": "081"}
	}`

	req, _ := http.NewRequest("POST", "/orders", bytes.NewBuffer([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// In current code, Flip error returns 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.False(t, response["success"].(bool))
	assert.Contains(t, response["message"].(string), "Flip API Error")
}
