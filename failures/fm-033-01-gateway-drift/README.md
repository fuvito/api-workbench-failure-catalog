# fm-033-01 — Gateway Drift

> Full catalog entry: https://clrsig.com/failures/fm-033/

**Difficulty:** Medium
**Detected by:** `clrsig run` of the same contract checks through two gateways
**Port:** 19074 (staging), 19075 (production)

**Primary failure mode:** [FM-033 — API Gateway Routing and Policy Drift Failures](https://clrsig.com/failures/fm-033/)  

## What this failure is

Staging and production have separate gateways. In the broken version, production has drifted:

| Rule | Staging | Production (broken) |
|---|---|---|
| `/v2/orders` is served by | Version 2 | Version 1 |
| `/v2/reports` needs an API key | Yes | No |
| `X-Request-ID` is passed through | Yes | No |

One process stands for both gateways.

This pattern appears when:
- Gateway routes and policies are edited by hand in each environment
- A change is released to one gateway and not another
- Authentication is configured per route at the gateway

The consequence: tests pass in staging while production serves an old version or an open route.

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-033-01-gateway-drift/broken/main.go &
clrsig run failures/fm-033-01-gateway-drift/scenarios/detect.yaml \
  --env failures/fm-033-01-gateway-drift/environments/local.yaml
```

Expected: **clrsig exits 1** — all four staging control steps pass. In production, only "reports work with a key" passes.

Stop the server:

```bash
# macOS / Linux
fuser -k 19075/tcp

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19075 -State Listen).OwningProcess
```

## Why clrsig catches it

- **One suite, two gateways.** Identical checks run against both URLs, so a difference shows up as a production-only failure.
- **Policy, not just routing.** A request without a key must be refused through each gateway.
- **Shape check.** The `specFile:` binding catches the version 1 order shape.

To use it on your own API, set both URLs and use requests that are safe to send to production: this scenario only sends `GET` requests. If your environments use different keys, give each its own variable. You can also keep one set of steps and run it twice with two environment files.

**What this doesn't cover:** it tests behaviour through the gateways, and doesn't read or compare gateway configuration. It doesn't check timeouts, rate limits or request transformations.

## Fix

```bash
go run failures/fm-033-01-gateway-drift/fixed/main.go &
clrsig run failures/fm-033-01-gateway-drift/scenarios/detect.yaml \
  --env failures/fm-033-01-gateway-drift/environments/local.yaml
# exits 0
```

**What changed:** `fixed/main.go` — both gateways are built from one configuration. With a real gateway, keep that configuration in version control and deploy it to each environment.

## GitHub Actions CI setup

```yaml
- { dir: "fm-033-01-gateway-drift", port: 19075, kind: "server" }
```

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-033-01-gateway-drift\broken\
$server = Start-Process "$env:TEMP\server-broken.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-033-01-gateway-drift/scenarios/detect.yaml `
  --env failures/fm-033-01-gateway-drift/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id $server.Id

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-033-01-gateway-drift\fixed\
$server = Start-Process "$env:TEMP\server-fixed.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-033-01-gateway-drift/scenarios/detect.yaml `
  --env failures/fm-033-01-gateway-drift/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
Stop-Process -Id $server.Id
```

## In the wild

This failure appears in:

- **Canary routes** left pointing at an old service
- **Authentication plugins** enabled in one environment only
- **Header allow-lists** that differ between gateways
