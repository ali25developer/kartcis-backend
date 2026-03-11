package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

type FlipBillRequest struct {
	Title                 string `json:"title"`
	Amount                int    `json:"amount"`
	Type                  string `json:"type"` // SINGLE
	SenderName            string `json:"sender_name"`
	SenderEmail           string `json:"sender_email"`
	SenderPhoneNumber     string `json:"sender_phone_number"`
	SenderAddress         string `json:"sender_address"`
	IsAddressRequired     int    `json:"is_address_required"`
	IsPhoneNumberRequired int    `json:"is_phone_number_required"`
	RedirectURL           string `json:"redirect_url,omitempty"`
}

type FlipBillResponse struct {
	ID          int    `json:"link_id"`
	BillID      int    `json:"bill_id"`
	ExternalID  string `json:"external_id"`
	Title       string `json:"title"`
	SenderName  string `json:"sender_name"`
	SenderEmail string `json:"sender_email"`
	Amount      int    `json:"amount"`
	Status      string `json:"status"`   // PENDING, SUCCESSFUL, CANCELLED
	PaymentURL  string `json:"link_url"` // v2 uses link_url
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func CreateFlipBill(orderID string, amount int, name, email, phone, redirectURL string, expiredAt *time.Time) (*FlipBillResponse, error) {
	apiKey := os.Getenv("FLIP_API_KEY")
	baseURL := os.Getenv("FLIP_BASE_URL")
	if baseURL == "" {
		baseURL = "https://bigflip.id/api/v2" // Reverted to v2 for Sandbox
	}

	data := url.Values{}
	data.Set("title", fmt.Sprintf("Pembayaran Order %s", orderID))
	data.Set("amount", fmt.Sprintf("%d", amount))
	data.Set("type", "SINGLE")
	data.Set("sender_name", name)
	data.Set("sender_email", email)
	data.Set("sender_phone_number", phone)
	data.Set("is_address_required", "0")
	data.Set("is_phone_number_required", "0")
	data.Set("step", "2") // 2 = Skip identity screen (requires x-www-form-urlencoded)

	if redirectURL != "" {
		data.Set("redirect_url", redirectURL)
	}

	if expiredAt != nil {
		// Flip v2 Date format: YYYY-MM-DD HH:mm+0700
		data.Set("expired_date", expiredAt.Format("2006-01-02 15:04-0700"))
	}

	req, err := http.NewRequest("POST", baseURL+"/pwf/bill", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(apiKey, "")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Debug Logs
	log.Printf("[Flip-Debug] Request URL: %s/pwf/bill", baseURL)
	log.Printf("[Flip-Debug] Authorization: Basic (API Key present)") // Log presence of API key, not the key itself
	log.Printf("[Flip-Debug] Payload: %s", data.Encode())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[Flip-Debug] Connection Error: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	log.Printf("[Flip-Debug] Response Status: %d", resp.StatusCode)
	log.Printf("[Flip-Debug] Response Body: %s", string(body))

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("flip api error: %s (status %d)", string(body), resp.StatusCode)
	}

	var flipResp FlipBillResponse
	if err := json.Unmarshal(body, &flipResp); err != nil {
		return nil, err
	}

	return &flipResp, nil
}
