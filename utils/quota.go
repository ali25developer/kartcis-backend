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

	// Group by ticket type and flash sale
	ticketTypeCounts := make(map[uint]int)
	flashSaleCounts := make(map[uint]int)

	for _, t := range tickets {
		ticketTypeCounts[t.TicketTypeID]++
		if t.FlashSaleID != nil {
			flashSaleCounts[*t.FlashSaleID]++
		}
	}

	// Restore normal ticket type available quota
	for ttID, count := range ticketTypeCounts {
		if err := tx.Model(&models.TicketType{}).Where("id = ?", ttID).
			Update("available", gorm.Expr("available + ?", count)).Error; err != nil {
			return err
		}
	}

	// Restore Flash Sale sold count
	for fsID, count := range flashSaleCounts {
		if err := tx.Model(&models.FlashSale{}).Where("id = ?", fsID).
			Update("sold", gorm.Expr("sold - ?", count)).Error; err != nil {
			return err
		}
	}
	return nil
}

// DeductQuota re-deducts ticket quotas when an order is revived
func DeductQuota(tx *gorm.DB, orderID uint) error {
	var tickets []models.Ticket
	if err := tx.Where("order_id = ?", orderID).Find(&tickets).Error; err != nil {
		return err
	}

	// Group by ticket type and flash sale
	ticketTypeCounts := make(map[uint]int)
	flashSaleCounts := make(map[uint]int)

	for _, t := range tickets {
		ticketTypeCounts[t.TicketTypeID]++
		if t.FlashSaleID != nil {
			flashSaleCounts[*t.FlashSaleID]++
		}
	}

	// Deduct normal ticket type available quota
	for ttID, count := range ticketTypeCounts {
		if err := tx.Model(&models.TicketType{}).Where("id = ?", ttID).
			Update("available", gorm.Expr("available - ?", count)).Error; err != nil {
			return err
		}
	}

	// Deduct Flash Sale sold count
	for fsID, count := range flashSaleCounts {
		if err := tx.Model(&models.FlashSale{}).Where("id = ?", fsID).
			Update("sold", gorm.Expr("sold + ?", count)).Error; err != nil {
			return err
		}
	}
	return nil
}
