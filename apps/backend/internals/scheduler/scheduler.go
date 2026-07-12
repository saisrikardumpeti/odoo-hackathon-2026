package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/models"
	"github.com/saisrikardumpeti/odoo-hackathon-2026/internals/repository"
)

func Start(ctx context.Context, store *repository.StorageRegistry) {
	ticker := time.NewTicker(3 * time.Minute)
	defer ticker.Stop()

	log.Println("Scheduler started (interval: 3 minutes)")

	for {
		select {
		case <-ctx.Done():
			log.Println("Scheduler stopped")
			return
		case <-ticker.C:
			run(ctx, store)
		}
	}
}

func run(ctx context.Context, store *repository.StorageRegistry) {
	if err := store.Booking.TransitionStatuses(ctx); err != nil {
		log.Printf("Scheduler: failed to transition booking statuses: %v", err)
	} else {
		log.Println("Scheduler: booking statuses transitioned")
	}

	if err := store.Booking.CreateReminders(ctx, 30); err != nil {
		log.Printf("Scheduler: failed to create booking reminders: %v", err)
	} else {
		log.Println("Scheduler: booking reminders created")
	}

	checkOverdueReturns(ctx, store)
}

func checkOverdueReturns(ctx context.Context, store *repository.StorageRegistry) {
	overdueAllocations, err := store.Allocation.ListOverdue(ctx)
	if err != nil {
		log.Printf("Scheduler: failed to list overdue allocations: %v", err)
		return
	}

	if len(overdueAllocations) == 0 {
		return
	}

	log.Printf("Scheduler: checking %d overdue allocations for notifications", len(overdueAllocations))

	notifType := "OverdueReturnAlert"
	for _, alloc := range overdueAllocations {
		if alloc.EmployeeID == nil {
			continue
		}

		exists, err := store.Notification.Exists(ctx, *alloc.EmployeeID, notifType, alloc.ID)
		if err != nil {
			log.Printf("Scheduler: failed to check existing notifications for allocation %s: %v", alloc.ID, err)
			continue
		}
		if exists {
			continue
		}

		entityType := "allocation"
		message := "Asset " + alloc.AssetTag + " (" + alloc.AssetName + ") is overdue for return."
		recipientID := *alloc.EmployeeID

		n := models.Notification{
			EmployeeID:        recipientID,
			Type:              notifType,
			Message:           message,
			RelatedEntityType: &entityType,
			RelatedEntityID:   &alloc.ID,
		}
		if err := store.Notification.Create(ctx, n); err != nil {
			log.Printf("Scheduler: failed to create overdue notification for allocation %s: %v", alloc.ID, err)
			continue
		}

		log.Printf("Scheduler: created OverdueReturnAlert for allocation %s (asset: %s)", alloc.ID, alloc.AssetTag)
	}
}
