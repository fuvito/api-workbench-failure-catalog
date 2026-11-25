# fm-019-01 — Webhook Processing

> Full catalog entry: https://clrsig.com/failures/fm-019/

**Difficulty:** Medium
**Detected by:** `clrsig run` that delivers signed events and reads the consumer's ledger
**Port:** 19070

**Primary failure mode:** [FM-019 — Webhook Delivery and Event-Processing Failures](https://clrsig.com/failures/fm-019/)  

## What this failure is

The server consumes billing webhooks and records them in a ledger. The broken version:

- Records the same event again when it is delivered twice.
- Accepts a body that was changed after it was signed, because it never checks the signature.
- Answers `200` when processing fails, so the sender never retries and the event is lost.

This pattern appears when:
- The consumer assumes each event arrives exactly once
- Signature checking is left for later, or done on re-serialized JSON
- The handler acknowledges first and processes afterwards, without a durable queue

The consequence: payments are counted twice, forged events take effect and real ones disappear.

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-019-01-webhook-processing/broken/main.go &
clrsig run failures/fm-019-01-webhook-processing/scenarios/detect.yaml \
  --env failures/fm-019-01-webhook-processing/environments/local.yaml
```

Expected: **clrsig exits 1** — 5 steps fail. The first delivery and the retried delivery pass.

Stop the server:

```bash
# macOS / Linux
fuser -k 19070/tcp

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19070 -State Listen).OwningProcess
```

## Why clrsig catches it

A webhook consumer answers `200` to almost everything, so the scenario also checks its effects:

- **Duplicate delivery.** The same signed bytes are sent twice, and the ledger must hold one entry.
- **Tampering.** A changed body with the original signature must get `401`.
- **Failure and retry.** A failed delivery must get `500`, and the retry must take effect.

**Signatures:** clrsig can't compute an HMAC, so the scenario holds bodies signed in advance with the demo's fake secret. Each signature matches exact bytes. For that reason the demo has no signed timestamp, and **replay protection based on timestamps isn't tested**. The scenario also doesn't test delivery delays, retry schedules or dead-letter queues. For ordering, see [fm-039-01-out-of-order-events](../fm-039-01-out-of-order-events/).

To use it on your own consumer, use a test secret, sign your own bodies with your provider's tools, and change the headers and the way effects are read. If your consumer rejects old timestamps, static signatures will expire. You need a safe way to cause a processing failure in a test environment. Restart or reset the consumer before each run.

## Fix

```bash
go run failures/fm-019-01-webhook-processing/fixed/main.go &
clrsig run failures/fm-019-01-webhook-processing/scenarios/detect.yaml \
  --env failures/fm-019-01-webhook-processing/environments/local.yaml
# exits 0
```

**What changed:** `fixed/main.go` — the signature is verified over the raw body, with a constant-time comparison, before the body is used. An `event_id` is remembered only after processing succeeds, and a repeat is acknowledged as `duplicate`. A processing failure returns `500`.

## GitHub Actions CI setup

```yaml
- { dir: "fm-019-01-webhook-processing", port: 19070, kind: "server" }
```

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-019-01-webhook-processing\broken\
$server = Start-Process "$env:TEMP\server-broken.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-019-01-webhook-processing/scenarios/detect.yaml `
  --env failures/fm-019-01-webhook-processing/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id $server.Id

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-019-01-webhook-processing\fixed\
$server = Start-Process "$env:TEMP\server-fixed.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-019-01-webhook-processing/scenarios/detect.yaml `
  --env failures/fm-019-01-webhook-processing/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
Stop-Process -Id $server.Id
```

## In the wild

This failure appears in:

- **Payment provider webhooks** retried after a lost acknowledgement
- **Public webhook URLs** that trust any request in the right shape
- **Handlers that catch all errors** and still return `200`
