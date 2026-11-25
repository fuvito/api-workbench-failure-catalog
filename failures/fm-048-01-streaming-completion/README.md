# fm-048-01 — Streaming Completion

> Full catalog entry: https://clrsig.com/failures/fm-048/

**Difficulty:** Medium
**Detected by:** `clrsig run` with an injected source failure and `type: jsonpath` completion and resume assertions
**Port:** 19083

**Primary failure mode:** [FM-048 — Streaming and Large-Response Handling Failures](https://clrsig.com/failures/fm-048/)  

## What this failure is

A records export is streamed. In the broken version, when the data source fails partway, the server closes the JSON neatly, reports `"complete": true` and returns `200`. The client receives 2 of 5 records and believes it has them all. It also ignores `after`, so an export can't be resumed.

The demo streams five records, and an `X-Test-Fault` header makes the source fail after the second.

This pattern appears when:
- Errors that happen after the response has started are logged and ignored
- A stream has no completion marker or record count
- There is no cursor to resume from

The consequence: partial data is treated as complete.

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-048-01-streaming-completion/broken/main.go &
clrsig run failures/fm-048-01-streaming-completion/scenarios/detect.yaml \
  --env failures/fm-048-01-streaming-completion/environments/local.yaml
```

Expected: **clrsig exits 1** — the full-export control step passes and the other two fail.

Stop the server:

```bash
# macOS / Linux
fuser -k 19083/tcp

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19083 -State Listen).OwningProcess
```

## Why clrsig catches it

The truncated export is valid JSON and matches the schema, so only its meaning is wrong:

- **Completion marker.** An interrupted export must say `complete: false` and where to continue.
- **Resume.** Continuing from `next_after` must start at the next record and reach the last one.

**This is partial coverage of FM-048.** It tests completion and resume semantics for a small response that is **one JSON document**. In particular:

- clrsig reads the whole response before checking it. It doesn't test chunk boundaries, slow consumers, cancellation or memory use.
- It doesn't parse NDJSON or server-sent events as streams, and it has no assertion on a raw text body.
- It reads at most 50 MB of a response.
- A connection that is dropped mid-body fails the step with a read error (shown in [fm-002-01-timeout-outcome](../fm-002-01-timeout-outcome/)), but a scenario can't *expect* that error.

To use it on your own API, you need a safe way to interrupt the data source in a test environment. Change the paths, markers and cursor parameter, and never accept a test header in production.

## Fix

```bash
go run failures/fm-048-01-streaming-completion/fixed/main.go &
clrsig run failures/fm-048-01-streaming-completion/scenarios/detect.yaml \
  --env failures/fm-048-01-streaming-completion/environments/local.yaml
# exits 0
```

**What changed:** `fixed/main.go` — `complete` reflects whether every record was sent, `next_after` is the last ID sent, and the export starts after `after`.

## GitHub Actions CI setup

```yaml
- { dir: "fm-048-01-streaming-completion", port: 19083, kind: "server" }
```

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-048-01-streaming-completion\broken\
$server = Start-Process "$env:TEMP\server-broken.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-048-01-streaming-completion/scenarios/detect.yaml `
  --env failures/fm-048-01-streaming-completion/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id $server.Id

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-048-01-streaming-completion\fixed\
$server = Start-Process "$env:TEMP\server-fixed.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-048-01-streaming-completion/scenarios/detect.yaml `
  --env failures/fm-048-01-streaming-completion/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
Stop-Process -Id $server.Id
```

## In the wild

This failure appears in:

- **Large exports** that end early after a database timeout but still look well formed
- **Paged streams** without a completion marker
- **Download jobs** that must restart from the beginning after any failure
