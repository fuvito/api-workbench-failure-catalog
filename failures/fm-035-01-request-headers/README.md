# fm-035-01 — Request Header Handling

> Full catalog entry: https://clrsig.com/failures/fm-035/

**Difficulty:** Medium
**Detected by:** `clrsig run` with chosen request headers and `type: status`, `type: header` and `type: jsonpath` assertions
**Port:** 19059

**Primary failure mode:** [FM-035 — Request Header Handling and Validation Failures](https://clrsig.com/failures/fm-035/)  

## What this failure is

Simple requests work, but the server mishandles request headers. In the broken version:

| Header | Mistake |
|---|---|
| `Authorization` | `bearer` in lowercase is rejected, although the scheme name is case-insensitive |
| `X-Correlation-ID` | The ID is lowercased, so it no longer matches the caller's logs |
| `X-Correlation-ID` | Invalid values are accepted and echoed back |
| `If-None-Match` | Only one exact value works; a list or a weak tag (`W/"v1"`) doesn't |
| `If-Match` | Ignored, so an old or unconditional update overwrites newer data |

This pattern appears when:
- Headers are compared as whole strings instead of being parsed
- IDs are "normalized" before they are returned or logged
- A header is declared in the spec but no code reads it, or a proxy removes it

The consequence: valid clients are refused, requests can't be traced, caching doesn't work and updates are lost.

## Prerequisites

- **Go 1.24+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-035-01-request-headers/broken/main.go &
clrsig run failures/fm-035-01-request-headers/scenarios/detect.yaml \
  --env failures/fm-035-01-request-headers/environments/local.yaml
```

Expected: **clrsig exits 1** — 7 steps fail and 3 pass.

Stop the server:

```bash
# macOS / Linux
fuser -k 19059/tcp

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19059 -State Listen).OwningProcess
```

## Why clrsig catches it

Checking responses against the spec can't tell how a request header was read. The scenario sends specific header values and checks the effect:

- **No hard-coded version.** The first step extracts the `ETag`, and later steps send it back.
- **Both sides of a rule.** A valid correlation ID must come back unchanged, and an invalid one must get `400` and not be echoed.
- **State check.** After the refused updates, a read confirms the title wasn't overwritten.

**What this doesn't cover:** clrsig sends one value per header name, so the scenario can't test duplicate headers. It doesn't test oversized headers. The demo has no proxy or gateway: to find headers that are dropped or rewritten on the way, run the scenario against your public gateway URL.

To use it on your own API, set `baseUrl` and `token`, and change the paths, header names, bodies and expected statuses to match your documented policy. The scenario updates the resource, so only point it at disposable test data.

## Fix

```bash
go run failures/fm-035-01-request-headers/fixed/main.go &
clrsig run failures/fm-035-01-request-headers/scenarios/detect.yaml \
  --env failures/fm-035-01-request-headers/environments/local.yaml
# exits 0
```

**What changed:** `fixed/main.go`:

- The auth scheme is compared without regard to case; the token is compared exactly.
- A valid correlation ID is returned unchanged. An invalid one gets `400` with a newly generated ID. Without one, the server generates one.
- `If-None-Match` accepts a list, weak tags and `*`.
- Updates require `If-Match`: `428` without it and `412` if it doesn't match. Weak tags never match. The check and the write happen under one lock.

For concurrent updates in more detail, see [fm-027-01-lost-update](../fm-027-01-lost-update/).

## GitHub Actions CI setup

```yaml
- { dir: "fm-035-01-request-headers", port: 19059, kind: "server" }
```

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-035-01-request-headers\broken\
$server = Start-Process "$env:TEMP\server-broken.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-035-01-request-headers/scenarios/detect.yaml `
  --env failures/fm-035-01-request-headers/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id $server.Id

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-035-01-request-headers\fixed\
$server = Start-Process "$env:TEMP\server-fixed.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-035-01-request-headers/scenarios/detect.yaml `
  --env failures/fm-035-01-request-headers/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
Stop-Process -Id $server.Id
```

## In the wild

This failure appears in:

- **Gateways and proxies** that drop or rewrite `If-Match`, tracing or custom headers
- **Hand-written auth checks** that compare the whole `Authorization` value as one string
- **Caching layers** that never send `304` because they can't parse real `If-None-Match` values
