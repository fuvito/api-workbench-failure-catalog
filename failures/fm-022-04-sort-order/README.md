# fm-022-04 — Sort Order Contract

> Full catalog entry: https://clrsig.com/failures/fm-035/

**Difficulty:** Easy
**Detected by:** `clrsig run` with `type: jsonpath` assertions
**Port:** 19035

**Primary failure mode:** [FM-022 — Syntactically Valid but Semantically Invalid Payloads](https://clrsig.com/failures/fm-022/)  
**Also illustrates:** [FM-003 — Pagination Contract Failures](https://clrsig.com/failures/fm-003/)  

## What this failure is

The server accepts a `sort` query parameter but ignores it, always returning results in insertion order. The spec declares `sort` as an accepted parameter with values `price_asc`, `price_desc`, and `name_asc`, implying those sort orders will be honoured.

This pattern appears when:
- A developer adds the parameter to the spec during design but forgets to implement it
- Sorting was removed as a "performance optimization" without updating the spec
- A caching layer returns a cached (unsorted) response regardless of query parameters

The consequence: consumers that build UIs with sort controls observe that the order never changes. The bug is silent — the server returns 200 with valid data, just in the wrong order.

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-035-sort-order/broken/main.go &
clrsig run failures/fm-035-sort-order/scenarios/detect.yaml \
  --env failures/fm-035-sort-order/environments/local.yaml
```

Expected: **clrsig exits 1** — `sort=price_asc` returns `$.items[0].price = 79.99` (insertion order) instead of `29.99`; `sort=price_desc` also returns `79.99` instead of `349.99`.

Stop the server:

```bash
# macOS / Linux
fuser -k 19035/tcp

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19035).OwningProcess
```

## Why clrsig catches it

The scenario uses `type: jsonpath` assertions to check the first item's price in each sort order:

```yaml
- type: jsonpath
  path: "$.items[0].price"
  operator: eq
  expected: 29.99
```

The broken server always returns `79.99` as the first item (insertion order), failing both assertions.

## Fix

```bash
go run failures/fm-035-sort-order/fixed/main.go &
clrsig run failures/fm-035-sort-order/scenarios/detect.yaml \
  --env failures/fm-035-sort-order/environments/local.yaml
# exits 0
```

**What changed:** `broken/main.go` — added a `switch` on the `sort` query parameter that copies and sorts the product slice before serializing.

## GitHub Actions CI setup

```yaml
- { id: "035", slug: "sort-order", kind: "server" }
```

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-035-sort-order\broken\
& "$env:TEMP\server-broken.exe"
# In a second terminal:
clrsig run failures/fm-035-sort-order/scenarios/detect.yaml `
  --env failures/fm-035-sort-order/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19035).OwningProcess

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-035-sort-order\fixed\
& "$env:TEMP\server-fixed.exe"
# In a second terminal:
clrsig run failures/fm-035-sort-order/scenarios/detect.yaml `
  --env failures/fm-035-sort-order/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
```

## In the wild

This failure appears in:

- **Spec-first APIs** where the parameter was designed before implementation and the implementation never arrived
- **Cached list endpoints** where the caching key doesn't include query parameters, so all sort variants return the same cached response
- **Search engines** where relevance score ordering is hardcoded and the `sort` parameter silently has no effect
- **Database queries** without an `ORDER BY` clause — results are returned in an undefined heap order that happens to be stable in development but changes under load
