package jobs

import (
	"io"
	"kartcis-backend/config"
	"kartcis-backend/models"
	"kartcis-backend/utils"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
)

func StartPaymentCheckerJob() {
	// Run every 2 minutes
	ticker := time.NewTicker(2 * time.Minute)
	go func() {
		for range ticker.C {
			CheckBankJagoEmails()
		}
	}()
}

func CheckBankJagoEmails() {
	host := os.Getenv("IMAP_HOST")
	port := os.Getenv("IMAP_PORT")
	user := os.Getenv("IMAP_USER")
	pass := os.Getenv("IMAP_PASS")

	if host == "" || user == "" || pass == "" {
		// log.Println("[PaymentJob] IMAP not configured, skipping...")
		return
	}

	// Connect to server
	c, err := client.DialTLS(host+":"+port, nil)
	if err != nil {
		log.Println("[PaymentJob] Dial error:", err)
		return
	}
	defer c.Logout()

	// Login
	if err := c.Login(user, pass); err != nil {
		log.Println("[PaymentJob] Login error:", err)
		return
	}

	// Select INBOX
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Println("[PaymentJob] Select error:", err)
		return
	}

	if mbox.Messages == 0 {
		return
	}

	criteria := imap.NewSearchCriteria()
	criteria.Since = time.Now().Add(-24 * time.Hour)
	// Broaden to anything from jago.com to handle no-reply or noreply
	criteria.Header.Set("From", "jago.com")

	log.Println("[PaymentJob] Checking for new emails (Manual/Auto)...")
	ids, err := c.Search(criteria)
	log.Printf("[PaymentJob] Found %d matching emails in last 24h\n", len(ids))
	if err != nil {
		log.Println("[PaymentJob] Search error:", err)
		return
	}

	if len(ids) == 0 {
		return
	}

	seqset := new(imap.SeqSet)
	seqset.AddNum(ids...)

	// Get the whole message body
	var section imap.BodySectionName
	items := []imap.FetchItem{section.FetchItem(), imap.FetchEnvelope}

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, items, messages)
	}()

	for msg := range messages {
		log.Printf("[PaymentJob] Processing Email Subject: %s\n", msg.Envelope.Subject)
		r := msg.GetBody(&section)
		if r == nil {
			continue
		}

		// Create a new mail reader
		mr, err := mail.CreateReader(r)
		if err != nil {
			continue
		}

		// Read parts
		var body string
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			} else if err != nil {
				break
			}

			switch h := p.Header.(type) {
			case *mail.InlineHeader:
				contentType, _, _ := h.ContentType()
				if contentType == "text/plain" || contentType == "text/html" {
					b, _ := io.ReadAll(p.Body)
					body = string(b)
				}
			}
		}

		if body != "" {
			processJagoEmail(body)
		}
	}

	if err := <-done; err != nil {
		log.Println("[PaymentJob] Fetch error:", err)
	}
}

func processJagoEmail(body string) {
	// Sample Jago Body: "Kamu dapet transferan nih! ... Nominal: Rp 50.412 ... Pengirim: JOHN DOE"
	// Regex to find Amount
	re := regexp.MustCompile(`Rp\s?([0-9.]+)`)
	match := re.FindStringSubmatch(body)
	if len(match) < 2 {
		return
	}

	amountStr := strings.ReplaceAll(match[1], ".", "")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return
	}

	log.Printf("[PaymentJob] Found Jago Transfer: Rp %v\n", amount)

	// Search for pending order with this exact amount
	var order models.Order
	// Using UniqueCode and TotalAmount as double check
	err = config.DB.Where("status = ? AND total_amount = ?", "pending", amount).First(&order).Error
	if err != nil {
		// No matching order
		return
	}

	// Double check if it's a Bank Transfer Jago payment
	if !strings.Contains(order.PaymentMethod, "BANK_TRANSFER_JAGO") {
		return
	}

	// Mark as Paid
	tx := config.DB.Begin()
	now := time.Now()
	order.Status = "paid"
	order.PaidAt = &now
	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		return
	}

	// Record history
	tx.Create(&models.OrderStatusHistory{
		OrderID:   order.ID,
		Status:    "paid",
		Notes:     "Verified automatically via Bank Jago Email",
		CreatedAt: now,
	})

	tx.Commit()

	log.Printf("[PaymentJob] Order %s marked as PAID via Email Verification\n", order.OrderNumber)

	// Send Ticket
	var tickets []models.Ticket
	config.DB.Preload("Event").Preload("TicketType").Where("order_id = ?", order.ID).Find(&tickets)
	utils.SendTicketEmail(order, tickets)
}
