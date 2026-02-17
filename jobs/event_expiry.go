package jobs

import (
	"fmt"
	"kartcis-backend/config"
	"kartcis-backend/models"
	"time"
)

// StartEventExpiryJob starts a background goroutine to mark past events as completed
func StartEventExpiryJob() {
	ticker := time.NewTicker(10 * time.Minute)
	go func() {
		for range ticker.C {
			expireEvents()
		}
	}()
}

func expireEvents() {
	if config.DB == nil {
		return
	}

	now := time.Now()

	// Find published events where EventDate is already in the past
	var events []models.Event
	err := config.DB.Where("status = ? AND event_date < ?", "published", now).Find(&events).Error
	if err != nil {
		fmt.Printf("[EventJob] Error fetching events: %v\n", err)
		return
	}

	if len(events) == 0 {
		return
	}

	fmt.Printf("[EventJob] Processing %d events to mark as COMPLETED...\n", len(events))

	for _, event := range events {
		// Update status to completed
		if err := config.DB.Model(&event).Update("status", "completed").Error; err != nil {
			fmt.Printf("[EventJob] Failed to complete Event %s: %v\n", event.Title, err)
		} else {
			fmt.Printf("[EventJob] Event %s (ID: %d) successfully marked as COMPLETED\n", event.Title, event.ID)
		}
	}
}
