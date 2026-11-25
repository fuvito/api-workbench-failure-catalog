# fm-042-01 — Fault Isolation

> Full catalog entry: https://clrsig.com/failures/fm-042/

**Difficulty:** Medium
**Detected by:** `clrsig run` that breaks a dependency through a test hook, with `type: status` and `type: duration` assertions
**Port:** 19066

**Primary failure mode:** [FM-042 — Dependency Failure and Fault Isolation](https://clrsig.com/failures/fm-042/)  

## What this failure is

Reading an order also fetches recommendations, which the order doesn't need. In the broken version, when the recommendation service is down, the order returns `503`. When it is slow, the order takes as long as it does.

The dependency is simulated inside the demo server. `PUT /test/recommendation-service` switches it between `up`, `down` and `slow`.

This pattern appears when:
- A response is "enriched" by a synchronous call to another service
- Calls to dependencies have no time limit
- Every dependency error is passed on to the caller

The consequence: an outage in a minor feature takes down a core operation.

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-042-01-fault-isolation/broken/main.go &
clrsig run failures/fm-042-01-fault-isolation/scenarios/detect.yaml \
  --env failures/fm-042-01-fault-isolation/environments/local.yaml
```

Expected: **clrsig exits 1** — "Order is still available" and "Order is still fast" fail; the other five steps pass.

Stop the server:

```bash
# macOS / Linux
fuser -k 19066/tcp

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19066 -State Listen).OwningProcess
```

## Why clrsig catches it

With every dependency up, the broken server is correct. The failure only appears when one is broken:

- **Outage.** The order must return `200` and say it is degraded.
- **Slowness.** `type: duration` requires an answer within `maxOrderMs`.
- **Real dependencies still fail.** The recommendations endpoint must return `503`, so hiding every error doesn't pass.

**What this doesn't cover:** it doesn't test circuit breakers, bulkheads, connection pools or behaviour under load.

To use it on your own API, you need a safe way to break one dependency in a test environment, such as a test hook, a stub or a fault-injection tool. Replace the `PUT` steps with it, and set `maxOrderMs` with a wide margin. Never expose a hook like this in production. If a run stops early, switch the dependency back on yourself.

## Fix

```bash
go run failures/fm-042-01-fault-isolation/fixed/main.go &
clrsig run failures/fm-042-01-fault-isolation/scenarios/detect.yaml \
  --env failures/fm-042-01-fault-isolation/environments/local.yaml
# exits 0
```

**What changed:** `fixed/main.go` — the dependency call has a 200 ms budget. If it fails or runs out of time, the order is returned with no recommendations and `degraded: true`.

## GitHub Actions CI setup

```yaml
- { dir: "fm-042-01-fault-isolation", port: 19066, kind: "server" }
```

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-042-01-fault-isolation\broken\
$server = Start-Process "$env:TEMP\server-broken.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-042-01-fault-isolation/scenarios/detect.yaml `
  --env failures/fm-042-01-fault-isolation/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id $server.Id

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-042-01-fault-isolation\fixed\
$server = Start-Process "$env:TEMP\server-fixed.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-042-01-fault-isolation/scenarios/detect.yaml `
  --env failures/fm-042-01-fault-isolation/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
Stop-Process -Id $server.Id
```

## In the wild

This failure appears in:

- **Product and order pages** that depend on review, recommendation or analytics services
- **Missing client timeouts**, so one slow service ties up every caller
- **Shared connection pools** exhausted by a single slow dependency
