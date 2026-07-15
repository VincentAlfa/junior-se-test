# PRD — Webhook Delivery Service

## 1. Context

This is a take-home assessment (AccelByte, Junior SE). Evaluated on engineering
judgment and reasoning about failure/retry/ordering trade-offs — not framework
choice, code volume, or test coverage percentage. Time budget: ~3 hours. If the
cap is hit, stop and document what's left in DECISIONS.md rather than pushing on.

## 2. Problem Statement

Build a service that:
1. Accepts events from internal producers via `POST /events` (customer/subscription
   ID + JSON payload).
2. Delivers each event to that customer's registered HTTP endpoint via POST.
3. Retries delivery when the endpoint is down, slow, or erroring.
4. Exposes delivery status per event: `pending` / `retrying` / `delivered` / `failed`.

A registered endpoint can be slow, erroring, or down for hours. Events must not be
lost during that window.

## 3. Goals

- Correct, defensible delivery semantics (documented, not just implemented).
- Retry logic with backoff and a clear terminal failure state.
- One customer's dead endpoint must not block or slow delivery to other customers.
- Delivery status is queryable per event at any time.
- Runnable locally with a single command, backed by a real (or well-justified
  in-memory) datastore.

## 4. Non-Goals (explicitly out of scope)

- Auth / login system
- Any UI / frontend
- Multi-tenant billing
- Production deployment (cloud infra, CI/CD, containers beyond local convenience)
- Exhaustive test coverage — a few meaningful tests over many shallow ones

## 5. System Overview

Two runnable programs in one Go module:

- **`cmd/server`** — the webhook delivery service (port 8080). Owns event ingestion,
  persistence, retry scheduling, and status API.
- **`cmd/mockendpoint`** — a throwaway stand-in for "the customer's website" (port
  9000). Deliberately flaky: configurable to always fail, fail N times then
  succeed, be slow, or behave randomly. Exists purely to prove retry behavior in
  the demo.

## 6. API Surface (draft, subject to revision during build)

```
POST /subscriptions
  { "customer_id": "abc", "url": "http://localhost:9000/receive" }
  → registers where a customer's events should be delivered

POST /events
  { "customer_id": "abc", "payload": { ... } }
  → 202 Accepted, { "event_id": "...", "status": "pending" }

GET /events/{id}
  → { "event_id": "...", "status": "pending|retrying|delivered|failed",
      "attempts": N, "last_attempt_at": "...", "next_retry_at": "..." }
```

## 7. Design Decisions Requiring Explicit Reasoning

These have no single right answer. Each must be a deliberate, defensible choice,
written up in DECISIONS.md:

1. **Delivery guarantee** — at-least-once vs at-most-once vs exactly-once, and
   what it demands of the customer (e.g. idempotency on their end).
2. **Retry policy** — attempt count, backoff schedule (fixed vs exponential),
   terminal state when an endpoint never recovers.
3. **Long outage isolation** — what happens to one customer's queue during a
   6-hour outage, and why it doesn't degrade delivery to everyone else.
4. **Ordering** — whether events for a given customer are delivered in the order
   they were produced, and whether that's a guarantee worth making here.

## 8. Definition of Done

- [ ] `go run ./cmd/server` and `go run ./cmd/mockendpoint` both start cleanly
- [ ] Can register a subscription, POST an event, and GET its status
- [ ] Retry behavior is visible in logs (attempt count, backoff delay, outcome)
- [ ] A permanently-down endpoint does not block delivery to a different customer
      (demonstrable, not just asserted)
- [ ] DECISIONS.md written covering section 7
- [ ] One-line run command documented in README.md

## 9. Explicitly Deferred (acceptable to leave out, note in DECISIONS.md)

- Persistent durability across process restarts (in-memory store is acceptable
  if justified)
- Webhook signing / payload verification
- Rate limiting / backpressure on `/events`
- Dead-letter queue tooling beyond a `failed` status