package store_test

import (
	"testing"
	"time"

	"webhook-delivery-service/internal/event"
	"webhook-delivery-service/internal/store"
)

func TestPollDueEvents_Isolation(t *testing.T) {
	s := store.NewStore()

	// Add 5 events for customer A
	for i := 0; i < 5; i++ {
		s.SaveEvent(&event.Event{
			ID:          "A" + string(rune('0'+i)),
			CustomerID:  "custA",
			Status:      event.StatusPending,
			NextRetryAt: time.Now().Add(-1 * time.Minute), // due now
		})
	}

	// Add 2 events for customer B
	for i := 0; i < 2; i++ {
		s.SaveEvent(&event.Event{
			ID:          "B" + string(rune('0'+i)),
			CustomerID:  "custB",
			Status:      event.StatusPending,
			NextRetryAt: time.Now().Add(-1 * time.Minute), // due now
		})
	}

	// Suppose custA already has 1 event in-flight. Max is 2 per customer.
	inFlightC := map[string]int{"custA": 1}
	inFlightE := map[string]struct{}{"A0": {}} // A0 is actively being processed

	limit := 10
	maxPerCustomer := 2

	due := s.PollDueEvents(limit, maxPerCustomer, inFlightC, inFlightE)

	counts := make(map[string]int)
	for _, e := range due {
		if e.ID == "A0" {
			t.Errorf("PollDueEvents returned A0 which was already in-flight")
		}
		counts[e.CustomerID]++
	}

	// custA should only get 1 more event (since it had 1 in flight, and max is 2)
	if counts["custA"] != 1 {
		t.Errorf("Expected 1 event for custA, got %d", counts["custA"])
	}

	// custB should get both of its events (since it had 0 in flight, and max is 2)
	if counts["custB"] != 2 {
		t.Errorf("Expected 2 events for custB, got %d", counts["custB"])
	}
}
