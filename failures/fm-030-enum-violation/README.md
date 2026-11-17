# FM-030 — Enum Value Violation

> Full catalog entry: https://clrsig.com/failures/fm-030/

**Difficulty:** Easy
**Detected by:** `clrsig run` with `specFile:` binding
**Port:** 19030

## What this failure is

The server returns field values that are not in the declared OpenAPI enum. Three variants are demonstrated:

- **Wrong case** — `"PENDING"` instead of `"pending"` (enum is case-sensitive)
- **Undeclared value** — `"SHIPPED"` is not in `[pending, processing, completed, cancelled]` at all
- **Mixed** — status is correct but `priority` is `"HIGH"` instead of `"high"`

This pattern appears when:
- The internal domain model uses uppercase constants (`ORDER_STATUS_PENDING`) and the serializer passes them through without mapping
- A new status is added to the backend (`shipped`) before the spec is updated
- Two teams own the spec and the implementation separately and drift apart
- An enum was normalized to lowercase in the spec but the database column still stores uppercase

The consequence: consumer code doing `if (order.status === "pending")` never matches `"PENDING"`, and switch statements fall through to default cases silently.

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-030-enum-violation/broken/main.go &
clrsig run failures/fm-030-enum-violation/scenarios/detect.yaml \
  --env failures/fm-030-enum-violation/environments/local.yaml
```

Expected: **clrsig exits 1** — orders 1 and 3 have wrong-case values; order 2 has an undeclared `"SHIPPED"` status.

Stop the server:

```bash
# macOS / Linux
fuser -k 19030/tcp

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19030).OwningProcess
```

## Why clrsig catches it

With `specFile: ../openapi.yaml`, clrsig validates enum constraints on every response. JSON Schema enum validation is exact — `"PENDING"` does not match `"pending"`:

```
✗ schema: Error at "/status": value must be one of: pending, processing, completed, cancelled
```

No explicit assertions needed — `assertions: []` is sufficient.

## Fix

```bash
go run failures/fm-030-enum-violation/fixed/main.go &
clrsig run failures/fm-030-enum-violation/scenarios/detect.yaml \
  --env failures/fm-030-enum-violation/environments/local.yaml
# exits 0
```

**What changed:** `broken/main.go` — all order statuses and priorities normalized to lowercase declared enum values; `"SHIPPED"` mapped to `"processing"`.

## GitHub Actions CI setup

```yaml
- { id: "030", slug: "enum-violation", kind: "server" }
```

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-030-enum-violation\broken\
& "$env:TEMP\server-broken.exe"
# In a second terminal:
clrsig run failures/fm-030-enum-violation/scenarios/detect.yaml `
  --env failures/fm-030-enum-violation/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19030).OwningProcess

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-030-enum-violation\fixed\
& "$env:TEMP\server-fixed.exe"
# In a second terminal:
clrsig run failures/fm-030-enum-violation/scenarios/detect.yaml `
  --env failures/fm-030-enum-violation/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
```

## In the wild

This failure appears in:

- **Domain model leakage** — Go `iota` constants or Java enums with `.name()` serialization emit uppercase identifiers directly onto the wire
- **Database migrations** — a column stored `ACTIVE`/`INACTIVE` and a spec rewrite normalized it to `active`/`inactive` without updating the query layer
- **New status codes** — a `shipped` status added for a fulfilment integration was never added to the OpenAPI spec, so consumers don't know it can appear
- **Locale-sensitive serializers** — `strings.ToUpper` or `.toLowerCase()` applied inconsistently across different handlers
