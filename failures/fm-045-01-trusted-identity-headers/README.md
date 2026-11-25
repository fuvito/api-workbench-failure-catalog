# fm-045-01 — Trusted Identity Headers

> Full catalog entry: https://clrsig.com/failures/fm-045/

**Difficulty:** Medium
**Detected by:** `clrsig run` with forged headers, sent through the gateway and straight to the backend
**Ports:** 19056 (gateway), 19057 (backend)

**Primary failure mode:** [FM-045 — Authentication and Authorization Confusion Across Services](https://clrsig.com/failures/fm-045/)  

## What this failure is

A gateway checks the caller's token and tells the backend who the caller is in the `X-User-ID` and `X-User-Role` headers. The backend believes those headers, whoever sent them. In the broken version:

| Request | Result |
|---|---|
| No token, forged headers, through the gateway | Admin access, because the gateway forwards requests without a token |
| Member token plus `X-User-Role: admin`, through the gateway | Admin access, because the gateway keeps headers the caller sent |
| Member token plus another `X-User-ID`, through the gateway | The member is treated as that user |
| Forged headers sent straight to the backend | Admin access, with or without a token |

This pattern appears when:
- A backend assumes only the gateway can reach it
- A gateway sets identity headers only when they are missing, or doesn't remove the caller's copies
- Some routes are public at the gateway, so requests without a token are passed through unchanged

The consequence: anyone who can send a request can choose which user and role they are.

One process runs both services. Port 19057 stands for an internal service that a caller can reach, for example from another workload on the same network.

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-045-01-trusted-identity-headers/broken/main.go &
clrsig run failures/fm-045-01-trusted-identity-headers/scenarios/detect.yaml \
  --env failures/fm-045-01-trusted-identity-headers/environments/local.yaml
```

Expected: **clrsig exits 1** — the four control steps pass and the other five fail.

Stop the server (one process serves both ports):

```bash
# macOS / Linux
fuser -k 19056/tcp

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19056 -State Listen).OwningProcess
```

## Why clrsig catches it

Normal use looks correct: the four control steps pass against the broken server, including a member being refused and a request without a token being refused. The `specFile:` binding doesn't detect the failure either, because every response has a declared status and a valid body. The scenario sends what a normal client never would:

- **Forged headers through the gateway.** Without a token, with a member's token, and claiming another user ID.
- **Forged headers straight to the backend.** The environment file holds both URLs.
- **Both layers are tested.** If you fix only the gateway, so that it removes the caller's headers and sets its own, all the gateway steps pass but both backend steps still fail.

**What this doesn't cover:** it only tests the headers it sends, so add the identity headers your own gateway and services use. It also doesn't test token audience or service-to-service credentials, which are part of FM-045 too.

To use it on your own API, set `baseUrl`, `backendUrl`, the tokens and the user ID in `environments/local.yaml`, then change the paths and header names. The scenario only sends `GET` requests. Test only systems you're allowed to test.

If your network blocks direct access to the backend, those two steps fail with a connection error instead of a status. That is a good sign, but it only shows the backend can't be reached from where you ran the test. Run the steps from inside the network, or remove them and test the backend's checks another way.

## Fix

```bash
go run failures/fm-045-01-trusted-identity-headers/fixed/main.go &
clrsig run failures/fm-045-01-trusted-identity-headers/scenarios/detect.yaml \
  --env failures/fm-045-01-trusted-identity-headers/environments/local.yaml
# exits 0
```

**What changed:** `fixed/main.go`:

- The backend verifies the bearer token itself and takes the user and role only from it. It ignores identity headers.
- The gateway rejects every request without a valid token, and removes `X-User-ID` and `X-User-Role` from the rest. Because the fixed backend ignores these headers, the scenario can't tell whether they were removed. Removing them protects other services that might read them.

Verifying the token again is one option. Others are a short-lived identity assertion signed by the gateway, which the backend verifies, or mutual TLS so that the backend accepts connections only from the gateway. Network rules alone leave the backend trusting anyone who gets inside them.

## GitHub Actions CI setup

```yaml
- { dir: "fm-045-01-trusted-identity-headers", port: 19056, kind: "server" }
```

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-045-01-trusted-identity-headers\broken\
$server = Start-Process "$env:TEMP\server-broken.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-045-01-trusted-identity-headers/scenarios/detect.yaml `
  --env failures/fm-045-01-trusted-identity-headers/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id $server.Id

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-045-01-trusted-identity-headers\fixed\
$server = Start-Process "$env:TEMP\server-fixed.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-045-01-trusted-identity-headers/scenarios/detect.yaml `
  --env failures/fm-045-01-trusted-identity-headers/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
Stop-Process -Id $server.Id
```

## In the wild

This failure appears in:

- **Microservices behind an API gateway**, where internal services read identity headers without checking them
- **Reverse proxies that add a user header** after login but don't remove one sent by the caller
- **Internal admin tools** that are protected only by being on a private network
- **New routes added to a gateway as public**, which still reach services that trust identity headers
