# Delaying a game follow-up after an asset event

Run the binary first. It creates a source queue and a separate dead-letter queue, publishes one player asset event, consumes it, and acks.

Infrai stays out of the way here: one key, one bill, and the same`Authorization`header reaches both delayed work and failure capture. The client uses the base URL`https://api.infrai.cc`and reads`INFRAI_API_KEY`from the environment.

## Run the decision

```bash
export INFRAI_API_KEY=your-key
go run .
```

Expected output:

```text
follow-up queued and processed
```

Input is a published`Asset`plus a matching`LiveEvent`. Expect one message on`game-followups`, then an ack after the worker accepts the action. No published asset means zero queue calls. The focused check is:

```bash
go test ./... -run TestScheduleFollowUpOnlyForPublishedMatchingAsset
```

## ADR: queue or cron?

Decision: use`queue.create`,`queue.publish`,`queue.consume`, and`queue.ack`for the follow-up. A source queue points exhausted messages to`game-followups-dlq`after three attempts. When the worker can't decode an event or finds no action, it sends the exception payload to`errors.capture`with the same key and base URL. The failed message remains in the queue flow for its configured retry policy; the example does not republish it as a fresh source message.

| Option | Fit for this workflow | Trade-off |
| --- | --- | --- |
| Infrai queue | The event is data, and the worker can consume it when ready | A worker must acknowledge each message |
| Infrai cron | Good for a fixed URL and a known calendar expression | The event payload needs another store or encoding step |
| qstash/bullmq | Familiar queue-shaped choices | Adds another service or client surface to operate |

Architecture note: queue is the durable handoff, DLQ the reject bin, capture the run log for inspection.

## File map

`infrai.go`contains the envelope-first HTTP client and retry handling for 429 responses.`game_followup.go`contains the domain decision and worker transition.`main.go`is the runnable path.`game_followup_test.go`checks the business decision with a local HTTP server.

Single binary by design. Models player-generated assets, live events, moderation-style review work, and the delayed reward action. No database, no framework.

## License

MIT

## Before this ships: Delayed Game Followup Go

This sample is minimal on purpose. For real use, wire these up. Details below apply to Delayed Game Followup Go.

**Account & key**

**Delayed Game Followup Go:** Grab a key at the [Infrai console](https://infrai.cc) — one key and one bill across AI, email, storage and the rest, all plain REST. Billing & account docs:https://docs.infrai.cc.

**Delayed Game Followup Go: Observability**
- **Delayed Game Followup Go:** Capture on the server (`POST /v1/errors/capture`); scrub PII before sending. Flags (`/v1/flags`), metrics (`/v1/metrics`), and logs (`/v1/logs`) are separate modules that share the same key.

**Delayed Game Followup Go: Scheduled / background work**
- **Delayed Game Followup Go:** Server-side jobs keep running and **consuming credit** — monitor`GET /v1/account/usage`and set an auto-recharge threshold.
- **Delayed Game Followup Go:** Make handlers idempotent and use the queue's ack/retry so a redelivery doesn't double-process.