package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"webhook-delivery-service/internal/event"
	"webhook-delivery-service/internal/store"
)

type Handler struct {
	store *store.Store
}

func NewHandler(s *store.Store) *Handler {
	return &Handler{
		store: s,
	}
}

type subscriptionReq struct {
	CustomerID string `json:"customer_id"`
	URL        string `json:"url"`
}

func (h *Handler) HandleSubscriptions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req subscriptionReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.CustomerID == "" || req.URL == "" {
		http.Error(w, "customer_id and url are required", http.StatusBadRequest)
		return
	}

	h.store.SaveSubscription(req.CustomerID, req.URL)
	w.WriteHeader(http.StatusCreated)
}

type eventReq struct {
	CustomerID string          `json:"customer_id"`
	Payload    json.RawMessage `json:"payload"`
}

type eventResp struct {
	EventID string       `json:"event_id"`
	Status  event.Status `json:"status"`
}

func (h *Handler) HandleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req eventReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.CustomerID == "" {
		http.Error(w, "customer_id is required", http.StatusBadRequest)
		return
	}

	evt := &event.Event{
		ID:          uuid.New().String(),
		CustomerID:  req.CustomerID,
		Payload:     req.Payload,
		Status:      event.StatusPending,
		CreatedAt:   time.Now(),
		NextRetryAt: time.Now(), // ready immediately
	}

	h.store.SaveEvent(evt)

	resp := eventResp{
		EventID: evt.ID,
		Status:  evt.Status,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) HandleGetEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Assuming URL path is /events/{id}
	// A simple way without a router is to strip the prefix
	id := r.URL.Path[len("/events/"):]
	if id == "" {
		http.Error(w, "event id required", http.StatusBadRequest)
		return
	}

	evt, err := h.store.GetEvent(id)
	if err != nil {
		if err == store.ErrNotFound {
			http.Error(w, "event not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(evt)
}
