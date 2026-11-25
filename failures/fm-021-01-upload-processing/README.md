# fm-021-01 — Upload Processing

> Full catalog entry: https://clrsig.com/failures/fm-021/

**Difficulty:** Medium
**Detected by:** `clrsig run` with multipart uploads, an injected storage failure and record counts
**Port:** 19071

**Primary failure mode:** [FM-021 — File Upload and Downstream Processing Failures](https://clrsig.com/failures/fm-021/)  

## What this failure is

The API says `201` when a file is uploaded. The broken server saves the file record first and ignores a storage failure, so the client is told the file is stored when it isn't. It also accepts empty files and never checks the client's checksum.

Object storage is simulated in memory, and an `X-Test-Fault: storage-write` header makes the write fail.

This pattern appears when:
- Metadata and file content are written to different systems
- Errors from the second write are logged and swallowed
- Success is returned before processing finishes

The consequence: records point to files that don't exist, and nobody finds out until someone downloads one.

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-021-01-upload-processing/broken/main.go &
clrsig run failures/fm-021-01-upload-processing/scenarios/detect.yaml \
  --env failures/fm-021-01-upload-processing/environments/local.yaml
```

Expected: **clrsig exits 1** — the first three steps pass and the other four fail.

Stop the server:

```bash
# macOS / Linux
fuser -k 19071/tcp

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19071 -State Listen).OwningProcess
```

## Why clrsig catches it

- **Success must be true.** A control upload is downloaded again to show that `201` means the file is there.
- **Injected failure.** A failed storage write must return `502`.
- **Validation.** An empty file and a wrong checksum must be refused.
- **No leftovers.** The record count must be unchanged after the three failed uploads.

The download step checks the status only: clrsig doesn't compare binary content.

**Fixture paths:** clrsig reads multipart files from disk. Its format documentation asks for absolute paths. This demo uses a path relative to the folder you run clrsig from, which worked when tested from the repo root; run it from there, or set `fixturesDir` to an absolute path with `--var fixturesDir=...`.

**What this doesn't cover:** large or interrupted transfers, duplicate uploads, virus scanning and other asynchronous processing.

To use it on your own API, replace the paths, field name and fixtures, and update the checksum. You need a safe way to make storage fail in a test environment. The scenario creates files, so only point it at disposable test data.

## Fix

```bash
go run failures/fm-021-01-upload-processing/fixed/main.go &
clrsig run failures/fm-021-01-upload-processing/scenarios/detect.yaml \
  --env failures/fm-021-01-upload-processing/environments/local.yaml
# exits 0
```

**What changed:** `fixed/main.go` — empty files and checksum mismatches are refused first. The file is stored before it is recorded, and a storage failure returns `502` with no record. With real systems the two writes can't share a transaction, so you also need cleanup for files stored without a record.

## GitHub Actions CI setup

```yaml
- { dir: "fm-021-01-upload-processing", port: 19071, kind: "server" }
```

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-021-01-upload-processing\broken\
$server = Start-Process "$env:TEMP\server-broken.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-021-01-upload-processing/scenarios/detect.yaml `
  --env failures/fm-021-01-upload-processing/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id $server.Id

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-021-01-upload-processing\fixed\
$server = Start-Process "$env:TEMP\server-fixed.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-021-01-upload-processing/scenarios/detect.yaml `
  --env failures/fm-021-01-upload-processing/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
Stop-Process -Id $server.Id
```

## In the wild

This failure appears in:

- **Avatar and attachment uploads** where the database row is saved before the object-store write
- **Import pipelines** that return success when the file is only queued
- **Zero-byte files** from interrupted clients, accepted and processed later as valid
