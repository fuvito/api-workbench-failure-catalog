# fm-030-01 — API Key Rotation

> Full catalog entry: https://clrsig.com/failures/fm-030/

**Difficulty:** Medium
**Detected by:** `clrsig run` that performs a rotation and checks old and new keys at each stage
**Port:** 19069

**Primary failure mode:** [FM-030 — API Key and Client-Credential Rotation Failures](https://clrsig.com/failures/fm-030/)  

## What this failure is

A client needs time to deploy a new API key. The broken server revokes the old key the moment a new one is created, so every client still using it fails. It also returns all key secrets, old and new, when keys are listed.

This pattern appears when:
- "Rotate" is implemented as "replace"
- Each client can have only one key
- The key record is returned as it is stored

The consequence: rotating a key causes an outage, so teams put it off, and anyone who can list keys can use them.

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-030-01-key-rotation/broken/main.go &
clrsig run failures/fm-030-01-key-rotation/scenarios/detect.yaml \
  --env failures/fm-030-01-key-rotation/environments/local.yaml
```

Expected: **clrsig exits 1** — "Old key still works during the overlap" and the key list step fail. The other six steps pass.

Stop the server:

```bash
# macOS / Linux
fuser -k 19069/tcp

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19069 -State Listen).OwningProcess
```

## Why clrsig catches it

- **The overlap is tested as a stage.** The old key is used again after the new one is created.
- **The new secret is extracted** from the response, so nothing is hard-coded.
- **No secrets in the list.** The spec sets `additionalProperties: false` on listed keys, so a `secret` field fails the `specFile:` check.
- **Revocation still works.** Control steps confirm the old key stops working once revoked.

**What this doesn't cover:** it doesn't test time-based expiry of the overlap or of keys, cached credentials, or logs.

**Secrets in output:** a failing check prints the server's response. When this was tested against the broken server, clrsig printed the listed key's `secret` field as `***`. Don't rely on that for other field names or other versions: use a test account, limit who can read the output, and revoke any key the run creates.

The scenario creates and revokes keys and expects the old key to be listed first. It needs a fresh server each run. To use it on your own API, change the paths, header name and bodies, and only use disposable test keys.

## Fix

```bash
go run failures/fm-030-01-key-rotation/fixed/main.go &
clrsig run failures/fm-030-01-key-rotation/scenarios/detect.yaml \
  --env failures/fm-030-01-key-rotation/environments/local.yaml
# exits 0
```

**What changed:** `fixed/main.go` — creating a key marks active keys `retiring` and leaves them working until they are revoked. The list returns only each key's ID and status.

## GitHub Actions CI setup

```yaml
- { dir: "fm-030-01-key-rotation", port: 19069, kind: "server" }
```

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-030-01-key-rotation\broken\
$server = Start-Process "$env:TEMP\server-broken.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-030-01-key-rotation/scenarios/detect.yaml `
  --env failures/fm-030-01-key-rotation/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id $server.Id

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-030-01-key-rotation\fixed\
$server = Start-Process "$env:TEMP\server-fixed.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-030-01-key-rotation/scenarios/detect.yaml `
  --env failures/fm-030-01-key-rotation/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
Stop-Process -Id $server.Id
```

## In the wild

This failure appears in:

- **"Regenerate key" buttons** that invalidate the old key immediately
- **OAuth client secrets** where only one secret can be active
- **Admin APIs** that return stored secrets in list responses
