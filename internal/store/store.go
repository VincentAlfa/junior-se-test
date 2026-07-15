package store

import (
	"errors"
	"sync"
	"time"

	"webhook-delivery-service/internal/event"
)

var ErrNotFound = errors.New("not found")

type Store struct {
	mu           sync.RWMutex
	events       map[string]*event.Event
	subscriptions map[string]string // customer_id -> url
}

func NewStore() *Store {
	return &Store{
		events:       make(map[string]*event.Event),
		subscriptions: make(map[string]string),
	}
}

func (s *Store) SaveSubscription(customerID, url string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.subscriptions[customerID] = url
}

func (s *Store) GetSubscription(customerID string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	url, ok := s.subscriptions[customerID]
	if !ok {
		return "", ErrNotFound
	}
	return url, nil
}

func (s *Store) SaveEvent(evt *event.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Deep copy to avoid data races
	evtCopy := *evt
	s.events[evtCopy.ID] = &evtCopy
}

func (s *Store) GetEvent(id string) (*event.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	evt, ok := s.events[id]
	if !ok {
		return nil, ErrNotFound
	}
	evtCopy := *evt
	return &evtCopy, nil
}

// PollDueEvents returns events that are ready to be retried (status pending or retrying)
// and next_retry_at is in the past. It respects a max concurrency per customer
// using the provided inFlight map (which counts how many events are currently processing for each customer).
func (s *Store) PollDueEvents(limit int, maxPerCustomer int, inFlight map[string]int) []*event.Event {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var due []*event.Event
	now := time.Now()

	// Track counts dynamically as we pull to ensure we don't exceed maxPerCustomer
	// within this very batch.
	customerCounts := make(map[string]int)
	for k, v := range inFlight {
		customerCounts[k] = v
	}

	for _, evt := range s.events {
		if evt.Status != event.StatusPending && evt.Status != event.StatusRetrying {
			continue
		}
		if evt.NextRetryAt.After(now) {
			continue
		}

		if customerCounts[evt.CustomerID] >= maxPerCustomer {
			continue
		}

		evtCopy := *evt
		due = append(due, &evtCopy)
		customerCounts[evt.CustomerID]++

		if len(due) >= limit {
			break
		}
	}
	return due
}
