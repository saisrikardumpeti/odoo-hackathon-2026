package scheduler

import (
	"context"
	"log"
	"time"

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
}
