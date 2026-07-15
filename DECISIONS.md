# Design Decisions

This document outlines the core architectural and semantic decisions made for the webhook delivery service, as well as the trade-offs considered.

## 1. Delivery Guarantee: At-Least-Once

**Decision:** The service provides an *at-least-once* delivery guarantee.
**Reasoning:** In a distributed system, network requests can fail in ways that make it impossible for the sender to know if the receiver successfully processed the payload (e.g., the customer's server processes the event but crashes before sending a 200 OK, or a network partition drops the ACK). To prevent data loss, the service must retry on timeouts or missing ACKs. Consequently, customers might receive the same event multiple times and are expected to build idempotent webhook handlers. Exactly-once delivery is not feasible over standard HTTP without complex two-phase commit protocols, and at-most-once would result in silent drops during transient outages.

## 2. Retry Policy: Exponential Backoff

**Decision:** Failed deliveries are retried up to 5 times (total of 5 attempts) using an exponential backoff schedule (5s, 25s, 2m 5s, ~10m).
**Reasoning:** 
- **Backoff Schedule:** The exponential increase allows for quick recovery from momentary glitches (like a short network blip or quick deployment restart) while avoiding hammering a service that is down for an extended period (reducing load on both our service and their recovering infrastructure).
- **Terminal State:** If the endpoint fails on the 5th attempt, the event transitions to a terminal `failed` state. The webhook is dropped to prevent unbounded queue growth and resource exhaustion on our end.

## 3. Long Outage Isolation

**Decision:** The delivery worker uses asynchronous polling and bounded per-customer concurrency, rather than blocking worker goroutines with `time.Sleep`.
**Reasoning:** If a customer's endpoint goes offline for hours, naive worker pools can quickly become exhausted as every worker thread blocks on a failing request or a long sleep, starving other customers' events. 
To achieve true isolation:
1. **No Sleeping Workers:** Delivery attempts are scheduled by setting `NextRetryAt`. A background poller queries the store for events ready to be sent and dispatches them to a non-blocking worker pool.
2. **Short Timeouts:** HTTP clients are configured with a strict 5-second timeout to ensure threads are quickly returned to the pool.
3. **Concurrency Limiting:** The poller tracks in-flight requests per customer. If a customer hits their concurrency cap (e.g., 10 concurrent requests), the poller simply skips their queued events during that cycle, ensuring capacity is always preserved for healthy customers.

## 4. Ordering

**Decision:** Events are delivered concurrently and unordered.
**Reasoning:** Enforcing strict FIFO ordering creates head-of-line blocking. If a customer's first event fails, pausing all subsequent events until the first one succeeds (or fails permanently 10 minutes later) severely degrades throughput and latency for fresh events. Instead, events are pushed out as fast as possible. Customers who require strict ordering must rely on sequence numbers or the `created_at` timestamp embedded in the payload to reassemble order on their side.

## 5. Deliberately Deferred (Out of Scope)

Per PRD §9, the following have been excluded from this implementation to focus on the core retry mechanics:
- **Persistent Storage:** The `store` layer uses an in-memory `sync.RWMutex` protected map. While SQLite was considered, the in-memory approach provides the cleanest demonstration of retry logic without adding unnecessary I/O setup overhead. Data will be lost on process restart.
- **Authentication & Webhook Signing:** Not implemented.
- **Rate Limiting:** No backpressure on the `/events` ingestion endpoint.
