# fm-026-01 — Content Negotiation

> Full catalog entry: https://clrsig.com/failures/fm-026/

**Difficulty:** Medium
**Detected by:** `clrsig run` with chosen `Accept` and `Content-Type` headers and `type: status` and `type: header` assertions
**Port:** 19062

**Primary failure mode:** [FM-026 — HTTP Content-Negotiation Failures](https://clrsig.com/failures/fm-026/)  

## What this failure is

The API offers an order as `application/json` (version 1) and as `application/vnd.example.order-v2+json` (version 2). The broken server mishandles both negotiation headers:

- It never reads `Accept`, so a client asking for version 2, or for XML, silently gets version 1 with `200`.
- It compares the whole request `Content-Type`, so `application/json; charset=utf-8` gets `415`.

This pattern appears when:
- A new representation is documented before negotiation is implemented
- A gateway strips vendor media types
- `Content-Type` is compared as a string instead of being parsed

The consequence: clients parse the wrong shape of data, and valid requests are refused.

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-026-01-content-negotiation/broken/main.go &
clrsig run failures/fm-026-01-content-negotiation/scenarios/detect.yaml \
  --env failures/fm-026-01-content-negotiation/environments/local.yaml
```

Expected: **clrsig exits 1** — the two control steps pass and the other four fail.

Stop the server:

```bash
# macOS / Linux
fuser -k 19062/tcp

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19062 -State Listen).OwningProcess
```

## Why clrsig catches it

The wrong representation is still a valid response: it has status `200` and matches the version 1 schema, so checking against the spec alone passes it. The scenario states what each `Accept` value must produce:

- **Which version.** It checks the response `Content-Type` and a field that exists only in version 2.
- **Quality values.** A request that prefers version 2 must get it.
- **Refusals.** XML must get `406`, and a `text/plain` body must get `415`.
- **Valid variations.** A `charset` parameter must not cause a refusal.

The request steps use `body.type: raw` with an explicit `Content-Type` header, so the scenario controls exactly what is sent.

To use it on your own API, change the paths, media types, bodies and version-specific field. The `POST` steps create records, so only point it at disposable test data. If your API documents a default instead of `406`, change that step.

## Fix

```bash
go run failures/fm-026-01-content-negotiation/fixed/main.go &
clrsig run failures/fm-026-01-content-negotiation/scenarios/detect.yaml \
  --env failures/fm-026-01-content-negotiation/environments/local.yaml
# exits 0
```

**What changed:** `fixed/main.go` — `negotiate` picks from the offered types using quality values and wildcards, returns `406` when none is acceptable and sends `Vary: Accept`. The request `Content-Type` is parsed with `mime.ParseMediaType`, so parameters are allowed. The negotiation covers what this demo needs; it isn't a complete implementation of RFC 9110.

## GitHub Actions CI setup

```yaml
- { dir: "fm-026-01-content-negotiation", port: 19062, kind: "server" }
```

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-026-01-content-negotiation\broken\
$server = Start-Process "$env:TEMP\server-broken.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-026-01-content-negotiation/scenarios/detect.yaml `
  --env failures/fm-026-01-content-negotiation/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id $server.Id

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-026-01-content-negotiation\fixed\
$server = Start-Process "$env:TEMP\server-fixed.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-026-01-content-negotiation/scenarios/detect.yaml `
  --env failures/fm-026-01-content-negotiation/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
Stop-Process -Id $server.Id
```

## In the wild

This failure appears in:

- **Versioned vendor media types** that fall back to the old version without any error
- **Gateways** that rewrite `Accept`
- **Strict string checks** that break clients whose HTTP library adds `charset=utf-8`
