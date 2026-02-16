package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"kartcis-backend/models"
	"log"
	"net/smtp"
	"os"
	"strings"
	"time"
)

type TicketEmailData struct {
	CustomerName         string
	OrderNumber          string
	EventTitle           string
	EventImage           string
	EventDate            string
	EventTime            string
	Venue                string
	City                 string
	TicketTypeName       string
	TicketCode           string
	Quantity             int
	CustomFieldResponses []CustomFieldResponse
}

type PaymentEmailData struct {
	CustomerName         string
	OrderNumber          string
	TotalAmount          string
	PaymentMethod        string
	VirtualAccountNumber string
	PaymentURL           string // For E-Wallet / Deep Link
	ExpiryTime           string
	CheckoutURL          string
	PaymentInstructions  string
	UniqueCode           int
}

type CustomFieldResponse struct {
	Label string
	Value string
}

type ResetPasswordEmailData struct {
	CustomerName string
	ResetURL     string
}

type CancellationEmailData struct {
	CustomerName string
	OrderNumber  string
	TotalAmount  string
	Reason       string
}

func SendTicketEmail(order models.Order, tickets []models.Ticket) {
	if len(tickets) == 0 {
		return
	}
	for _, ticket := range tickets {
		go sendTicketEmail(order, ticket)
	}
}

func SendPaymentInstructionEmail(order models.Order) {
	go sendPaymentEmail(order)
}

func SendOrderCancelledEmail(order models.Order, reason string) {
	go sendCancellationEmail(order, reason)
}

// For simplicity, we send one email per ticket or one email for the whole order?
// The template suggests one ticket detail, but let's send for the first ticket or loop.
// Usually, users get one email per ticket if it's a PDF delivery, or one summary.
// The provided template looks like an E-Ticket for a single attendee.

func sendTicketEmail(order models.Order, ticket models.Ticket) {
	// Load SMTP config
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	from := os.Getenv("SMTP_FROM")

	if smtpHost == "" || smtpUser == "" {
		log.Println("[Mailer] SMTP not configured, skipping email send for Order:", order.OrderNumber)
		return
	}

	// Parse Custom Fields
	var responses []CustomFieldResponse
	if ticket.CustomFieldResponses != "" && ticket.CustomFieldResponses != "null" {
		// Try map first
		var rawMap map[string]interface{}
		if err := json.Unmarshal([]byte(ticket.CustomFieldResponses), &rawMap); err == nil {
			for k, v := range rawMap {
				responses = append(responses, CustomFieldResponse{
					Label: k,
					Value: fmt.Sprintf("%v", v),
				})
			}
		} else {
			// Try slice of objects
			var rawSlice []map[string]interface{}
			if err := json.Unmarshal([]byte(ticket.CustomFieldResponses), &rawSlice); err == nil {
				for _, item := range rawSlice {
					label, _ := item["label"].(string)
					if label == "" {
						label, _ = item["key"].(string)
					}
					value := fmt.Sprintf("%v", item["value"])
					if label != "" && value != "<nil>" {
						responses = append(responses, CustomFieldResponse{
							Label: label,
							Value: value,
						})
					}
				}
			}
		}
	}

	data := TicketEmailData{
		CustomerName:         ticket.AttendeeName,
		OrderNumber:          order.OrderNumber,
		EventTitle:           ticket.Event.Title,
		EventImage:           formatImageURL(ticket.Event.Image),
		EventDate:            ticket.Event.EventDate.Format("02 Jan 2006"),
		EventTime:            ticket.Event.EventTime,
		Venue:                ticket.Event.Venue,
		City:                 ticket.Event.City,
		TicketTypeName:       ticket.TicketType.Name,
		TicketCode:           ticket.TicketCode,
		Quantity:             1,
		CustomFieldResponses: responses,
	}

	tmpl, err := template.New("ticket").Parse(htmlTemplate)
	if err != nil {
		log.Println("[Mailer] Template Parse Error:", err)
		return
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		log.Println("[Mailer] Template Execute Error:", err)
		return
	}

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	to := []string{ticket.AttendeeEmail}

	subject := fmt.Sprintf("[Berhasil] E-Tiket Anda untuk %s - #%s", ticket.Event.Title, order.OrderNumber)
	msg := []byte("From: " + from + "\r\n" +
		"To: " + ticket.AttendeeEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n\r\n" +
		body.String())

	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, msg)
	if err != nil {
		log.Println("[Mailer] SendMail Error:", err)
	} else {
		log.Printf("[Mailer] Email sent successfully to %s for ticket %s\n", ticket.AttendeeEmail, ticket.TicketCode)
	}
}

