package delivery

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"webhook-delivery-service/internal/event"
	"webhook-delivery-service/internal/store"
)

type Worker struct {
	store      *store.Store
	httpClient *http.Client
	mu         sync.Mutex
	inFlight   map[string]struct{} // tracks event ID
	inFlightC  map[string]int      // tracks customer ID -> count
}

func NewWorker(s *store.Store) *Worker {
	return &Worker{
		store: s,
		httpClient: &http.Client{
			Timeout: 5 * time.Second, // strict 5s timeout
		},
		inFlight:  make(map[string]struct{}),
		inFlightC: make(map[string]int),
	}
}

func (w *Worker) Start(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.poll()
		}
	}
}

func (w *Worker) poll() {
	w.mu.Lock()
	inFlightCopy := make(map[string]int)
	for k, v := range w.inFlightC {
		inFlightCopy[k] = v
	}
	inFlightECopy := make(map[string]struct{})
	for k := range w.inFlight {
		inFlightECopy[k] = struct{}{}
	}
	w.mu.Unlock()

	// ponytail: Worker pool is theoretically unbounded (spawns 1 goroutine per due event). Ceiling: Goroutine explosion if limit is set very high and endpoints are slow. Upgrade path: bounded worker pool pattern using channels.
	// limit to pulling e.g. 100 events, max 10 per customer
	dueEvents := w.store.PollDueEvents(100, 10, inFlightCopy, inFlightECopy)

	w.mu.Lock()
	for _, evt := range dueEvents {
		if _, ok := w.inFlight[evt.ID]; !ok {
			w.inFlight[evt.ID] = struct{}{}
			w.inFlightC[evt.CustomerID]++
			go w.deliver(evt)
		}
	}
	w.mu.Unlock()
}

func (w *Worker) deliver(evt *event.Event) {
	defer func() {
		w.mu.Lock()
		delete(w.inFlight, evt.ID)
		w.inFlightC[evt.CustomerID]--
		w.mu.Unlock()
	}()

	url, err := w.store.GetSubscription(evt.CustomerID)
	if err != nil {
		log.Printf("Event %s failed: subscription not found for customer %s", evt.ID, evt.CustomerID)
		evt.Status = event.StatusFailed
		evt.LastAttemptAt = time.Now()
		w.store.SaveEvent(evt)
		return
	}

	req, _ := http.NewRequest(http.MethodPost, url, bytes.NewReader(evt.Payload))
	req.Header.Set("Content-Type", "application/json")

	attemptNum := evt.Attempts + 1
	log.Printf("Delivery attempt %d for event %s to %s (customer: %s)", attemptNum, evt.ID, url, evt.CustomerID)

	resp, err := w.httpClient.Do(req)

	success := err == nil && resp.StatusCode >= 200 && resp.StatusCode < 300

	if resp != nil {
		resp.Body.Close()
	}

	evt.Attempts = attemptNum
	evt.LastAttemptAt = time.Now()

	if success {
		log.Printf("Event %s delivered successfully on attempt %d", evt.ID, attemptNum)
		evt.Status = event.StatusDelivered
	} else {
		if err != nil {
			log.Printf("Event %s delivery failed on attempt %d: error %v", evt.ID, attemptNum, err)
		} else {
			log.Printf("Event %s delivery failed on attempt %d: status %d", evt.ID, attemptNum, resp.StatusCode)
		}

		if attemptNum >= 5 {
			evt.Status = event.StatusFailed
			log.Printf("Event %s reached max attempts (%d), marked as failed", evt.ID, attemptNum)
		} else {
			evt.Status = event.StatusRetrying
			// Backoff: 5s, 25s, 125s, 625s
			multiplier := 1
			for i := 1; i < attemptNum; i++ {
				multiplier *= 5
			}
			delay := time.Duration(5 * multiplier) * time.Second
			evt.NextRetryAt = evt.LastAttemptAt.Add(delay)
			log.Printf("Event %s scheduled for retry in %s (at %s)", evt.ID, delay, evt.NextRetryAt.Format(time.RFC3339))
		}
	}

	w.store.SaveEvent(evt)
}
