package jobs

import (
	"fmt"
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
	// Run every 1 minute
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			CheckBankJagoEmails("Auto")
		}
	}()
}

func CheckBankJagoEmails(source string) {
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
		log.Printf("[%s-PaymentJob] Dial error: %v\n", source, err)
		return
	}
	defer c.Logout()

	// Login
	if err := c.Login(user, pass); err != nil {
		log.Printf("[%s-PaymentJob] Login error: %v\n", source, err)
		return
	}

	// Select INBOX
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		log.Printf("[%s-PaymentJob] Select error: %v\n", source, err)
		return
	}

	if mbox.Messages == 0 {
		return
	}

	criteria := imap.NewSearchCriteria()
	criteria.Since = time.Now().Add(-24 * time.Hour)
	criteria.WithoutFlags = []string{imap.SeenFlag} // Hanya cari yang UNSEEN (belum dibaca)

	log.Printf("[%s-PaymentJob] Fetching unread emails from last 24h...", source)
	ids, err := c.Search(criteria)
	if err != nil {
		log.Println("[PaymentJob] Search error:", err)
		return
	}
	log.Printf("[%s-PaymentJob] Found %d total unseen emails in Inbox\n", source, len(ids))

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
		subject := strings.ToLower(msg.Envelope.Subject)
		from := ""
		if len(msg.Envelope.From) > 0 {
			from = strings.ToLower(msg.Envelope.From[0].Address())
		}

		log.Printf("[%s-PaymentJob] Scanner: %s (From: %s)\n", source, msg.Envelope.Subject, from)

		// Filter manual dalam Go agar tidak tergantung filter server IMAP yang kaku
		if !strings.Contains(subject, "sejumlah uang") && !strings.Contains(from, "jago.com") {
			continue
		}

		log.Printf("[%s-PaymentJob] MATCH! Processing: %s\n", source, msg.Envelope.Subject)
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
			ProcessJagoEmail(body, source, msg.Envelope.MessageId, msg.Envelope.Date)
		}
	}

	if err := <-done; err != nil {
		log.Printf("[%s-PaymentJob] Fetch error: %v\n", source, err)
	}
}

func ProcessJagoEmail(body string, source string, messageID string, emailDate time.Time) {
	// 1. Clean HTML tags if present
	body = stripHTML(body)

	// 1.1 Transactional check for message deduplication
	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var existingTx models.BankTransaction
	if tx.Where("reference_id = ?", messageID).First(&existingTx).Error == nil {
		tx.Rollback()
		return
	}

	// 2. Parse Email Body with very flexible Regex
	// Support: Nominal, Total, Jumlah, Amount
	// Support: Pengirim, Dari
	re := regexp.MustCompile(`(?i)(?:Nominal|Total|Jumlah|Jumlah\stransaksi)[:\s]*Rp\s?([0-9.]+)`)
	match := re.FindStringSubmatch(body)
	if len(match) < 2 {
		re = regexp.MustCompile(`(?i)Rp\s?([0-9.]+)`)
		match = re.FindStringSubmatch(body)
		if len(match) < 2 {
			tx.Rollback()
			return
		}
	}

	amountStr := strings.ReplaceAll(match[1], ".", "")
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		tx.Rollback()
		return
	}

	// Extract Sender Name (Support Dari or Pengirim)
	senderRe := regexp.MustCompile(`(?i)(?:Pengirim|Dari)[:\s]*([^\r\n<]+)`)
	senderMatch := senderRe.FindStringSubmatch(body)
	senderName := "Unknown"
	if len(senderMatch) >= 2 {
		senderName = strings.TrimSpace(senderMatch[1])
	}

	// 3. Search for matching order
	var order models.Order
	// ONLY check pending orders. Do NOT check expired/cancelled to prevent
	// reviving invalid orders (especially due to past duplicate unique code bugs)
	statusOptions := []string{"pending"}

	// Toleransi 10 menit (email bank kadang sedikit telat dari detik pembuatan order)
	if err := tx.Where("status IN ? AND total_amount = ? AND created_at <= ?",
		statusOptions, amount, emailDate.Add(10*time.Minute)).
		Order("created_at desc").
		First(&order).Error; err != nil {
		// Log matching failed (Maybe already Paid or not our order)
		// We STILL record this message ID to prevent re-processing every minute
		tx.Create(&models.BankTransaction{
			ReferenceID:     messageID,
			Amount:          amount,
			Sender:          senderName,
			BankName:        "Bank Jago (Unmatched)",
			TransactionDate: emailDate,
			RawData:         body,
			CreatedAt:       time.Now(),
		})
		tx.Commit()
		return
	}

	// 4. Payment Method Validation
	if !strings.Contains(order.PaymentMethod, "BANK_TRANSFER_JAGO") && order.PaymentMethod != "MANUAL_JAGO" {
		log.Printf("[%s-PaymentJob] Ignored order %s. Payment Method mismatch: %s\n", source, order.OrderNumber, order.PaymentMethod)
		// Still record to log to prevent check every time
		tx.Create(&models.BankTransaction{
			OrderID:         &order.ID,
			ReferenceID:     messageID,
			Amount:          amount,
			Sender:          senderName,
			BankName:        "Bank Jago (Mismatch Method)",
			TransactionDate: emailDate,
			RawData:         body,
			CreatedAt:       time.Now(),
		})
		tx.Commit()
		return
	}

	// 6. Mark as Paid
	now := time.Now()
	order.Status = "paid"
	order.PaidAt = &now
	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		return
	}

	// 7. Record History & Transaction
	tx.Create(&models.OrderStatusHistory{
		OrderID:   order.ID,
		Status:    "paid",
		Notes:     fmt.Sprintf("Verified %s via Email (%s). Original Status: %s", source, messageID, order.Status),
		CreatedAt: now,
	})

	tx.Create(&models.BankTransaction{
		OrderID:         &order.ID,
		ReferenceID:     messageID,
		Amount:          amount,
		Sender:          senderName,
		BankName:        "Bank Jago",
		TransactionDate: emailDate,
		RawData:         body,
		CreatedAt:       now,
	})

	if err := tx.Commit().Error; err != nil {
		return
	}

	log.Printf("[%s-PaymentJob] Order %s marked as PAID successfully\n", source, order.OrderNumber)

	// 8. Send Ticket (Outside transaction)
	var tickets []models.Ticket
	config.DB.Preload("Event").Preload("TicketType").Where("order_id = ?", order.ID).Find(&tickets)
	utils.SendTicketEmail(order, tickets)
}

func stripHTML(html string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(html, " ")
}
