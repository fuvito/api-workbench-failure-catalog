# fm-049-01 — Documentation Drift

> Full catalog entry: https://clrsig.com/failures/fm-049/

**Difficulty:** Easy
**Detected by:** `clrsig run` of the documented examples, copied into scenario steps
**Port:** 19081

**Primary failure mode:** [FM-049 — API Documentation and Example Drift](https://clrsig.com/failures/fm-049/)  

## What this failure is

The documentation shows how to create a ticket and how to filter tickets. In the broken version, the code has moved on: the request field `subject` was renamed to `title`, and the `priority` filter was removed.

This pattern appears when:
- Documentation is written once and edited by hand
- A rename is made in code without updating the examples
- Examples are never executed

The consequence: new integrators copy an example, get an error and can't tell whether they or the documentation are wrong.

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-049-01-documentation-drift/broken/main.go &
clrsig run failures/fm-049-01-documentation-drift/scenarios/detect.yaml \
  --env failures/fm-049-01-documentation-drift/environments/local.yaml
```

Expected: **clrsig exits 1** — both documented examples fail.

Stop the server:

```bash
# macOS / Linux
fuser -k 19081/tcp

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19081 -State Listen).OwningProcess
```

## Why clrsig catches it

- **Examples are executed.** Each step copies a documented request: headers, field names and query parameters.
- **Responses are checked too** against the spec through `specFile:`.

**Copying is manual.** clrsig doesn't run the examples inside an OpenAPI document by itself, so keep these steps in step with your documentation. `clrsig conform` builds requests from spec examples, which can help, but wasn't used or tested for this demo. `clrsig validate --spec` can warn about some examples that break their declared format.

**What this doesn't cover:** examples published only on web pages or in SDKs, unless you copy them here too.

To use it on your own API, add one step per documented example and keep placeholders only for credentials and the base URL. Create examples write data, so use a test account.

## Fix

```bash
go run failures/fm-049-01-documentation-drift/fixed/main.go &
clrsig run failures/fm-049-01-documentation-drift/scenarios/detect.yaml \
  --env failures/fm-049-01-documentation-drift/environments/local.yaml
# exits 0
```

**What changed:** `fixed/main.go` — the server accepts what the documentation shows. Updating the documentation instead is also a fix, but it breaks clients that already followed it.

## GitHub Actions CI setup

```yaml
- { dir: "fm-049-01-documentation-drift", port: 19081, kind: "server" }
```

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-049-01-documentation-drift\broken\
$server = Start-Process "$env:TEMP\server-broken.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-049-01-documentation-drift/scenarios/detect.yaml `
  --env failures/fm-049-01-documentation-drift/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id $server.Id

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-049-01-documentation-drift\fixed\
$server = Start-Process "$env:TEMP\server-fixed.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-049-01-documentation-drift/scenarios/detect.yaml `
  --env failures/fm-049-01-documentation-drift/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
Stop-Process -Id $server.Id
```

## In the wild

This failure appears in:

- **Quick-start guides** that fail on the first request
- **Renamed request fields** that older examples still show
- **Removed filters** that remain in reference pages
