# fm-050-01 — Configuration Drift

> Full catalog entry: https://clrsig.com/failures/fm-050/

**Difficulty:** Medium
**Detected by:** `clrsig run` that records values from staging and requires production to match
**Port:** 19072 (staging), 19073 (production)

**Primary failure mode:** [FM-050 — Environment Configuration and Feature-Flag Drift](https://clrsig.com/failures/fm-050/)  

## What this failure is

A new tax rule was tested in staging with its flag on. In the broken version, production has three instances: one runs an older configuration, one has the new version with the flag off, and one matches staging. The same quote gives different totals depending on which instance answers.

One process runs both environments. Instances take turns answering, so the demo behaves the same way every run.

This pattern appears when:
- Flags are changed by hand in each environment
- A rollout reaches only some instances
- Instances read configuration at start-up and aren't all restarted

The consequence: behaviour that passed in staging differs in production, sometimes only for some requests.

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-050-01-flag-drift/broken/main.go &
clrsig run failures/fm-050-01-flag-drift/scenarios/detect.yaml \
  --env failures/fm-050-01-flag-drift/environments/local.yaml
```

Expected: **clrsig exits 1** — production quotes 1 and 3 and configuration steps 1 and 3 fail. The one instance that matches staging answers every second step of three, so some production steps pass.

Stop the server:

```bash
# macOS / Linux
fuser -k 19073/tcp

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19073 -State Listen).OwningProcess
```

## Why clrsig catches it

Every response is valid on its own. The failure is a difference:

- **Compare, don't hard-code.** The scenario extracts staging values and uses them as the expected production values.
- **Repeat.** One matching response proves little when several instances share an address, so production is called three times.
- **Configuration and behaviour.** It compares both the reported configuration and the quote.

**What this doesn't cover:** three calls through a load balancer don't guarantee that every instance answered. Percentage rollouts and per-tenant overrides may make differences intentional: compare only what should match.

To use it on your own API, set `stagingUrl` and `prodUrl` and choose requests that are safe to send to production. This scenario only sends `GET` requests. If you expose a configuration endpoint, require authentication and never include secret values.

## Fix

```bash
go run failures/fm-050-01-flag-drift/fixed/main.go &
clrsig run failures/fm-050-01-flag-drift/scenarios/detect.yaml \
  --env failures/fm-050-01-flag-drift/environments/local.yaml
# exits 0
```

**What changed:** `fixed/main.go` — every instance in both environments gets the same versioned release configuration.

## GitHub Actions CI setup

```yaml
- { dir: "fm-050-01-flag-drift", port: 19073, kind: "server" }
```

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-050-01-flag-drift\broken\
$server = Start-Process "$env:TEMP\server-broken.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-050-01-flag-drift/scenarios/detect.yaml `
  --env failures/fm-050-01-flag-drift/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id $server.Id

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-050-01-flag-drift\fixed\
$server = Start-Process "$env:TEMP\server-fixed.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-050-01-flag-drift/scenarios/detect.yaml `
  --env failures/fm-050-01-flag-drift/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
Stop-Process -Id $server.Id
```

## In the wild

This failure appears in:

- **Feature flags** toggled in a dashboard for one environment only
- **Rolling deployments** that stop partway
- **Environment variables** edited on a single server
