# Pipeline Starter

## Design Intent

This project is a small, in-process pipeline framework written in Go. It demonstrates clean interfaces, stage chaining, schema-based validation, and graceful shutdown.

The focus is on **clarity, testability, and explicit tradeoffs**, not production-scale reliability.

---

## Scope and Non-Goals

This project intentionally avoids production concerns such as retries, delivery guarantees, persistence, idempotency, and throughput tuning.

Errors are handled in a simple **fail-fast** manner and messages are processed entirely in memory. These choices keep behavior easy to reason about while clearly showing how the system could be extended.

In this MVP, each pipeline instance reads from one source. To handle multiple sources, run multiple pipeline instances or add a combined source in a later version.

---

## High-Level Design

```
Source -> Validation Stage -> Transform Stage -> Destination
```

* Components communicate through channels.
* `context.Context` controls lifecycle and shutdown.

---

## Execution Model

`SliceSource.Start` runs in its own goroutine and writes messages to a channel.
`Pipeline.Run` reads from that channel and processes messages one at a time.

Flow:

```
Source goroutine -> message channel -> Pipeline.Run -> stages -> destination
```

What this means:

* The source is asynchronous.
* The pipeline is sequential (one message in flight at a time).
* Processing is pull-paced by `Pipeline.Run`.
* If downstream is slow, upstream naturally blocks on channel send.

---

## Key Design Decisions

### Interfaces for `Source`, `Stage`, and `Destination`

Interfaces keep the pipeline decoupled from specific implementations. This allows components to be tested independently and makes it easy to add new sources, stages, or destinations.
The tradeoff is a small amount of extra boilerplate.

---

### Channels between components

Channels provide safe message passing and built-in flow control. If downstream processing slows, upstream sends block naturally, keeping memory usage predictable.
Channel-based code can be harder to debug than fully synchronous code, but keeps the concurrency model simple.

---

### Validation as a separate stage

Validation is implemented as its own stage instead of being embedded in the source. This keeps responsibilities clear and allows validation logic to be reused across pipelines, at the cost of one additional processing step.

---

### In-memory schema selection by `type` + `version`

Schemas are stored in a simple in-memory map keyed by message type and version. This makes schema selection explicit and easy to follow while meeting the exercise requirements.
Schemas are static and reset on process restart.

---

## Failure, Termination, and Robustness (MVP)

This implementation prioritizes predictable behavior over automatic recovery.

* **Fail-fast errors:** Any source, stage, or destination error immediately stops the pipeline.
* **Graceful shutdown:** All components respect `context.Context`. Cancellation stops processing and ensures no goroutines are left running.
* **No partial processing:** Messages are processed one at a time and either fully complete or are dropped on failure.
* **Built-in flow control:** Channels naturally prevent unbounded buffering.

For this MVP, “robustness” means **clear failure semantics and clean termination**, not production-grade recovery.

---

## Development Process

The system was built iteratively with an emphasis on clarity and testability:

1. Clarified scope and non-goals.
2. Defined core interfaces (`Source`, `Stage`, `Destination`).
3. Built minimal concrete components to validate end-to-end flow.
4. Implemented `Pipeline.Run` with fail-fast behavior and context cancellation.
5. Added component-level and pipeline-level tests.
6. Introduced schema validation using message `type + version`.
7. Documented decisions, limitations, and extension paths.

---

## Requirements

* Go 1.22+ (see `go.mod`)

---

## How To Run

```bash
go test ./...
go run .
```

---

## Example Pipeline

A working example is included in `main.go`.

The example wires:

```
SliceSource -> ValidationStage -> SetMetaStage -> StdoutDestination
```

It emits demo messages, validates each message using schema selection by `type+version`, applies a metadata transform, and prints the final output.

---

## Current Validation Rule

* Schemas are selected by `Message.Type` + `Message.Version`.
* Schema format is minimal: `RequiredFields []string`.
* Validation checks required payload keys exist.

---

## Next V1 (With More Time)

* **Retries:** fixed retry count at destination boundary for retryable errors.
* **Idempotency:** message IDs with in-memory deduplication.
* **Failure handling:** optional dead-letter destination for invalid messages.

---

## Extension Path

1. Add richer schema rules.
2. Add additional transform stages.
3. Change error policy from fail-fast to drop or dead-letter if needed.
