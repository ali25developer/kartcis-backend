package jobs

import (
	"fmt"
	"kartcis-backend/config"
	"kartcis-backend/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() {
	// Use SQLite in-Memory with shared cache and higher timeout for concurrent testing
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared&_busy_timeout=5000"), &gorm.Config{})
	config.DB = db

	db.AutoMigrate(
		&models.Order{},
		&models.Ticket{},
		&models.TicketType{},
		&models.FlashSale{},
		&models.BankTransaction{},
		&models.OrderStatusHistory{},
		&models.Event{},
	)
}

func TestPaymentChecker_HandleOldEmail(t *testing.T) {
	setupTestDB()

	// 1. Create a "New" Order
	order := models.Order{
		OrderNumber:   "ORD-NEW-101",
		TotalAmount:   100302,
		Status:        "pending",
		PaymentMethod: "BANK_TRANSFER_JAGO",
		CreatedAt:     time.Now(),
	}
	config.DB.Create(&order)

	// 2. Simulate an "Old" Email (Arrived 1 hour BEFORE order was even created)
	oldEmailDate := time.Now().Add(-1 * time.Hour)
	emailBody := "Nominal: Rp 100.302. Pengirim: ALICE"
	messageID := "<old-email-123@jago.com>"

	// Run process (This should be ignored because order.CreatedAt > emailDate)
	ProcessJagoEmail(emailBody, "Test", messageID, oldEmailDate)

	// Verify order status is still pending
	var updatedOrder models.Order
	config.DB.First(&updatedOrder, order.ID)
	assert.Equal(t, "pending", updatedOrder.Status, "Order should NOT be completed by an old email")

	// Verify log recorded as Unmatched
	var tx models.BankTransaction
	config.DB.Where("reference_id = ?", messageID).First(&tx)
	assert.Contains(t, tx.BankName, "Unmatched", "Old email should be logged as unmatched")
}

func TestPaymentChecker_ValidPayment(t *testing.T) {
	setupTestDB()

	// 1. Create Pending Order
	order := models.Order{
		OrderNumber:   "ORD-VALID-202",
		TotalAmount:   50000,
		Status:        "pending",
		PaymentMethod: "BANK_TRANSFER_JAGO",
		CreatedAt:     time.Now().Add(-10 * time.Minute), // Created 10 mins ago
	}
	config.DB.Create(&order)

	// 2. Valid Email (Arrived 5 mins ago)
	emailDate := time.Now().Add(-5 * time.Minute)
	emailBody := "Nominal: Rp 50.000. Pengirim: BOB"
	messageID := "<valid-email-456@jago.com>"

	ProcessJagoEmail(emailBody, "Test", messageID, emailDate)

	// Verify Paid
	var updatedOrder models.Order
	config.DB.First(&updatedOrder, order.ID)
	assert.Equal(t, "paid", updatedOrder.Status)
	assert.NotNil(t, updatedOrder.PaidAt)

	// Verify Logged correctly
	var tx models.BankTransaction
	config.DB.Where("reference_id = ?", messageID).First(&tx)
	assert.Equal(t, float64(50000), tx.Amount)
	assert.Equal(t, "BOB", tx.Sender)
}

func TestPaymentChecker_Deduplication(t *testing.T) {
	setupTestDB()

	order := models.Order{
		TotalAmount:   75000,
		Status:        "pending",
		PaymentMethod: "BANK_TRANSFER_JAGO",
		CreatedAt:     time.Now().Add(-1 * time.Hour),
	}
	config.DB.Create(&order)

	emailBody := "Nominal: Rp 75.000"
	messageID := "<unique-msg-999>"
	emailDate := time.Now()

	// Process 1st time
	ProcessJagoEmail(emailBody, "Test", messageID, emailDate)

	// Process 2nd time (should skip)
	ProcessJagoEmail(emailBody, "Test", messageID, emailDate)

	var count int64
	config.DB.Model(&models.BankTransaction{}).Where("reference_id = ?", messageID).Count(&count)
	assert.Equal(t, int64(1), count, "Email should only be recorded once")
}

func TestQuota_ReviveExpiredOrder(t *testing.T) {
	setupTestDB()

	// 1. Setup Ticket Type with limited quota
	tt := models.TicketType{
		Name:      "Early Bird",
		Available: 0, // SOLD OUT
		Quota:     10,
	}
	config.DB.Create(&tt)

	// 2. Create an EXPIRED order that originally held 2 tickets
	order := models.Order{
		OrderNumber:   "ORD-EXP-303",
		Status:        "expired",
		TotalAmount:   200000,
		PaymentMethod: "BANK_TRANSFER_JAGO",
		CreatedAt:     time.Now().Add(-2 * time.Hour),
	}
	config.DB.Create(&order)

	config.DB.Create(&models.Ticket{OrderID: &order.ID, TicketTypeID: tt.ID})
	config.DB.Create(&models.Ticket{OrderID: &order.ID, TicketTypeID: tt.ID})

	// 3. Late Email payment arrives
	emailDate := time.Now().Add(-1 * time.Hour) // Paid after it expired but email arrived later
	emailBody := "Nominal: Rp 200.000"
	messageID := "<late-payment-001>"

	ProcessJagoEmail(emailBody, "Test", messageID, emailDate)

	// Verify order is PAID anyway (Customer Satisfaction)
	var updatedOrder models.Order
	config.DB.First(&updatedOrder, order.ID)
	assert.Equal(t, "paid", updatedOrder.Status)

	// Verify quota is now -2 (Overbooked but tracked)
	var updatedTT models.TicketType
	config.DB.First(&updatedTT, tt.ID)
	assert.Equal(t, -2, updatedTT.Available, "Quota should be deducted (negative if sold out) to maintain integrity")
}

