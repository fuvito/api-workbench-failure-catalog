# fm-006-01 — Old Client Compatibility

> Full catalog entry: https://clrsig.com/failures/fm-006/

**Difficulty:** Easy
**Detected by:** `clrsig run` with the old client's OpenAPI spec as `specFile:`
**Port:** 19079

**Primary failure mode:** [FM-006 — Mobile and Older-Client Compatibility Failures](https://clrsig.com/failures/fm-006/)  

## What this failure is

Mobile app 1.x is still installed on many phones. It handles three order statuses and reads `tracking_code` at the top level. The broken server ships a change for the new app to everyone: a new status, `awaiting_pickup`, and `tracking_code` moved inside `shipment`.

This pattern appears when:
- The backend is released together with the newest app, and older versions are forgotten
- Adding an enum value is treated as a safe change
- A field is moved to a "better" place

The consequence: installed apps crash or show blank screens until users update, and some never do.

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-006-01-old-client-compatibility/broken/main.go &
clrsig run failures/fm-006-01-old-client-compatibility/scenarios/detect.yaml \
  --env failures/fm-006-01-old-client-compatibility/environments/local.yaml
```

Expected: **clrsig exits 1** — the old-app step fails and the new-app step passes.

Stop the server:

```bash
# macOS / Linux
fuser -k 19079/tcp

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19079 -State Listen).OwningProcess
```

## Why clrsig catches it

The new server may match its own new spec perfectly. The key is which contract you test against:

- **The old client's spec.** `openapi.yaml` here is the contract app 1.x was released with. The `specFile:` check fails on the unknown status and the missing `tracking_code`.
- **Old requests.** The step sends exactly what the old app sends, with no version header.
- **New behaviour still ships.** The second step opts into version 2, with spec checking off for that step.

To use it on your own API, save the spec each supported client version was built against, and run one scenario like this per version. Change the request headers to match each client. `clrsig diff` between old and new specs is a useful earlier check, but it only compares documents; this checks the running server.

**What this doesn't cover:** only the behaviour these requests exercise. It can't tell which client versions are still in use, or test offline synchronization.

## Fix

```bash
go run failures/fm-006-01-old-client-compatibility/fixed/main.go &
clrsig run failures/fm-006-01-old-client-compatibility/scenarios/detect.yaml \
  --env failures/fm-006-01-old-client-compatibility/environments/local.yaml
# exits 0
```

**What changed:** `fixed/main.go` — version 2 is returned only to callers that send `Api-Version: 2`. Others get the old shape, with the new status mapped to the closest one old apps know. Responses send `Vary: Api-Version`.

## GitHub Actions CI setup

```yaml
- { dir: "fm-006-01-old-client-compatibility", port: 19079, kind: "server" }
```

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-006-01-old-client-compatibility\broken\
$server = Start-Process "$env:TEMP\server-broken.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-006-01-old-client-compatibility/scenarios/detect.yaml `
  --env failures/fm-006-01-old-client-compatibility/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id $server.Id

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-006-01-old-client-compatibility\fixed\
$server = Start-Process "$env:TEMP\server-fixed.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-006-01-old-client-compatibility/scenarios/detect.yaml `
  --env failures/fm-006-01-old-client-compatibility/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
Stop-Process -Id $server.Id
```

## In the wild

This failure appears in:

- **New enum values** that break exhaustive `switch` statements in mobile apps
- **Restructured responses** released with a new app version
- **Changed defaults** that old clients rely on without sending the parameter
