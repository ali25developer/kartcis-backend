package utils

import (
	"kartcis-backend/models"

	"gorm.io/gorm"
)

// RestoreQuota restores ticket quotas when an order is cancelled or expired
func RestoreQuota(tx *gorm.DB, orderID uint) error {
	var tickets []models.Ticket
	if err := tx.Where("order_id = ?", orderID).Find(&tickets).Error; err != nil {
		return err
	}

	// Group by ticket type
	counts := make(map[uint]int)
	for _, t := range tickets {
		counts[t.TicketTypeID]++
	}

	for ttID, count := range counts {
		// Get the current Quota to ensure we don't restore beyond it
		var tt models.TicketType
		if err := tx.First(&tt, ttID).Error; err != nil {
			continue
		}

		newAvailable := tt.Available + count
		if newAvailable > tt.Quota {
			newAvailable = tt.Quota
		}

		if err := tx.Model(&models.TicketType{}).Where("id = ?", ttID).
			Update("available", newAvailable).Error; err != nil {
			return err
		}
	}
	return nil
}
