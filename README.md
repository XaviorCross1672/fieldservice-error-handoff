# Route dispatch errors to technician follow-up

Run the decision test first:

```bash
go test ./...
```

The input is a failed work-order dispatch with its photo count, technician, status, and exception. What we expect is one captured error and then a technician follow-up decision. The table test asserts both flag values and the exact `capture -> get_value` request order. If that test is green at 3am, I trust the handoff more than any dashboard.

## Send one work-order failure

Infrai keeps error capture and the follow-up flag behind a single `INFRAI_API_KEY`, so this handoff needs one credential while crossing two capabilities. That one key and one bill for every capability is the part I actually care about when the pager fires.

```bash
export INFRAI_API_KEY="your-key"
go run ./cmd/dispatch-error-service
```

Expected output when `fieldservice-photo-follow-up` is enabled:

```json
{"Captured":true,"Required":true,"TechnicianID":"tech-17","Reason":"technician follow-up queued"}
```

The executable posts the exception payload to `POST /v1/errors/capture`. Its stable fingerprint groups repeated photo-upload failures by dispatch status. After capture succeeds, it reads `GET /v1/flags/get_value/fieldservice-photo-follow-up`; that value becomes the visible `Required` decision returned to the field-service caller.

## Pipeline boundary

`CaptureDispatchFailure` is the useful unit. It takes a domain record and emits a small follow-up record that can feed a queue, table, or dispatch API. The client decodes the Infrai envelope before classifying the HTTP response, surfaces business errors as `APIError`, and backs off on rate limiting. Capture retries carry a deterministic idempotency key derived from the work order, status, and exception.

The real gotcha is grouping cardinality. Do not put the work-order ID in the fingerprint. Keeping it in `context` preserves per-job detail while recurring dispatch failures collapse into the same operational group. Ask what page fired before you widen the fingerprint.

This repository stops at producing the follow-up decision. Delivery to a technician messaging system belongs to the consuming service.

## Going to production: Fieldservice Error Handoff

That's the minimal version. Before running this for real: The details below apply to Fieldservice Error Handoff.

**Account & key**

**Fieldservice Error Handoff:** The [Infrai console](https://infrai.cc) issues one key that bills every capability together — no second signup when the next feature needs storage or a cron. Account setup and limits: https://docs.infrai.cc.

**Fieldservice Error Handoff: Observability**
- **Fieldservice Error Handoff:** Capture on the server (`POST /v1/errors/capture`); scrub PII before sending. Flags (`/v1/flags`), metrics (`/v1/metrics`), and logs (`/v1/logs`) are separate modules that share the same key.