func sendPaymentEmail(order models.Order) {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	from := os.Getenv("SMTP_FROM")

	if smtpHost == "" || smtpUser == "" {
		return
	}

	expiryTime := order.CreatedAt.Add(24 * time.Hour)
	if order.ExpiresAt != nil {
		expiryTime = *order.ExpiresAt
	}

	data := PaymentEmailData{
		CustomerName:         order.CustomerName,
		OrderNumber:          order.OrderNumber,
		TotalAmount:          fmt.Sprintf("Rp %s", FormatPrice(order.TotalAmount)),
		PaymentMethod:        order.PaymentMethod,
		VirtualAccountNumber: order.VirtualAccountNumber,
		PaymentURL:           order.PaymentURL,
		ExpiryTime:           expiryTime.Format("02 Jan 2006, 15:04"),
		CheckoutURL:          fmt.Sprintf("%s/payment/%s", os.Getenv("FRONTEND_URL"), order.OrderNumber),
		PaymentInstructions:  order.PaymentInstructions,
		UniqueCode:           order.UniqueCode,
	}

	tmpl, err := template.New("payment").Parse(paymentHtmlTemplate)
	if err != nil {
		log.Println("[Mailer] Payment Template Parse Error:", err)
		return
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		log.Println("[Mailer] Payment Template Execute Error:", err)
		return
	}

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	to := []string{order.CustomerEmail}

	subject := fmt.Sprintf("Instruksi Pembayaran Kartcis.ID - %s", order.OrderNumber)
	msg := []byte("From: " + from + "\r\n" +
		"To: " + order.CustomerEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n\r\n" +
		body.String())

	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, msg)
	if err != nil {
		log.Println("[Mailer] SendMail Payment Error:", err)
	}
}

func sendCancellationEmail(order models.Order, reason string) {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	from := os.Getenv("SMTP_FROM")

	if smtpHost == "" || smtpUser == "" {
		return
	}

	data := CancellationEmailData{
		CustomerName: order.CustomerName,
		OrderNumber:  order.OrderNumber,
		TotalAmount:  fmt.Sprintf("Rp %s", FormatPrice(order.TotalAmount)),
		Reason:       reason,
	}

	tmpl, err := template.New("cancellation").Parse(cancellationHtmlTemplate)
	if err != nil {
		log.Println("[Mailer] Cancellation Template Parse Error:", err)
		return
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		log.Println("[Mailer] Cancellation Template Execute Error:", err)
		return
	}

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	to := []string{order.CustomerEmail}

	subject := fmt.Sprintf("Pesanan Dibatalkan - %s", order.OrderNumber)
	msg := []byte("From: " + from + "\r\n" +
		"To: " + order.CustomerEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n\r\n" +
		body.String())

	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, msg)
	if err != nil {
		log.Println("[Mailer] SendMail Cancellation Error:", err)
	}
}

func FormatPrice(price float64) string {
	p := int64(price)
	s := fmt.Sprintf("%d", p)
	res := ""
	for i, v := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			res += "."
		}
		res += string(v)
	}
	return res
}

func formatImageURL(imagePath string) string {
	if imagePath == "" {
		return "https://via.placeholder.com/600x300?text=No+Image"
	}
	if strings.HasPrefix(imagePath, "http") {
		return imagePath
	}

	// Fallback to API_URL
	baseURL := os.Getenv("API_URL")
	if baseURL == "" {
		// Try to construct from FRONTEND_URL but change port
		frontendURL := os.Getenv("FRONTEND_URL")
		if frontendURL != "" {
			baseURL = strings.Replace(frontendURL, "5173", "8000", 1)
		} else {
			baseURL = "http://localhost:8000"
		}
	}

	apiPrefix := os.Getenv("API_PREFIX")
	if apiPrefix == "" {
		apiPrefix = "/api/v1"
	}

	cleanPath := strings.TrimPrefix(imagePath, "/")
	// If the path doesn't already contain "uploads", add it
	if !strings.HasPrefix(cleanPath, "uploads") {
		cleanPath = "uploads/" + cleanPath
	}

	return fmt.Sprintf("%s%s/%s", strings.TrimSuffix(baseURL, "/"), apiPrefix, cleanPath)
}

const htmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>E-Tiket Konfirmasi Kartcis.ID</title>
    <style>
        body { font-family: 'Inter', -apple-system, BlinkMacSystemFont, Arial, sans-serif; background-color: #f9fafb; margin: 0; padding: 0; }
        .wrapper { padding: 20px; }
        .container { max-width: 600px; margin: 0 auto; background-color: #ffffff; border: 1px solid #e5e7eb; border-radius: 12px; overflow: hidden; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
        .header { background: linear-gradient(to right, #b31356, #d61a6b); padding: 32px 24px; text-align: left; color: #ffffff; }
        .header h1 { margin: 0; font-size: 24px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.5px; }
        .header p { margin: 8px 0 0; font-size: 15px; opacity: 0.95; line-height: 1.4; }
        .content { padding: 24px; }
        .ticket-item { border: 2px solid #e2e8f0; border-radius: 12px; margin-bottom: 24px; overflow: hidden; }
        .event-banner { background-color: #f1f5f9; text-align: center; }
        .event-banner img { width: 100%; height: auto; display: block; max-height: 250px; object-fit: cover; }
        .ticket-details { padding: 24px; }
        .event-title { font-size: 20px; font-weight: 700; color: #111827; margin: 0 0 12px; line-height: 1.2; }
        .badge { display: inline-block; padding: 4px 12px; border-radius: 99px; font-size: 11px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.5px; }
        .badge-sky { background-color: #fff1f2; color: #b31356; border: 1px solid #fecdd3; }
        .info-grid { margin: 20px 0; border-top: 1px solid #f1f5f9; padding-top: 16px; }
        .info-item { display: flex; align-items: flex-start; font-size: 14px; color: #374151; margin-bottom: 12px; }
        .info-icon { margin-right: 12px; width: 18px; text-align: center; }
        .custom-fields { background-color: #f8fafc; border: 1px solid #e2e8f0; border-radius: 8px; padding: 16px; margin: 20px 0; }
        .custom-fields-title { font-size: 11px; font-weight: 700; color: #475569; text-transform: uppercase; margin-bottom: 12px; border-bottom: 1px solid #e2e8f0; padding-bottom: 6px; }
        .ticket-footer { border-top: 1px solid #e5e7eb; padding-top: 16px; display: table; width: 100%; }
        .footer-left { display: table-cell; width: 60%; vertical-align: middle; }
        .footer-right { display: table-cell; width: 40%; vertical-align: middle; text-align: right; }
        .code-label { font-size: 11px; color: #64748b; text-transform: uppercase; font-weight: 600; margin-bottom: 4px; }
        .code-value { font-family: 'Courier New', monospace; font-size: 18px; font-weight: 700; color: #b31356; }
        .qr-section { background: #ffffff; padding: 40px 24px; text-align: center; border-top: 4px solid #ffd54c; }
        .qr-code-img { background: white; padding: 15px; border: 2px solid #f1f5f9; border-radius: 12px; box-shadow: 0 4px 6px -1px rgba(0,0,0,0.1); margin-bottom: 16px; }
        .legal-footer { text-align: center; padding: 32px 24px; font-size: 12px; color: #94a3b8; line-height: 1.6; }
    </style>
</head>
<body>
    <div class="wrapper">
        <div class="container">
            <div class="header">
                <h1>Tiket Pesanan Anda</h1>
                <p>Halo <b>{{.CustomerName}}</b>, pembayaran Anda telah berhasil dikonfirmasi. Berikut adalah detail tiket untuk pesanan <b>#{{.OrderNumber}}</b>.</p>
            </div>
            <div class="content">
                <div class="ticket-item">
                    <div class="event-banner">
                        <img src="{{.EventImage}}" alt="Event Image">
                    </div>
                    <div class="ticket-details">
                        <div class="badge badge-sky">{{.TicketTypeName}}</div>
                        <h2 class="event-title">{{.EventTitle}}</h2>
                        <div class="info-grid">
                            <div class="info-item">
                                <span class="info-icon">📅</span> 
                                <span><b>Tanggal:</b> {{.EventDate}}</span>
                            </div>
                            <div class="info-item">
                                <span class="info-icon">🕒</span> 
                                <span><b>Waktu:</b> {{.EventTime}} WIB</span>
                            </div>
                            <div class="info-item">
                                <span class="info-icon">📍</span> 
                                <span><b>Lokasi:</b> {{.Venue}}, {{.City}}</span>
                            </div>
                        </div>
						{{if .CustomFieldResponses}}
						<div class="custom-fields">
							<div class="custom-fields-title">Data Pengunjung</div>
							<table width="100%" style="font-size: 14px; border-collapse: collapse;">
								{{range .CustomFieldResponses}}
								<tr>
									<td style="color: #64748b; padding: 4px 0;">{{.Label}}</td>
									<td style="font-weight: 600; color: #1e293b; text-align: right; padding: 4px 0;">{{.Value}}</td>
								</tr>
								{{end}}
							</table>
						</div>
						{{end}}
                        <div class="ticket-footer">
                            <div class="footer-left">
                                <div class="code-label">Kode Unik Tiket</div>
                                <div class="code-value">{{.TicketCode}}</div>
                            </div>
                            <div class="footer-right">
                                <div class="code-label">Jumlah</div>
                                <div style="font-weight: 700; color: #1e293b; font-size: 16px;">{{.Quantity}} Tiket</div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
            <div class="qr-section">
                <p style="font-size: 13px; font-weight: 700; color: #64748b; text-transform: uppercase; margin: 0 0 16px;">Tunjukkan QR Code ini saat masuk</p>
                <img src="https://api.qrserver.com/v1/create-qr-code/?size=180x180&data={{.TicketCode}}" class="qr-code-img" width="180" height="180" alt="Check-in QR Code">
                <p style="font-size: 14px; color: #334155; margin-top: 10px;">Gunakan layar penuh dan cerahkan pencahayaan HP Anda saat di-*scan* oleh petugas.</p>
            </div>
            <div class="legal-footer">
                Pesanan ini diproses secara aman oleh <b>Kartcis.ID</b>.<br>
                Jika ada pertanyaan mengenai event, Anda dapat menghubungi penyelenggara secara langsung atau balas email ini untuk bantuan customer service kami.
            </div>
        </div>
    </div>
</body>
</html>
`

const paymentHtmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Instruksi Pembayaran Kartcis.ID</title>
    <style>
        body { font-family: 'Inter', Arial, sans-serif; background-color: #f3f4f6; margin: 0; padding: 0; }
        .wrapper { padding: 20px; }
        .container { max-width: 500px; margin: 0 auto; background-color: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 4px 6px -1px rgba(0,0,0,0.1); }
        .header { background-color: #b31356; padding: 24px; text-align: center; color: #ffffff; }
        .content { padding: 32px 24px; }
        .amount-box { background-color: #fff1f2; border: 1px solid #fecdd3; border-radius: 8px; padding: 20px; text-align: center; margin-bottom: 24px; }
        .amount-label { font-size: 13px; color: #64748b; margin-bottom: 8px; }
        .amount-value { font-size: 28px; font-weight: 800; color: #b31356; }
        .method-info { margin-bottom: 24px; border-bottom: 1px solid #f1f5f9; padding-bottom: 16px; }
        .label { font-size: 13px; color: #64748b; margin-bottom: 4px; }
        .value { font-size: 16px; font-weight: 600; color: #1e293b; }
        .va-box { background-color: #f1f5f9; padding: 12px; border-radius: 6px; font-family: monospace; font-size: 18px; color: #b31356; letter-spacing: 1px; margin-top: 4px; }
        .expiry-alert { background-color: #fffbeb; border: 1px solid #fef3c7; border-radius: 8px; padding: 16px; margin-top: 24px; }
        .expiry-text { font-size: 14px; color: #92400e; }
        .btn { display: block; background-color: #b31356; color: #ffffff !important; text-align: center; padding: 14px; border-radius: 8px; text-decoration: none; font-weight: 700; margin-top: 24px; border-bottom: 4px solid #ffd54c; }
        .footer { text-align: center; padding: 24px; font-size: 12px; color: #94a3b8; }
    </style>
</head>
<body>
    <div class="wrapper">
        <div class="container">
            <div class="header">
                <h2 style="margin:0">Instruksi Pembayaran</h2>
            </div>
            <div class="content">
                <p>Halo <b>{{.CustomerName}}</b>,</p>
                <p>Pesanan Anda <b>#{{.OrderNumber}}</b> telah kami terima. Silakan selesaikan pembayaran agar tiket dapat segera kami kirimkan.</p>
                
                <div class="amount-box">
                    <div class="amount-label">Total Pembayaran</div>
                    <div class="amount-value">{{.TotalAmount}}</div>
                </div>

                <div class="method-info">
                    <div class="label">Metode Pembayaran</div>
                    <div class="value">{{.PaymentMethod}}</div>
                </div>

                {{if .VirtualAccountNumber}}
                <div class="method-info">
                    <div class="label">Nomor Rekening / Virtual Account</div>
                    <div class="va-box">{{.VirtualAccountNumber}}</div>
                </div>
                {{end}}

                {{if .PaymentInstructions}}
                <div class="method-info">
                    <div class="label">Instruksi Pembayaran</div>
                    <div class="value" style="font-size: 14px; font-weight: 400; color: #475569; line-height: 1.5;">{{.PaymentInstructions}}</div>
                </div>
                {{end}}

                {{if .PaymentURL}}
                <div class="method-info">
                    <div class="label">Link Pembayaran</div>
                    <div style="margin-top: 8px;">
                        <a href="{{.PaymentURL}}" style="display: inline-block; background-color: #10b981; color: #ffffff; padding: 10px 16px; border-radius: 6px; text-decoration: none; font-weight: 600; font-size: 14px;">Bayar Sekarang</a>
                    </div>
                </div>
                {{end}}

                <div class="expiry-alert">
                    <div class="expiry-text">
                        ⚠️ <b>Batas Waktu Pembayaran:</b><br>
                        Segera bayar sebelum <b>{{.ExpiryTime}} WIB</b> atau pesanan Anda akan dibatalkan otomatis.
                    </div>
                </div>

                <a href="{{.CheckoutURL}}" class="btn">Lihat Detail Pesanan</a>
            </div>
            <div class="footer">
                &copy; 2026 Kartcis.ID. Seluruh hak cipta dilindungi.
            </div>
        </div>
    </div>
</body>
</html>
`

func SendResetPasswordEmail(email, name, token string) {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	from := os.Getenv("SMTP_FROM")

	if smtpHost == "" || smtpUser == "" {
		// Log but don't crash
		log.Println("[Mailer] SMTP not configured, Reset Password link: ", fmt.Sprintf("%s/reset-password?token=%s&email=%s", os.Getenv("FRONTEND_URL"), token, email))
		return
	}

	// Construct Link
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}
	resetLink := fmt.Sprintf("%s/reset-password?token=%s&email=%s", frontendURL, token, email)

	data := ResetPasswordEmailData{
		CustomerName: name,
		ResetURL:     resetLink,
	}

	tmpl, err := template.New("reset_password").Parse(passwordResetHtmlTemplate)
	if err != nil {
		log.Println("[Mailer] Reset Template Parse Error:", err)
		return
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		log.Println("[Mailer] Reset Template Execute Error:", err)
		return
	}

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	to := []string{email}

	subject := "Reset Password - Kartcis.ID"
	msg := []byte("From: " + from + "\r\n" +
		"To: " + email + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=UTF-8\r\n\r\n" +
		body.String())

	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, msg)
	if err != nil {
		log.Println("[Mailer] SendMail Reset Error:", err)
	} else {
		log.Printf("[Mailer] Reset Password email sent to %s\n", email)
	}
}

const passwordResetHtmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Reset Password Kartcis.ID</title>
    <style>
        body { font-family: 'Inter', Arial, sans-serif; background-color: #f3f4f6; margin: 0; padding: 0; }
        .wrapper { padding: 40px 20px; }
        .container { max-width: 500px; margin: 0 auto; background-color: #ffffff; border-radius: 12px; box-shadow: 0 4px 6px -1px rgba(0,0,0,0.1); padding: 40px; text-align: center; }
        h2 { color: #1e293b; margin-top: 0; }
        p { color: #64748b; font-size: 16px; line-height: 1.5; margin-bottom: 24px; }
        .btn { display: inline-block; background-color: #b31356; color: #ffffff !important; padding: 14px 24px; border-radius: 6px; text-decoration: none; font-weight: 600; font-size: 16px; border-bottom: 3px solid #ffd54c; }
        .footer { margin-top: 32px; font-size: 12px; color: #94a3b8; }
    </style>
</head>
<body>
    <div class="wrapper">
        <div class="container">
            <h2>Reset Password</h2>
            <p>Halo <b>{{.CustomerName}}</b>,<br>Kami menerima permintaan untuk mereset password akun Kartcis.ID Anda. Klik tombol di bawah ini untuk membuat password baru:</p>
            
            <a href="{{.ResetURL}}" class="btn">Reset Password Saya</a>
            
            <p style="margin-top: 24px; font-size: 14px;">Link ini berlaku selama 1 jam. Jika Anda tidak merasa meminta reset password, abaikan email ini.</p>
        </div>
        <div class="footer">
            &copy; 2026 Kartcis.ID
        </div>
    </div>
</body>
</html>
`

const cancellationHtmlTemplate = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Pesanan Dibatalkan - Kartcis.ID</title>
    <style>
        body { font-family: 'Inter', Arial, sans-serif; background-color: #f3f4f6; margin: 0; padding: 0; }
        .wrapper { padding: 40px 20px; }
        .container { max-width: 500px; margin: 0 auto; background-color: #ffffff; border-radius: 12px; box-shadow: 0 4px 6px -1px rgba(0,0,0,0.1); overflow: hidden; }
        .header { background-color: #1e293b; padding: 24px; text-align: center; color: #ffffff; }
        .content { padding: 32px 24px; text-align: center; }
        h2 { color: #e11d48; margin-top: 0; }
        p { color: #64748b; font-size: 16px; line-height: 1.5; margin-bottom: 24px; }
        .order-info { background-color: #f8fafc; border: 1px solid #e2e8f0; border-radius: 8px; padding: 16px; margin-bottom: 24px; text-align: left; }
        .label { font-size: 13px; color: #64748b; }
        .value { font-size: 15px; font-weight: 600; color: #1e293b; }
        .footer { text-align: center; padding: 24px; font-size: 12px; color: #94a3b8; }
    </style>
</head>
<body>
    <div class="wrapper">
        <div class="container">
            <div class="header">
                <h3 style="margin:0">Kartcis.ID</h3>
            </div>
            <div class="content">
                <h2>Pesanan Dibatalkan</h2>
                <p>Halo <b>{{.CustomerName}}</b>,<br>Pesanan Anda <b>#{{.OrderNumber}}</b> telah dibatalkan.</p>
                
                <div class="order-info">
                    <div style="margin-bottom: 12px;">
                        <span class="label">Alasan Pembatalan:</span><br>
                        <span class="value">{{.Reason}}</span>
                    </div>
                    <div>
                        <span class="label">Total Nominal:</span><br>
                        <span class="value">{{.TotalAmount}}</span>
                    </div>
                </div>

                <p style="font-size: 14px;">Jika Anda merasa ini adalah kesalahan atau sudah melakukan pembayaran, silakan hubungi tim support kami dengan melampirkan bukti transfer.</p>
            </div>
            <div class="footer">
                &copy; 2026 Kartcis.ID. Seluruh hak cipta dilindungi.
            </div>
        </div>
    </div>
</body>
</html>
`
