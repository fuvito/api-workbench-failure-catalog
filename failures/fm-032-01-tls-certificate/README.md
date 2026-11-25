# fm-032-01 — TLS Certificates

> Full catalog entry: https://clrsig.com/failures/fm-032/

**Difficulty:** Medium
**Detected by:** `clrsig run` over HTTPS, trusting a test CA with `--ca-cert`
**Port:** 19076 (health, HTTP), 19077, 19085 and 19086 (TLS)

**Primary failure mode:** [FM-032 — TLS/mTLS and Certificate Lifecycle Failures](https://clrsig.com/failures/fm-032/)  

## What this failure is

Three TLS endpoints each have a certificate mistake in the broken version:

| Endpoint | Mistake | Error from clrsig |
|---|---|---|
| API, 19077 | The intermediate certificate isn't sent | `certificate signed by unknown authority` |
| Admin, 19085 | The certificate has expired | `certificate has expired or is not yet valid` |
| Partner, 19086 | The certificate is for another host name | `certificate is valid for old-partner.example.test, not localhost` |

When it starts, the server creates a new test CA and certificates. It writes only the CA's **public** certificate to `.certs/ca.pem` (ignored by git). Private keys are kept in memory and never saved, so no keys are in the repository.

This pattern appears when:
- A renewed certificate is installed without its intermediate
- Renewal is manual and nobody is watching the expiry date
- A certificate is copied to a host it wasn't issued for

The consequence: connections fail before any HTTP request is made, often for only some clients or endpoints.

**Run from the repo root:** the server writes the CA to a path relative to where it starts. Start the server before running clrsig.

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-032-01-tls-certificate/broken/main.go &
clrsig run failures/fm-032-01-tls-certificate/scenarios/detect.yaml \
  --env failures/fm-032-01-tls-certificate/environments/local.yaml \
  --ca-cert failures/fm-032-01-tls-certificate/.certs/ca.pem
```

Expected: **clrsig exits 1** — the plain HTTP control step passes and the three TLS steps fail.

Stop the server:

```bash
# macOS / Linux
fuser -k 19076/tcp

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19076 -State Listen).OwningProcess
```

## Why clrsig catches it

clrsig verifies certificates like any strict client, so a TLS failure fails the step before there is a response. Each endpoint is a separate step, so the report shows which one failed and why.

- **Only the root is trusted.** `--ca-cert` (or `CLRSIG_CA_CERT`) adds the test root to the system's trusted roots. The server must send the intermediate itself.
- **Don't skip verification.** Turning verification off would hide all three failures.

**What this doesn't cover:** certificates that will expire soon (only already expired ones fail), revocation, TLS versions and cipher policy, and mutual TLS. clrsig accepts a client certificate with `--client-cert` and `--client-key`, but this demo doesn't use it. A scenario can't *expect* a handshake to fail, so it can't assert that a client without a certificate is refused. If a host has several servers behind one address, one run may not reach all of them.

To use it on your own endpoints, set the URLs in `environments/local.yaml`. For publicly trusted certificates, leave out `--ca-cert`. For a private CA, pass its root certificate.

## Fix

```bash
go run failures/fm-032-01-tls-certificate/fixed/main.go &
clrsig run failures/fm-032-01-tls-certificate/scenarios/detect.yaml \
  --env failures/fm-032-01-tls-certificate/environments/local.yaml \
  --ca-cert failures/fm-032-01-tls-certificate/.certs/ca.pem
# exits 0
```

**What changed:** `fixed/main.go` — every TLS endpoint sends an in-date certificate for `localhost`, followed by the intermediate.

**CI:** the workflows load the optional file `ci.env` from an example's folder. This one sets `CLRSIG_CA_CERT`.

## GitHub Actions CI setup

```yaml
- { dir: "fm-032-01-tls-certificate", port: 19076, kind: "server" }
```

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-032-01-tls-certificate\broken\
$server = Start-Process "$env:TEMP\server-broken.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-032-01-tls-certificate/scenarios/detect.yaml `
  --env failures/fm-032-01-tls-certificate/environments/local.yaml `
  --ca-cert failures/fm-032-01-tls-certificate/.certs/ca.pem
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id $server.Id

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-032-01-tls-certificate\fixed\
$server = Start-Process "$env:TEMP\server-fixed.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-032-01-tls-certificate/scenarios/detect.yaml `
  --env failures/fm-032-01-tls-certificate/environments/local.yaml `
  --ca-cert failures/fm-032-01-tls-certificate/.certs/ca.pem
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
Stop-Process -Id $server.Id
```

## In the wild

This failure appears in:

- **Load balancers** where one node still has the old certificate or chain
- **Internal services** whose certificates expire because no one renews them
- **Host renames** where the old certificate stays in place
