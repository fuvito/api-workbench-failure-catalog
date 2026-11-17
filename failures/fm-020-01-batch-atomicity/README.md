# fm-020-01 — Batch Operation Atomicity

**Difficulty:** Medium | **Detected by:** `type: status` assertion | **Port:** 19020

**Primary failure mode:** [FM-020 — Distributed Transaction and Partial-Workflow Failures](https://clrsig.com/failures/fm-020/)  

## What this failure is

A batch create endpoint partially succeeds instead of rolling back atomically. When a batch contains one invalid item (empty `name`), the broken server creates the valid ones and skips the invalid one, returning `201`. The spec implies all-or-nothing — the endpoint must return `422` and create nothing if any item is invalid.

## Reproduce it

```bash
go run failures/fm-020-batch-atomicity/broken/main.go &
clrsig run failures/fm-020-batch-atomicity/scenarios/detect.yaml \
  --env failures/fm-020-batch-atomicity/environments/local.yaml
```

Expected: **exits 1** — broken server returns `201` (partial success); assertion expects `422`.

## Why clrsig catches it

Step 1 asserts `status 422` on the batch POST with one invalid item. Broken returns `201` → fails. Step 2 would additionally check `$.count == 0` but is skipped after step 1 fails.

## Fix

```bash
go run failures/fm-020-batch-atomicity/fixed/main.go &
clrsig run failures/fm-020-batch-atomicity/scenarios/detect.yaml \
  --env failures/fm-020-batch-atomicity/environments/local.yaml
# exits 0
```

**What changed:** `fixed/main.go` validates all items before creating any; returns `422` on first validation failure.

## CI

```yaml
- { id: "020", slug: "batch-atomicity", kind: "server" }
```
