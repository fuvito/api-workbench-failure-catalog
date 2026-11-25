# fm-040-01 — Partial Write

> Full catalog entry: https://clrsig.com/failures/fm-040/

**Difficulty:** Medium
**Detected by:** `clrsig run` that causes a mid-write failure and reads back what was stored
**Port:** 19067

**Primary failure mode:** [FM-040 — Database Transaction and Atomicity Failures](https://clrsig.com/failures/fm-040/)  

## What this failure is

Creating an invoice writes a header and then its lines. In the broken version, if a later write fails, the request returns an error but the header and earlier lines stay. A retry with the same number then gets `409`.

The demo keeps data in memory and simulates the failures: an `X-Test-Fault` header fails the second line write, and two lines with the same `sku` break a unique constraint.

This pattern appears when:
- Related writes are not in one transaction
- Each repository call commits on its own
- An error is returned without rolling back

The consequence: incomplete records are treated as real ones, and clients can't retry.

[fm-020-01-batch-atomicity](../fm-020-01-batch-atomicity/) covers a batch of independent items with one invalid input. This example is about one record whose related writes fail partway.

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-040-01-partial-write/broken/main.go &
clrsig run failures/fm-040-01-partial-write/scenarios/detect.yaml \
  --env failures/fm-040-01-partial-write/environments/local.yaml
```

Expected: **clrsig exits 1** — the three reads and retries after a failure fail; the other three steps pass.

Stop the server:

```bash
# macOS / Linux
fuser -k 19067/tcp

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19067 -State Listen).OwningProcess
```

## Why clrsig catches it

The error status is correct in both versions, so a test that only checks the error passes. The scenario then looks at what was stored:

- **Read after failure.** The invoice must not exist.
- **Retry.** The same number must be accepted once the fault is gone.
- **Two causes.** An infrastructure failure and a constraint violation, which often take different code paths.

**What this doesn't cover:** it doesn't test a real database, isolation levels, deadlocks or concurrent transactions. It only sees what the read endpoint returns, not orphaned rows in other tables.

To use it on your own API, you need a safe way to cause a mid-write failure in a test environment. Replace the `X-Test-Fault` header with yours, and never accept such a header in production.

## Fix

```bash
go run failures/fm-040-01-partial-write/fixed/main.go &
clrsig run failures/fm-040-01-partial-write/scenarios/detect.yaml \
  --env failures/fm-040-01-partial-write/environments/local.yaml
# exits 0
```

**What changed:** `fixed/main.go` — `stage` builds the whole invoice without storing it. It is stored with one write only if every line succeeded. With a real database, use one transaction and roll it back on any error.

## GitHub Actions CI setup

```yaml
- { dir: "fm-040-01-partial-write", port: 19067, kind: "server" }
```

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-040-01-partial-write\broken\
$server = Start-Process "$env:TEMP\server-broken.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-040-01-partial-write/scenarios/detect.yaml `
  --env failures/fm-040-01-partial-write/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id $server.Id

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-040-01-partial-write\fixed\
$server = Start-Process "$env:TEMP\server-fixed.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-040-01-partial-write/scenarios/detect.yaml `
  --env failures/fm-040-01-partial-write/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
Stop-Process -Id $server.Id
```

## In the wild

This failure appears in:

- **Order and invoice creation** that commits the header before the lines
- **ORM code** that saves child records in separate transactions
- **Error handlers** that return early and skip the rollback