func TestPaymentChecker_NormalTemplateVariations(t *testing.T) {
	setupTestDB()

	// 1. Test "TOTAL" instead of "Nominal"
	config.DB.Create(&models.Order{OrderNumber: "VAR-1", TotalAmount: 150000, Status: "pending", PaymentMethod: "BANK_TRANSFER_JAGO", CreatedAt: time.Now().Add(-1 * time.Hour)})
	ProcessJagoEmail("TOTAL: Rp 150.000", "Email", "<id-1>", time.Now())

	var o1 models.Order
	config.DB.Where("order_number = ?", "VAR-1").First(&o1)
	assert.Equal(t, "paid", o1.Status, "Template with 'TOTAL' should work")

	// 2. Test "Jumlah" with lowercase
	config.DB.Create(&models.Order{OrderNumber: "VAR-2", TotalAmount: 75000, Status: "pending", PaymentMethod: "BANK_TRANSFER_JAGO", CreatedAt: time.Now().Add(-1 * time.Hour)})
	ProcessJagoEmail("jumlah: Rp 75.000", "Email", "<id-2>", time.Now())

	var o2 models.Order
	config.DB.Where("order_number = ?", "VAR-2").First(&o2)
	assert.Equal(t, "paid", o2.Status, "Template with 'jumlah' should work")
}

func TestPaymentChecker_ExtremeTemplatesAndEdgeCases(t *testing.T) {
	setupTestDB()

	// 1. Extreme Sender Name (Special Characters, Numbers, Dots)
	config.DB.Create(&models.Order{OrderNumber: "EXT-1", TotalAmount: 55000, Status: "pending", PaymentMethod: "BANK_TRANSFER_JAGO", CreatedAt: time.Now().Add(-1 * time.Hour)})
	ProcessJagoEmail("Nominal: Rp 55.000. Pengirim: PT. MAJU-JAYA (DARI: 0812345)!!!", "Email", "<ext-id-1>", time.Now())

	var tx models.BankTransaction
	config.DB.Where("reference_id = ?", "<ext-id-1>").First(&tx)
	assert.Equal(t, "PT. MAJU-JAYA (DARI: 0812345)!!!", tx.Sender, "Should capture extreme sender names correctly")

	// 2. Mismatch Amount (Safety Check)
	config.DB.Create(&models.Order{OrderNumber: "EXT-2", TotalAmount: 100302, Status: "pending", PaymentMethod: "BANK_TRANSFER_JAGO", CreatedAt: time.Now().Add(-1 * time.Hour)})
	ProcessJagoEmail("Nominal: Rp 100.301", "Email", "<ext-id-2>", time.Now()) // Missing 1 rupiah

	var o2 models.Order
	config.DB.Where("order_number = ?", "EXT-2").First(&o2)
	assert.Equal(t, "pending", o2.Status, "Order should NOT be paid if amount is off by 1 rupiah")

	// 3. Email Arriving Exactly at the same millisecond as CreatedAt (Grace Period Test)
	now := time.Now()
	config.DB.Create(&models.Order{OrderNumber: "EXT-3", TotalAmount: 12000, Status: "pending", PaymentMethod: "BANK_TRANSFER_JAGO", CreatedAt: now})
	ProcessJagoEmail("Nominal: Rp 12.000", "Email", "<ext-id-3>", now) // Exact same time

	var o3 models.Order
	config.DB.Where("order_number = ?", "EXT-3").First(&o3)
	assert.Equal(t, "paid", o3.Status, "Should pass if time is exactly the same")
}

func TestPaymentChecker_HighVolumeConcurrency(t *testing.T) {
	setupTestDB()

	count := 5 // Stable for SQLite
	baseAmount := 100000.0
	for i := 0; i < count; i++ {
		amount := baseAmount + float64(101+i)
		config.DB.Create(&models.Order{
			OrderNumber:   fmt.Sprintf("ORD-CONC-%d", i),
			TotalAmount:   amount,
			Status:        "pending",
			PaymentMethod: "BANK_TRANSFER_JAGO",
			CreatedAt:     time.Now().Add(-1 * time.Hour),
		})
	}

	startSignal := make(chan bool)
	doneSignal := make(chan bool)
	emailDate := time.Now()

	for i := 0; i < count; i++ {
		amount := baseAmount + float64(101+i)
		id := fmt.Sprintf("<msg-unique-%d@jago.com>", i)
		go func(mid string, amt float64) {
			<-startSignal
			time.Sleep(time.Duration(i*10) * time.Millisecond) // Spread them out slightly for SQLite
			ProcessJagoEmail(fmt.Sprintf("Nominal: Rp %v", amt), "SyncTest", mid, emailDate)
			doneSignal <- true
		}(id, amount)
	}

	close(startSignal)
	for i := 0; i < count; i++ {
		<-doneSignal
	}

	var paidCount int64
	config.DB.Model(&models.Order{}).Where("status = ? AND order_number LIKE ?", "paid", "ORD-CONC-%").Count(&paidCount)
	assert.Equal(t, int64(count), paidCount)
}

