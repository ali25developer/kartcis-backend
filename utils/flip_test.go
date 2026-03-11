package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateFlipBill_Detailed(t *testing.T) {
	// Mock Flip API Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify URL Path
		assert.Equal(t, "/pwf/bill", r.URL.Path)
		assert.Equal(t, "POST", r.Method)

		// Verify Basic Auth
		username, password, ok := r.BasicAuth()
		assert.True(t, ok)
		assert.Equal(t, "test_api_key", username)
		assert.Equal(t, "", password)

		// Verify Content-Type
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Decode and verify payload
		var payload FlipBillRequest
		err := json.NewDecoder(r.Body).Decode(&payload)
		assert.NoError(t, err)

		// V2 specific checks
		assert.Equal(t, 3, payload.Step)
		assert.Equal(t, "SINGLE", payload.Type)
		assert.Equal(t, 0, payload.IsAddressRequired)
		assert.Equal(t, 0, payload.IsPhoneNumberRequired)
		assert.Contains(t, payload.Title, "ORD-123")

		// Return Mock Response
		resp := FlipBillResponse{
			ID:          12345,
			BillID:      67890,
			ExternalID:  "ORD-123",
			Title:       payload.Title,
			Status:      "PENDING",
			PaymentURL:  "https://flip.id/p/mock-link", // This field in JSON should be link_url
			CreatedAt:   "2026-03-11 09:00",
		}
		w.WriteHeader(http.StatusOK)
		// Custom JSON encoding to send link_url (v2) instead of payment_url (v3)
		jsonResp := map[string]interface{}{
			"link_id":     resp.ID,
			"bill_id":     resp.BillID,
			"external_id": resp.ExternalID,
			"title":       resp.Title,
			"status":      resp.Status,
			"link_url":    resp.PaymentURL,
			"created_at":  resp.CreatedAt,
		}
		json.NewEncoder(w).Encode(jsonResp)
	}))
	defer server.Close()

	// Setup Environment
	os.Setenv("FLIP_API_KEY", "test_api_key")
	os.Setenv("FLIP_BASE_URL", server.URL) // Redirect to mock server
	defer os.Unsetenv("FLIP_API_KEY")
	defer os.Unsetenv("FLIP_BASE_URL")

	// Call the function
	orderID := "ORD-123"
	amount := 50000
	name := "John Doe"
	email := "john@example.com"
	phone := "08123456789"
	redirectURL := "https://myapp.com/success"

	resp, err := CreateFlipBill(orderID, amount, name, email, phone, redirectURL)

	// Final assertions
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "https://flip.id/p/mock-link", resp.PaymentURL)
	assert.Equal(t, 12345, resp.ID)
	assert.Equal(t, 67890, resp.BillID)
}

func TestCreateFlipBill_ErrorHandling(t *testing.T) {
	// Mock Server returning 422 Validation Error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		fmt.Fprint(w, `{"code":"VALIDATION_ERROR","errors":[{"attribute":"step","code":1079,"message":"Param step is invalid"}]}`)
	}))
	defer server.Close()

	os.Setenv("FLIP_API_KEY", "test_api_key")
	os.Setenv("FLIP_BASE_URL", server.URL)
	defer os.Unsetenv("FLIP_API_KEY")
	defer os.Unsetenv("FLIP_BASE_URL")

	resp, err := CreateFlipBill("ORD-ERR", 1000, "User", "user@test.com", "081", "")

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "VALIDATION_ERROR")
	assert.Contains(t, err.Error(), "Param step is invalid")
}
