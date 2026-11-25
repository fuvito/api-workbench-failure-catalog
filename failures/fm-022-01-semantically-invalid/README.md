# fm-022-01 — Syntactically Valid but Semantically Invalid Payloads

> Full catalog entry: https://clrsig.com/failures/fm-022/

**Difficulty:** Medium
**Detected by:** `clrsig run` jsonpath value assertions
**Port:** 19022

**Primary failure mode:** [FM-022 — Syntactically Valid but Semantically Invalid Payloads](https://clrsig.com/failures/fm-022/)  

## What this failure is

The server returns a syntactically valid JSON response with all required fields present and correctly typed — schema validation passes without complaint. The semantic error is in the math: `total` is set to `unit_price` instead of `unit_price × quantity`.

A consumer that checks the schema sees a `number` in `total` and considers the response valid. A consumer that checks the *value* — `total == unit_price * quantity` — immediately sees the discrepancy. This is the class of bugs that only domain-aware assertions catch.

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-022-01-semantically-invalid/broken/main.go &
clrsig run failures/fm-022-01-semantically-invalid/scenarios/detect.yaml \
  --env failures/fm-022-01-semantically-invalid/environments/local.yaml
```

Expected: **clrsig exits 1** — both orders return `total` equal to `unit_price` (10.00 and 24.99) instead of the correct products (30.00 and 49.98).

Stop the server before running the fixed version:

```bash
# macOS / Linux
pkill -f 'broken/main.go'

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19022).OwningProcess
```

## Why clrsig catches it

The scenario uses `jsonpath` assertions to check computed values directly:

```yaml
- name: "order 1: total must equal unit_price × quantity (3 × 10.00 = 30.00)"
  request:
    method: GET
    path: /orders/1
  assertions:
    - type: jsonpath
      path: $.total
      expected: 30.00
```

The assertion extracts `$.total` from the response and compares it to the expected product. When `total` is `10.00` instead of `30.00`, clrsig flags the mismatch and exits 1.

Schema validation alone would not catch this — both `10.00` and `30.00` satisfy `type: number`.

## Fix

```bash
go run failures/fm-022-01-semantically-invalid/fixed/main.go &
clrsig run failures/fm-022-01-semantically-invalid/scenarios/detect.yaml \
  --env failures/fm-022-01-semantically-invalid/environments/local.yaml
# exits 0
```

**What the demo does:** corrects the hardcoded `Total` values to `30.00` (3 × 10.00) and `49.98` (2 × 24.99). Both jsonpath assertions pass.

**What the production fix looks like** — compute `total` at write time rather than storing a potentially stale copy:

```go
// Always derive total from the source fields — never store it independently
func (o Order) ComputedTotal() float64 {
    return float64(o.Quantity) * o.UnitPrice
}
```

Or add a database constraint / application-layer validation that rejects writes where `total != unit_price * quantity`.

## Adapt this to your API

Replace `baseUrl` in `environments/local.yaml` and write jsonpath assertions for every derived or computed field in your API:

```yaml
steps:
  - name: invoice total must equal sum of line items
    request:
      method: GET
      path: /invoices/42
    assertions:
      - type: jsonpath
        path: $.total_amount
        expected: 149.97   # 3 × 49.99
```

Any field whose value is derivable from other fields in the response is a candidate for a jsonpath assertion. If the math is wrong, schema validation will never tell you.

## GitHub Actions CI setup

This sub-project uses `kind: "server"` — same CI flow as FM-001, FM-005, FM-010, FM-016.

### Matrix entry

```yaml
- { id: "022", slug: "semantically-invalid", kind: "server" }
```

Add to both `catalog.yml` and `catalog-main.yml`.

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-022-semantically-invalid\broken\
& "$env:TEMP\server-broken.exe"
# In a second terminal:
clrsig run failures/fm-022-01-semantically-invalid/scenarios/detect.yaml `
  --env failures/fm-022-01-semantically-invalid/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19022).OwningProcess

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-022-semantically-invalid\fixed\
& "$env:TEMP\server-fixed.exe"
# In a second terminal:
clrsig run failures/fm-022-01-semantically-invalid/scenarios/detect.yaml `
  --env failures/fm-022-01-semantically-invalid/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
```

## In the wild

Semantically invalid payloads appear when:

- **Computed fields stored instead of derived** — an order total is written to the database at checkout time; a later price correction or quantity adjustment updates the line items but not the stored total
- **Currency conversion cached** — an amount in a secondary currency is cached at the time of a transaction; exchange rates move but the cached value is never invalidated
- **Aggregates out of sync** — a summary record (e.g., `account.balance`) diverges from the sum of its constituent transactions after a failed partial update
- **Copy-paste data bugs** — a developer copies seed data or test fixtures and accidentally duplicates a value across records with different quantities
- **Rounding applied inconsistently** — `unit_price` is stored with 4 decimal places, `total` is computed with 2, the rounded values no longer multiply cleanly
