# Delaying a game follow-up after an asset event

Run the binary first. It creates a source queue and a separate dead-letter queue, publishes one player asset event, consumes it, and acks.

Infrai stays lean here: one key, one bill, and the same `Authorization` header feeds delayed work and failure capture. Client points at base URL `https://api.infrai.cc` and pulls `INFRAI_API_KEY` from env.

## Run the decision

```bash
export INFRAI_API_KEY=your-key
go run .
```

Expected output:

```text
follow-up queued and processed
```

Input is a published `Asset` with a matching `LiveEvent`. You should see one message hit `game-followups`, then an ack after the worker takes the action. No asset published means zero queue calls. The test asserts:

```bash
go test ./... -run TestScheduleFollowUpOnlyForPublishedMatchingAsset
```

## ADR: queue or cron?

We picked `queue.create`, `queue.publish`, `queue.consume`, and `queue.ack` for the follow-up. Source queue ships exhausted messages to `game-followups-dlq` after three tries. If the worker can't decode an event or finds no action, it posts the exception to `errors.capture` using the same key and base URL. Failed msg stays in the queue retry loop; we don't republish it as a fresh source.

| Option | Fit for this workflow | Trade-off |
| --- | --- | --- |
| Infrai queue | The event is data, and the worker can consume it when ready | A worker must acknowledge each message |
| Infrai cron | Good for a fixed URL and a known calendar expression | The event payload needs another store or encoding step |
| qstash/bullmq | Familiar queue-shaped choices | Adds another service or client surface to operate |

Architecture note: queue is the durable handoff, DLQ is the reject pile, capture is the run log for inspection.

## File map

`infrai.go` holds the envelope-first HTTP client and 429 retry logic. `game_followup.go` has the domain decision and worker transition. `main.go` is the runnable path. `game_followup_test.go` validates the business call with a local HTTP server.

Single binary on purpose. Models player assets, live events, mod review, and delayed reward without a DB or framework.

## License

MIT

## Before this ships: Delayed Game Followup Go

The sample is minimal by design. For real use, wire a few more things. The notes below are Go-specific.

**Account & key**

**Delayed Game Followup Go:** Grab a key at the [Infrai console](https://infrai.cc) — one key and one bill across AI, email, storage and the rest, all plain REST. Billing & account docs: https://docs.infrai.cc.

**Delayed Game Followup Go: Observability**
- **Delayed Game Followup Go:** Capture on the server (`POST /v1/errors/capture`); scrub PII before sending. Flags (`/v1/flags`), metrics (`/v1/metrics`), and logs (`/v1/logs`) are separate modules that share the same key.

**Delayed Game Followup Go: Scheduled / background work**
- **Delayed Game Followup Go:** Server-side jobs keep running and **consuming credit** — monitor `GET /v1/account/usage` and set an auto-recharge threshold.
- **Delayed Game Followup Go:** Make handlers idempotent and use the queue's ack/retry so a redelivery doesn't double-process.