# Webhook Delivery Service

A robust, in-memory webhook delivery service with exponential backoff retries and isolated worker concurrency.

## Running the Service

To run both the delivery service and the mock endpoint in a single command, open two separate terminal windows or use a background process:

**Run the delivery server (Port 8080):**
```bash
go run ./cmd/server
```

**Run the mock customer endpoint (Port 9000):**
```bash
go run ./cmd/mockendpoint -mode=flaky
```
*(Modes available: `flaky`, `fail`, `succeed`)*

## API Examples

**1. Register a subscription:**
```bash
curl -X POST http://localhost:8080/subscriptions \
  -H "Content-Type: application/json" \
  -d '{"customer_id": "cust_123", "url": "http://localhost:9000/receive"}'
```

**2. Send an event:**
```bash
curl -X POST http://localhost:8080/events \
  -H "Content-Type: application/json" \
  -d '{"customer_id": "cust_123", "payload": {"foo": "bar"}}'
```
*Note the returned `event_id`.*

**3. Check event status:**
```bash
curl http://localhost:8080/events/<EVENT_ID>
```
