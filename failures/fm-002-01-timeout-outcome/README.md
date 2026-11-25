# fm-002-01 — Timeouts and Incomplete Responses

> Full catalog entry: https://clrsig.com/failures/fm-002/

**Difficulty:** Medium
**Detected by:** `clrsig run` with `type: duration` and status assertions, an injected delay and a cut-off response
**Port:** 19082

**Primary failure mode:** [FM-002 — Silent Timeout and Incomplete Response Failures](https://clrsig.com/failures/fm-002/)  

## What this failure is

An order is saved, then the work that follows is slow. The broken server:

- Keeps the caller waiting as long as it takes, with no limit.
- Creates a second order when the client retries with the same `Idempotency-Key`.
- When an export fails partway, has already sent `200` and the full `Content-Length`, so the body is cut off.

The delay and the export failure are triggered by an `X-Test-Fault` header.

This pattern appears when:
- Handlers have no time budget
- A client that timed out retries, and the server can't tell it is a retry
- A response is streamed before the server knows it can finish it

The consequence: clients don't know whether an operation happened, retries duplicate it, and truncated data arrives under a success status.

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-002-01-timeout-outcome/broken/main.go &
clrsig run failures/fm-002-01-timeout-outcome/scenarios/detect.yaml \
  --env failures/fm-002-01-timeout-outcome/environments/local.yaml
```

Expected: **clrsig exits 1** — the export control step passes and the other four fail.

Stop the server:

```bash
# macOS / Linux
fuser -k 19082/tcp

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19082 -State Listen).OwningProcess
```

## Why clrsig catches it

- **Bounded time.** `type: duration` requires an answer within `maxResponseMs`.
- **Unknown outcome.** The timeout response must say the outcome is unknown.
- **Safe retry.** The retry must return the existing order, and the count must be `1`.
- **Cut-off response.** clrsig reports it as an error, `reading response body: unexpected EOF`, which fails the step.

**What this doesn't cover:** it doesn't test client or gateway timeouts, or a delay before the order is saved. clrsig has no way to assert that a request *should* time out or be cut off: those always fail the step. For idempotency in more detail, see [fm-018-01-idempotency](../fm-018-01-idempotency/).

To use it on your own API, you need a safe way to slow a request and fail a response in a test environment, such as a fault-injection proxy. Replace the `X-Test-Fault` header, use your own timeout response and limits, and never accept a test header in production. The scenario creates orders, so only use disposable test data.

## Fix

```bash
go run failures/fm-002-01-timeout-outcome/fixed/main.go &
clrsig run failures/fm-002-01-timeout-outcome/scenarios/detect.yaml \
  --env failures/fm-002-01-timeout-outcome/environments/local.yaml
# exits 0
```

**What changed:** `fixed/main.go` — after 500 ms the server answers `504` with `outcome: unknown`. A retry with the same key returns the stored order with `200`. The export is built completely before anything is sent, so a failure can still be reported as `500`. For exports too large to build first, see [fm-048-01-streaming-completion](../fm-048-01-streaming-completion/).

## GitHub Actions CI setup

```yaml
- { dir: "fm-002-01-timeout-outcome", port: 19082, kind: "server" }
```

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-002-01-timeout-outcome\broken\
$server = Start-Process "$env:TEMP\server-broken.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-002-01-timeout-outcome/scenarios/detect.yaml `
  --env failures/fm-002-01-timeout-outcome/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id $server.Id

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-002-01-timeout-outcome\fixed\
$server = Start-Process "$env:TEMP\server-fixed.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-002-01-timeout-outcome/scenarios/detect.yaml `
  --env failures/fm-002-01-timeout-outcome/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
Stop-Process -Id $server.Id
```

## In the wild

This failure appears in:

- **Payment and order APIs** where a client timeout leads to a second charge
- **Gateways** that return a timeout while the backend carries on and commits
- **Report downloads** that fail partway after `200` has been sent
