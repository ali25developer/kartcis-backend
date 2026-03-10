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
	Type                  string `json:"type"`           // SINGLE
	Step                  int    `json:"step,omitempty"` // v2: 1, 2, 3
	SenderName            string `json:"sender_name"`
	SenderEmail           string `json:"sender_email"`
	SenderPhoneNumber     string `json:"sender_phone_number"`
	SenderAddress         string `json:"sender_address"`
	IsAddressRequired     int    `json:"is_address_required"`
	IsPhoneNumberRequired int    `json:"is_phone_number_required"`
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

func CreateFlipBill(orderID string, amount int, name, email, phone string) (*FlipBillResponse, error) {
	apiKey := os.Getenv("FLIP_API_KEY")
	baseURL := os.Getenv("FLIP_BASE_URL")
	if baseURL == "" {
		baseURL = "https://bigflip.id/api/v2" // Reverted to v2
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
