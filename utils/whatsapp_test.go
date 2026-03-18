package utils

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

func TestSendWA_Manual(t *testing.T) {
	// Load .env to get DATABASE_URL
	_ = godotenv.Load("../.env")

	// Override DATABASE_URL if running outside docker but db is on localhost
	// If the user is running this on the host machine, 'db' might not be reachable.
	// We'll try to use the env value exactly as is first.

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" || strings.Contains(dbURL, "host=db") {
		// If using 'db' as host (docker), try 127.0.0.1 for host machine testing
		dbURL = "host=127.0.0.1 user=postgres password=password dbname=kartcis port=5432 sslmode=disable"
		os.Setenv("DATABASE_URL", dbURL)
	}

	fmt.Printf("--- WhatsApp Test Debug ---\n")
	fmt.Printf("DB URL: %s\n", dbURL)
	InitWA()

	// Wait for connection (up to 20 seconds)
	connected := false
	for i := 0; i < 10; i++ {
		if WAClient != nil && WAClient.IsConnected() && WAClient.IsLoggedIn() {
			connected = true
			break
		}
		fmt.Printf("Waiting for WA connection... (%d/10)\n", i+1)
		time.Sleep(2 * time.Second)
	}

	if !connected {
		t.Skip("WhatsApp not connected or logged in. Please ensure the session is active.")
		return
	}

	// Verify sender if possible (optional)
	if WAClient.Store.ID != nil {
		fmt.Printf("Connected as: %s\n", WAClient.Store.ID.String())
	}

	recipients := []string{"083127246830", "081248804671"}
	message := "beli kartcis ke event ini https://kartcis.id/event/8"

	fmt.Println("Starting to send messages...")
	for _, phone := range recipients {
		err := SendWAMessage(phone, message)
		if err != nil {
			fmt.Printf("❌ Gagal kirim ke %s: %v\n", phone, err)
			t.Errorf("Gagal kirim ke %s: %v", phone, err)
		} else {
			fmt.Printf("✅ Berhasil kirim ke %s\n", phone)
		}
		// A bit of delay to avoid rate limiting
		time.Sleep(3 * time.Second)
	}
}