func TestPaymentChecker_DuplicateID_RaceCondition(t *testing.T) {
	setupTestDB()

	count := 5 // Stable for SQLite
	order := models.Order{
		OrderNumber:   "ORD-RACE-777",
		TotalAmount:   999000,
		Status:        "pending",
		PaymentMethod: "BANK_TRANSFER_JAGO",
		CreatedAt:     time.Now().Add(-1 * time.Hour),
	}
	config.DB.Create(&order)

	msgID := "<THE-SAME-MSG-ID>"
	emailBody := "Nominal: Rp 999.000"
	emailDate := time.Now()

	startSignal := make(chan bool)
	doneSignal := make(chan bool)

	for i := 0; i < count; i++ {
		go func(idx int) {
			<-startSignal
			time.Sleep(time.Duration(idx*10) * time.Millisecond) // Spread out
			ProcessJagoEmail(emailBody, "RaceTest", msgID, emailDate)
			doneSignal <- true
		}(i)
	}

	close(startSignal)
	for i := 0; i < count; i++ {
		<-doneSignal
	}

	var logCount int64
	config.DB.Model(&models.BankTransaction{}).Where("reference_id = ?", msgID).Count(&logCount)
	assert.Equal(t, int64(1), logCount)
}

func TestPaymentChecker_RealUserEmail(t *testing.T) {
	setupTestDB()

	// 1. Create Order matching the user's real email nominal
	order := models.Order{
		OrderNumber:   "ORD-REAL-999",
		TotalAmount:   163311,
		Status:        "pending",
		PaymentMethod: "BANK_TRANSFER_JAGO",
		CreatedAt:     time.Now().Add(-1 * time.Hour),
	}
	config.DB.Create(&order)

	// 2. Real Email Content Snippet (with HTML tags and NEW labels)
	emailBody := `
		<td class="transfer-table-title">Dari</td>
		<td class="transfer-table-content">INTAN SRI RAHAY</td>
		<td class="transfer-table-title">Jumlah</td>
		<td class="transfer-table-content">Rp163.311</td>
	`
	messageID := "<real-jago-msg-id-001>"
	emailDate := time.Now()

	ProcessJagoEmail(emailBody, "RealTest", messageID, emailDate)

	// Verify Paid
	var updatedOrder models.Order
	config.DB.Where("order_number = ?", "ORD-REAL-999").First(&updatedOrder)
	assert.Equal(t, "paid", updatedOrder.Status, "Should handle 'Dari' and 'Jumlah' labels correctly")

	// Verify Sender Name captured correctly
	var tx models.BankTransaction
	config.DB.Where("reference_id = ?", messageID).First(&tx)
	assert.Equal(t, "INTAN SRI RAHAY", tx.Sender)
}

func TestPaymentChecker_IdenticalUniqueCodes(t *testing.T) {
	setupTestDB()

	// SCENARIO: Two different users get the SAME total amount (e.g. due to unique code loop)
	amount := 150123.0

	// Order A (User 1) - Created 2 hours ago
	orderA := models.Order{OrderNumber: "ORD-A", TotalAmount: amount, Status: "pending", PaymentMethod: "BANK_TRANSFER_JAGO", CreatedAt: time.Now().Add(-2 * time.Hour)}
	config.DB.Create(&orderA)

	// Order B (User 2) - Created 1 hour ago (Most Recent)
	orderB := models.Order{OrderNumber: "ORD-B", TotalAmount: amount, Status: "pending", PaymentMethod: "BANK_TRANSFER_JAGO", CreatedAt: time.Now().Add(-1 * time.Hour)}
	config.DB.Create(&orderB)

	// 1. Email for first payment arrives
	ProcessJagoEmail("Jumlah: Rp 150.123", "Email", "<msg-transfer-1>", time.Now())

	// Result: Should match the MOST RECENT pending order (Order B)
	var checkB models.Order
	config.DB.Where("order_number = ?", "ORD-B").First(&checkB)
	assert.Equal(t, "paid", checkB.Status, "Recent order should be paid first")

	var checkA models.Order
	config.DB.Where("order_number = ?", "ORD-A").First(&checkA)
	assert.Equal(t, "pending", checkA.Status, "Older order should still be pending")

	// 2. Email for second payment arrives
	ProcessJagoEmail("Jumlah: Rp 150.123", "Email", "<msg-transfer-2>", time.Now())

	// Result: Since Order B is already PAID, it should now match Order A
	config.DB.Where("order_number = ?", "ORD-A").First(&checkA)
	assert.Equal(t, "paid", checkA.Status, "Older order should now be paid because the recent one is already cleared")
}
