# fm-041-01 — Retry Storm

> Full catalog entry: https://clrsig.com/failures/fm-041/

**Difficulty:** Hard
**Detected by:** `clrsig run` that breaks a dependency through a test hook and asserts how many calls reached it
**Port:** 19068

**Primary failure mode:** [FM-041 — Retry Storm and Cascading Failure](https://clrsig.com/failures/fm-041/)  

## What this failure is

A reservation calls an inventory service. When that call fails, the broken server retries at once, up to 10 times, for every request, and even when inventory rejected the request for good.

Inventory is simulated inside the demo server. `PUT /test/inventory` sets its mode and resets a call counter, and `GET /test/inventory` reads the counter.

This pattern appears when:
- Retries are immediate and unlimited, or limited only per request
- Every error is retried, including ones that can't succeed
- Several layers each retry, multiplying the calls

The consequence: a struggling service receives many times its normal traffic and can't recover.

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-041-01-retry-storm/broken/main.go &
clrsig run failures/fm-041-01-retry-storm/scenarios/detect.yaml \
  --env failures/fm-041-01-retry-storm/environments/local.yaml
```

Expected: **clrsig exits 1** — the call-count steps, the fast-failure step and the rejection steps fail.

Stop the server:

```bash
# macOS / Linux
fuser -k 19068/tcp

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19068 -State Listen).OwningProcess
```

## Why clrsig catches it

From the outside both servers answer `503` during an outage. The difference is the traffic they send to the dependency, so the demo exposes a counter and the scenario asserts it:

- **Limit.** One request may cause at most 3 calls.
- **Circuit breaking.** A second request must add no calls and say when to try again.
- **Permanent errors.** A rejection must cause exactly one call.

**What this doesn't cover:** it doesn't check backoff delays or jitter, measure recovery, or send concurrent load. It shows limits for one caller, not how retries multiply across several services.

To use it on your own API, you need a safe way to make a dependency fail and to count the calls it receives, such as a stub or mock server in a test environment. Replace the `/test` steps with it, and set the limits to match your policy.

## Fix

```bash
go run failures/fm-041-01-retry-storm/fixed/main.go &
clrsig run failures/fm-041-01-retry-storm/scenarios/detect.yaml \
  --env failures/fm-041-01-retry-storm/environments/local.yaml
# exits 0
```

**What changed:** `fixed/main.go` — only temporary failures are retried, at most 3 attempts with growing, jittered delays. When a request uses them all, a simple circuit breaker makes later requests fail at once with `Retry-After` for 30 seconds. A rejection is returned after one call. The breaker is deliberately minimal: it has no half-open state.

## GitHub Actions CI setup

```yaml
- { dir: "fm-041-01-retry-storm", port: 19068, kind: "server" }
```

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-041-01-retry-storm\broken\
$server = Start-Process "$env:TEMP\server-broken.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-041-01-retry-storm/scenarios/detect.yaml `
  --env failures/fm-041-01-retry-storm/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id $server.Id

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-041-01-retry-storm\fixed\
$server = Start-Process "$env:TEMP\server-fixed.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-041-01-retry-storm/scenarios/detect.yaml `
  --env failures/fm-041-01-retry-storm/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
Stop-Process -Id $server.Id
```

## In the wild

This failure appears in:

- **HTTP client defaults** that retry without delay
- **Retries at the gateway, service and client** at the same time
- **Queue consumers** that redeliver a permanently failing message forever
