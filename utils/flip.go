package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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
	Step                  int    `json:"step"` // v2 requires int: 1, 2, or 3
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

func CreateFlipBill(orderID string, amount int, name, email, phone, redirectURL string) (*FlipBillResponse, error) {
	apiKey := os.Getenv("FLIP_API_KEY")
	baseURL := os.Getenv("FLIP_BASE_URL")
	if baseURL == "" {
		baseURL = "https://bigflip.id/api/v2" // Reverted to v2 for Sandbox
	}

	payload := FlipBillRequest{
		Title:                 fmt.Sprintf("Pembayaran Order %s", orderID),
		Amount:                amount,
		Type:                  "SINGLE",
		SenderName:            name,
		SenderEmail:           email,
		SenderPhoneNumber:     phone,
		IsAddressRequired:     0,
		IsPhoneNumberRequired: 0,
		RedirectURL:           redirectURL,
		Step:                  1, // v2: 1=input, 2=method, 3=confirmation
	}

	jsonPayload, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", baseURL+"/pwf/bill", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(apiKey, "")
	req.Header.Set("Content-Type", "application/json")

	// Debug Logs
	log.Printf("[Flip-Debug] Request URL: %s/pwf/bill", baseURL)
	log.Printf("[Flip-Debug] Authorization: Basic (API Key present)") // Log presence of API key, not the key itself
	log.Printf("[Flip-Debug] Payload: %s", string(jsonPayload))

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
