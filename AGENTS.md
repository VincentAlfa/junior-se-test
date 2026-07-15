# AGENTS.md

Context for AI coding agents (Claude Code, Copilot, etc.) working in this repo.
Read PRD.md first for the full problem statement and constraints.

## What this project is

A take-home assessment: a webhook delivery service in Go. Two runnable programs,
one module. Optimized for clear reasoning and demonstrable failure-handling, not
feature breadth or line count.

## Stack

- Go (matches the hiring company's stack — deliberate choice)
- `net/http` — no web framework needed, this project doesn't need one
- SQLite via `modernc.org/sqlite` (pure Go, no cgo) — or in-memory if we decide
  to defer persistence; decision lives in DECISIONS.md, not silently changed
- No external job-queue library — retry/backoff scheduling should be implemented
  directly (goroutines + channels or DB-polled scheduler) since that IS the point
  of the exercise

## Project layout

```
cmd/server/main.go        entrypoint for the webhook service (:8080)
cmd/mockendpoint/main.go  entrypoint for the fake flaky customer endpoint (:9000)
internal/event/           event model + status state machine
internal/delivery/        retry/backoff/worker logic
internal/store/           persistence layer
internal/api/             HTTP handlers
DECISIONS.md              half-page write-up of the 4 required design decisions
README.md                 one-line run command
```

## Non-negotiable constraints

- No auth, no UI, no billing, no production deployment tooling. Do not add these
  even if it seems "more complete" — it's explicitly out of scope per PRD.md §4.
- Do not reach for a message-queue dependency (Redis, RabbitMQ, etc.) — the
  in-process retry/worker design is the thing being evaluated.
- Do not silently pick a delivery guarantee, retry policy, isolation strategy, or
  ordering behavior. These four are named in PRD.md §7. If asked to implement
  retry logic, surface the trade-off in conversation before hardcoding it.
- Prefer a smaller, well-reasoned implementation over an exhaustive one. If a
  feature isn't clearly required by PRD.md, ask before building it.
- Tests: a few meaningful ones (e.g. retry backoff timing, isolation between
  customers) are worth more here than broad coverage. Don't chase a percentage.

## Working style

- I want ready-to-run code with brief reasoning up front, not just a code dump.
- When there's a design fork (e.g. "sleep in the worker" vs "schedule + poll"),
  state the trade-off in 2-3 sentences before picking one.
- Keep logging visible and readable — logs are the demo. Every retry attempt
  should log customer/event ID, attempt number, outcome, and next retry time.
- No em dashes in comments/docs, no filler, no AI-sounding boilerplate text.

## Definition of done for any change

Matches PRD.md §8. If a change doesn't move toward one of those checklist items,
flag it rather than building it.