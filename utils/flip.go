package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
	IsAddressRequired     bool   `json:"is_address_required"`
	IsPhoneNumberRequired bool   `json:"is_phone_number_required"`
	RedirectURL           string `json:"redirect_url,omitempty"`
	Step                  string `json:"step"` // checkout, checkout_seamless, direct_api
}

type FlipBillResponse struct {
	ID          int    `json:"link_id"`
	BillID      int    `json:"bill_id"`
	ExternalID  string `json:"external_id"`
	Title       string `json:"title"`
	SenderName  string `json:"sender_name"`
	SenderEmail string `json:"sender_email"`
	Amount      int    `json:"amount"`
	Status      string `json:"status"`      // PENDING, SUCCESSFUL, CANCELLED
	PaymentURL  string `json:"payment_url"` // v3 uses payment_url
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func CreateFlipBill(orderID string, amount int, name, email, phone, redirectURL string) (*FlipBillResponse, error) {
	apiKey := os.Getenv("FLIP_API_KEY")
	baseURL := os.Getenv("FLIP_BASE_URL")
	if baseURL == "" {
		baseURL = "https://bigflip.id/api/v3" // Upgraded to v3
	}

	payload := FlipBillRequest{
		Title:                 fmt.Sprintf("Pembayaran Order %s", orderID),
		Amount:                amount,
		Type:                  "SINGLE",
		SenderName:            name,
		SenderEmail:           email,
		SenderPhoneNumber:     phone,
		IsAddressRequired:     false,
		IsPhoneNumberRequired: false,
		RedirectURL:           redirectURL,
		Step:                  "checkout_seamless",
	}

	jsonPayload, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", baseURL+"/pwf/bill", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(apiKey, "")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("flip api error: %s (status %d)", string(body), resp.StatusCode)
	}

	var flipResp FlipBillResponse
	if err := json.Unmarshal(body, &flipResp); err != nil {
		return nil, err
	}

	return &flipResp, nil
}
