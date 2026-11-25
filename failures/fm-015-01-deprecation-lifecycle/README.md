# fm-015-01 — Deprecation Lifecycle

> Full catalog entry: https://clrsig.com/failures/fm-015/

**Difficulty:** Easy
**Detected by:** `clrsig run` with `type: status` and `type: header` assertions, plus required headers in the spec
**Port:** 19060

**Primary failure mode:** [FM-015 — Deprecated Endpoint and Lifecycle Failures](https://clrsig.com/failures/fm-015/)  

## What this failure is

The provider published a lifecycle: version 1 is deprecated but stays available until 31 December 2027, and version 0 is retired. The broken server doesn't follow it:

| Endpoint | Mistake |
|---|---|
| `/v1/invoices` | Returns `404` although its sunset date hasn't arrived |
| `/v1/customers` | Works, but sends no `Deprecation`, `Sunset` or replacement `Link` |
| `/v0/invoices` | Returns `404`, the same as a mistyped URL, instead of `410` naming the replacement |

This pattern appears when:
- A cleanup removes "old" routes without checking the announced date
- Deprecation is announced in documentation only
- Retiring a route means deleting its handler

The consequence: consumers either break without warning or never learn that they have to move.

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-015-01-deprecation-lifecycle/broken/main.go &
clrsig run failures/fm-015-01-deprecation-lifecycle/scenarios/detect.yaml \
  --env failures/fm-015-01-deprecation-lifecycle/environments/local.yaml
```

Expected: **clrsig exits 1** — the control step passes and the other three fail.

Stop the server:

```bash
# macOS / Linux
fuser -k 19060/tcp

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19060 -State Listen).OwningProcess
```

## Why clrsig catches it

- **Policy as assertions.** The expected `Deprecation` and `Sunset` values come from the environment file, so a changed or missing date fails.
- **Early removal.** The scenario keeps calling the deprecated endpoint and expects `200`.
- **Required headers.** The spec marks the lifecycle headers as required, so the `specFile:` check also reports a missing one.

The `Deprecation` header follows RFC 9745 and `Sunset` follows RFC 8594. Not every API uses these headers: test your own published policy.

**What this doesn't cover:** clrsig doesn't compare the `Sunset` date with today's date, so the scenario won't tell you when it has passed. After that date, change the step to expect your retirement response. It also can't tell whether consumers still call a deprecated endpoint.

To use it on your own API, set `baseUrl`, `deprecation` and `sunset`, and change the paths and the replacement links. The scenario only sends `GET` requests. To catch an endpoint removed from the spec itself, compare spec versions with `clrsig diff`.

## Fix

```bash
go run failures/fm-015-01-deprecation-lifecycle/fixed/main.go &
clrsig run failures/fm-015-01-deprecation-lifecycle/scenarios/detect.yaml \
  --env failures/fm-015-01-deprecation-lifecycle/environments/local.yaml
# exits 0
```

**What changed:** `fixed/main.go` — `deprecated` wraps each version 1 route and adds the three headers while still serving it. `retired` returns `410` with the replacement in `Link`.

## GitHub Actions CI setup

```yaml
- { dir: "fm-015-01-deprecation-lifecycle", port: 19060, kind: "server" }
```

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-015-01-deprecation-lifecycle\broken\
$server = Start-Process "$env:TEMP\server-broken.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-015-01-deprecation-lifecycle/scenarios/detect.yaml `
  --env failures/fm-015-01-deprecation-lifecycle/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id $server.Id

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-015-01-deprecation-lifecycle\fixed\
$server = Start-Process "$env:TEMP\server-fixed.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-015-01-deprecation-lifecycle/scenarios/detect.yaml `
  --env failures/fm-015-01-deprecation-lifecycle/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
Stop-Process -Id $server.Id
```

## In the wild

This failure appears in:

- **Route cleanups** that remove an endpoint before the date given to partners
- **Deprecations announced only in a changelog**, which automated clients never see
- **Retired versions that return `404`**, so teams look for a typo instead of migrating
