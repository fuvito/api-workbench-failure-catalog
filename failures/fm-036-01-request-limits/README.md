# fm-036-01 — Request Limits

> Full catalog entry: https://clrsig.com/failures/fm-036/

**Difficulty:** Easy
**Detected by:** `clrsig run` with requests at and just above each limit and `type: status` assertions
**Port:** 19063

**Primary failure mode:** [FM-036 — Request Size and Resource Exhaustion Failures](https://clrsig.com/failures/fm-036/)  

## What this failure is

The API documents three limits: a 1024-byte body, 10 tags and metadata nested 5 levels. The broken server enforces none of them properly:

- A body over the limit hits an internal buffer and returns `500`, as if the server had failed.
- 11 tags and metadata 6 levels deep are accepted with `201`.

This pattern appears when:
- A gateway and a backend have different limits
- Limits are written in the spec but not checked in code
- The body is read fully before its size is checked

The consequence: clients can't tell that their request was too large, and unbounded input can exhaust memory or processing time.

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-036-01-request-limits/broken/main.go &
clrsig run failures/fm-036-01-request-limits/scenarios/detect.yaml \
  --env failures/fm-036-01-request-limits/environments/local.yaml
```

Expected: **clrsig exits 1** — the three control steps pass and the other three fail.

Stop the server:

```bash
# macOS / Linux
fuser -k 19063/tcp

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19063 -State Listen).OwningProcess
```

## Why clrsig catches it

- **Both sides of each limit.** A control step at or below each limit must be accepted, so a server that rejects everything fails.
- **Documented status.** An oversized body must get `413`, not `500`.
- **Generated payloads.** `{{$randomString}}` creates the large values, so the scenario file stays readable.

The `specFile:` check alone doesn't find these failures here: it checks responses, and the scenario chooses what to send. The response schema's `maximum` on `tag_count` does flag the 11-tag response.

**What this doesn't cover:** the payloads are tiny and sent once. The scenario doesn't test large uploads, large headers, memory or CPU use, or load. Don't use it to stress a production system.

To use it on your own API, set the sizes just below and above your real limits and change the bodies and expected statuses. The accepted requests create records, so only point it at disposable test data.

## Fix

```bash
go run failures/fm-036-01-request-limits/fixed/main.go &
clrsig run failures/fm-036-01-request-limits/scenarios/detect.yaml \
  --env failures/fm-036-01-request-limits/environments/local.yaml
# exits 0
```

**What changed:** `fixed/main.go` — `http.MaxBytesReader` stops reading after 1024 bytes and returns `413`. Tag count and metadata depth are checked and return `400`.

The demo checks depth after parsing, which is safe only because the size limit comes first.

## GitHub Actions CI setup

```yaml
- { dir: "fm-036-01-request-limits", port: 19063, kind: "server" }
```

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-036-01-request-limits\broken\
$server = Start-Process "$env:TEMP\server-broken.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-036-01-request-limits/scenarios/detect.yaml `
  --env failures/fm-036-01-request-limits/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id $server.Id

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-036-01-request-limits\fixed\
$server = Start-Process "$env:TEMP\server-fixed.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-036-01-request-limits/scenarios/detect.yaml `
  --env failures/fm-036-01-request-limits/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
Stop-Process -Id $server.Id
```

## In the wild

This failure appears in:

- **Reverse proxies** that allow larger bodies than the application can parse
- **Array fields** with `maxItems` in the spec and no check in the handler
- **Free-form JSON fields** that accept any depth
