package jobs

import (
	"fmt"
	"kartcis-backend/config"
	"kartcis-backend/models"
	"kartcis-backend/utils"
	"time"

	"gorm.io/gorm"
)

// StartOrderExpiryJob starts a background goroutine to expire old pending orders
func StartOrderExpiryJob() {
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for range ticker.C {
			expireOrders()
		}
	}()
}

func expireOrders() {
	if config.DB == nil {
		return
	}

	expiryTime := time.Now().Add(-30 * time.Minute)
	var orders []models.Order

	// Find pending orders older than 30 minutes
	err := config.DB.Where("status = ? AND created_at <= ?", "pending", expiryTime).Find(&orders).Error
	if err != nil {
		fmt.Printf("[ExpiryJob] Error fetching orders: %v\n", err)
		return
	}

	if len(orders) == 0 {
		return
	}

	fmt.Printf("[ExpiryJob] Processing %d expired orders...\n", len(orders))

	for _, order := range orders {
		err := config.DB.Transaction(func(tx *gorm.DB) error {
			// 1. Update status to expired
			if err := tx.Model(&order).Update("status", "expired").Error; err != nil {
				return err
			}

			// 2. Restore Quota
			if err := utils.RestoreQuota(tx, order.ID); err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			fmt.Printf("[ExpiryJob] Failed to expire Order %s: %v\n", order.OrderNumber, err)
		} else {
			fmt.Printf("[ExpiryJob] Order %s successfully expired and quota restored\n", order.OrderNumber)
		}
	}
}
