# fm-013-01 — Undocumented Dependency

> Full catalog entry: https://clrsig.com/failures/fm-013/

**Difficulty:** Easy
**Detected by:** `clrsig run` of a consumer's expectations against the provider
**Port:** 19084

**Primary failure mode:** [FM-013 — Undocumented Dependencies and Accumulated Integration Failures](https://clrsig.com/failures/fm-013/)  

## What this failure is

A finance export job reads `legacyAccountCode` from accounts and skips closed accounts by the error code `ACCOUNT_CLOSED`. Neither is in the provider's spec. In the broken version, a cleanup removed both. The provider's own contract checks still pass.

This pattern appears when:
- Consumers depend on whatever the response happens to contain
- Fields that no known consumer uses are removed as dead code
- Error codes are treated as internal details

The consequence: an integration the provider didn't know about fails, perhaps weeks later at month end.

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-013-01-undocumented-dependency/broken/main.go &
clrsig run failures/fm-013-01-undocumented-dependency/scenarios/detect.yaml \
  --env failures/fm-013-01-undocumented-dependency/environments/local.yaml
```

Expected: **clrsig exits 1** — the spec control step passes, and both finance export steps fail.

Stop the server:

```bash
# macOS / Linux
fuser -k 19084/tcp

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19084 -State Listen).OwningProcess
```

## Why clrsig catches it

Nothing can find a dependency that nobody has written down. This catches a dependency **once a consumer's expectations are recorded**:

- **The provider's check is not enough.** The first step passes, because the removed values were never in the spec.
- **Consumer-owned steps.** Each step states what one consumer relies on, and names its owner. The provider runs them before releasing.

**What this doesn't cover:** consumers that haven't written a scenario remain invisible. A passing run doesn't mean no one depends on what you removed. Finding unknown consumers needs other evidence, such as traffic analysis or access logs. `clrsig impact` can show which scenarios touch an operation with a breaking spec change, but it can't help with fields the spec never listed, and it isn't used here.

To use it on your own API, ask each consumer team for the fields, codes and behaviour they rely on, add one step per expectation with an owner in `notes`, and run the scenario in the provider's CI.

## Fix

```bash
go run failures/fm-013-01-undocumented-dependency/fixed/main.go &
clrsig run failures/fm-013-01-undocumented-dependency/scenarios/detect.yaml \
  --env failures/fm-013-01-undocumented-dependency/environments/local.yaml
# exits 0
```

**What changed:** `fixed/main.go` — both values stay until the finance job has migrated. The longer-term fix is to document them, or agree a deprecation date: see [fm-015-01-deprecation-lifecycle](../fm-015-01-deprecation-lifecycle/).

## GitHub Actions CI setup

```yaml
- { dir: "fm-013-01-undocumented-dependency", port: 19084, kind: "server" }
```

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-013-01-undocumented-dependency\broken\
$server = Start-Process "$env:TEMP\server-broken.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-013-01-undocumented-dependency/scenarios/detect.yaml `
  --env failures/fm-013-01-undocumented-dependency/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id $server.Id

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-013-01-undocumented-dependency\fixed\
$server = Start-Process "$env:TEMP\server-fixed.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-013-01-undocumented-dependency/scenarios/detect.yaml `
  --env failures/fm-013-01-undocumented-dependency/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
Stop-Process -Id $server.Id
```

## In the wild

This failure appears in:

- **Batch jobs and spreadsheets** that read fields the API team has forgotten
- **Client code that branches on error text or codes** that were never promised
- **Internal integrations** with no owner on the provider's consumer list
