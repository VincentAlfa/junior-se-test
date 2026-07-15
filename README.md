# Webhook Delivery Service

A robust, in-memory webhook delivery service with exponential backoff retries and isolated worker concurrency.

## Running the Service

The assignment requires a **one-line run command**. You can run both the server and the mock endpoint simultaneously using this single command:

**For Mac/Linux (Bash/Zsh):**
```bash
go run ./cmd/mockendpoint -mode=flaky & go run ./cmd/server
```

**For Windows (PowerShell):**
```powershell
Start-Job { go run ./cmd/mockendpoint -mode=flaky }; go run ./cmd/server
```

*(This starts the mock endpoint in the background on port 9000, and runs the main delivery service in the foreground on port 8080).*

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
