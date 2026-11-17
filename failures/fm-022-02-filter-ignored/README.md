# fm-022-02 — Filter Parameter Ignored

**Difficulty:** Easy | **Detected by:** `type: jsonpath` assertion | **Port:** 19028

**Primary failure mode:** [FM-022 — Syntactically Valid but Semantically Invalid Payloads](https://clrsig.com/failures/fm-022/)  

## What this failure is

The `status` query parameter is accepted but ignored — the server always returns all orders regardless of the filter value. With 1 pending and 2 completed orders in the data set, `?status=completed` should return 2 and `?status=pending` should return 1.

## Reproduce it

```bash
go run failures/fm-028-filter-ignored/broken/main.go &
clrsig run failures/fm-028-filter-ignored/scenarios/detect.yaml \
  --env failures/fm-028-filter-ignored/environments/local.yaml
```

Expected: **exits 1** — `?status=completed` returns `count: 3` (all orders); assertion expects `2`.

## Why clrsig catches it

Both steps assert `$.count` via jsonpath. The broken server returns `3` regardless of filter, failing both.

## Fix

```bash
go run failures/fm-028-filter-ignored/fixed/main.go &
clrsig run failures/fm-028-filter-ignored/scenarios/detect.yaml \
  --env failures/fm-028-filter-ignored/environments/local.yaml
# exits 0
```

**What changed:** `fixed/main.go` filters the order slice by `Status == statusFilter` before responding.

## CI

```yaml
- { id: "028", slug: "filter-ignored", kind: "server" }
```
