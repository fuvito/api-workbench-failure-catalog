# fm-039-01 — Out-of-Order Events

> Full catalog entry: https://clrsig.com/failures/fm-039/

**Difficulty:** Medium
**Detected by:** `clrsig run` that delivers events in the wrong order and reads the stored state
**Port:** 19065

**Primary failure mode:** [FM-039 — Out-of-Order Event and State Transition Failures](https://clrsig.com/failures/fm-039/)  

## What this failure is

The server is an event consumer that tracks order status. The broken version applies whatever arrives last:

- A late `shipped` event moves a delivered order back to `shipped`.
- A delivered order can be cancelled.
- A `cancelled` event for an order that doesn't exist yet creates it.

This pattern appears when:
- A queue or webhook sender doesn't guarantee ordering, and the consumer assumes it does
- State is written with "last write wins"
- Events carry no sequence or version, or it is ignored

The consequence: stored state silently goes backwards or becomes impossible.

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-039-01-out-of-order-events/broken/main.go &
clrsig run failures/fm-039-01-out-of-order-events/scenarios/detect.yaml \
  --env failures/fm-039-01-out-of-order-events/environments/local.yaml
```

Expected: **clrsig exits 1** — the first two steps pass and the other five fail.

Stop the server:

```bash
# macOS / Linux
fuser -k 19065/tcp

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19065 -State Listen).OwningProcess
```

## Why clrsig catches it

Each event is valid on its own, so only the resulting state shows the problem. The scenario sends events in a chosen order and reads the order after them:

- **Stale event.** An older sequence must be acknowledged as `ignored` and change nothing.
- **Invalid transition.** A newer event that isn't allowed must get `409`.
- **Early event.** It must not create an order. `409` asks the sender to deliver it again later.

**What this doesn't cover:** clrsig sends one request at a time, so it doesn't test events processed concurrently. It also doesn't test that a real sender retries after `409`.

To use it on your own consumer, change the event bodies, expected results and the read endpoint. It needs a way to inject events and to read the resulting state, in a test environment. Your policy may buffer early events instead; change those expectations to match.

## Fix

```bash
go run failures/fm-039-01-out-of-order-events/fixed/main.go &
clrsig run failures/fm-039-01-out-of-order-events/scenarios/detect.yaml \
  --env failures/fm-039-01-out-of-order-events/environments/local.yaml
# exits 0
```

**What changed:** `fixed/main.go` — an event with a sequence that isn't newer is ignored. A newer event must be listed in `allowed` for the current status, otherwise it gets `409`. The check and the write happen under one lock.

## GitHub Actions CI setup

```yaml
- { dir: "fm-039-01-out-of-order-events", port: 19065, kind: "server" }
```

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-039-01-out-of-order-events\broken\
$server = Start-Process "$env:TEMP\server-broken.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-039-01-out-of-order-events/scenarios/detect.yaml `
  --env failures/fm-039-01-out-of-order-events/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id $server.Id

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-039-01-out-of-order-events\fixed\
$server = Start-Process "$env:TEMP\server-fixed.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-039-01-out-of-order-events/scenarios/detect.yaml `
  --env failures/fm-039-01-out-of-order-events/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
Stop-Process -Id $server.Id
```

## In the wild

This failure appears in:

- **Webhook consumers** that assume deliveries arrive in order
- **Partitioned queues** where related events land in different partitions
- **Retried messages** that arrive after newer ones
