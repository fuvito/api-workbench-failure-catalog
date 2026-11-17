# FM-026 — Async Job Status Never Transitions

**Difficulty:** Easy | **Detected by:** `type: jsonpath` assertion | **Port:** 19026

## What this failure is

Jobs are created but never transition beyond `"processing"`. The worker that advances job state was never implemented — the status is set on creation and never updated.

## Reproduce it

```bash
go run failures/fm-026-async-job/broken/main.go &
clrsig run failures/fm-026-async-job/scenarios/detect.yaml \
  --env failures/fm-026-async-job/environments/local.yaml
```

Expected: **exits 1** — `GET /jobs/1` returns `{"status":"processing"}`; assertion expects `"completed"`.

## Why clrsig catches it

Step 2 asserts `$.status == "completed"`. The broken server always returns `"processing"`, failing the jsonpath assertion.

## Fix

```bash
go run failures/fm-026-async-job/fixed/main.go &
clrsig run failures/fm-026-async-job/scenarios/detect.yaml \
  --env failures/fm-026-async-job/environments/local.yaml
# exits 0
```

**What changed:** `fixed/main.go` sets status to `"completed"` immediately on creation (synchronous simulation).

## CI

```yaml
- { id: "026", slug: "async-job", kind: "server" }
```
