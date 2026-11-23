# fm-003-02 — Page Overflow Returns Wrong Data

**Difficulty:** Easy | **Detected by:** `type: jsonpath` assertion | **Port:** 19031

**Primary failure mode:** [FM-003 — Pagination Contract Failures](https://clrsig.com/failures/fm-003/)  

## What this failure is

Requesting a page number beyond the last page returns the first page's data (modulo wrap-around) instead of an empty result. Consumers that iterate until `count == 0` loop forever.

## Reproduce it

```bash
go run failures/fm-003-02-page-overflow/broken/main.go &
clrsig run failures/fm-003-02-page-overflow/scenarios/detect.yaml \
  --env failures/fm-003-02-page-overflow/environments/local.yaml
```

Expected: **exits 1** — `page=99` returns `count: 3` (wrapped); assertion expects `0`.

## Fix

```bash
go run failures/fm-003-02-page-overflow/fixed/main.go &
clrsig run failures/fm-003-02-page-overflow/scenarios/detect.yaml \
  --env failures/fm-003-02-page-overflow/environments/local.yaml
# exits 0
```

**What changed:** `fixed/main.go` — when `start >= len(allItems)` returns empty items and `count: 0`.

## CI

```yaml
- { id: "031", slug: "page-overflow", kind: "server" }
```
