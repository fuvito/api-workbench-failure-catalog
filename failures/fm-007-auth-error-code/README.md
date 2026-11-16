# FM-007 — Authentication Error Code Confusion

> Full catalog entry: https://clrsig.com/failures/fm-007/

**Difficulty:** Easy
**Detected by:** `clrsig run` with explicit `type: status` assertions
**Port:** 19007

## What this failure is

The server returns `200 OK` with an error body for auth failures, instead of `401 Unauthorized`. The spec correctly declares `401` for invalid credentials and missing tokens, but the implementation returns `200` with `{"error": "invalid credentials"}`.

This pattern appears when:
- A framework middleware that "handles" auth errors short-circuits with a 200 response
- The team checks `response.body.error` in client code instead of `response.status`
- An older service was designed around `{"success": false}` envelopes and never migrated to proper HTTP semantics

The consequence: HTTP clients, proxies, and monitoring tools that rely on status codes for routing, caching, and alerting all treat the failed auth as a success. Rate limiters won't trigger on 200. Circuit breakers won't open.

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-007-auth-error-code/broken/main.go &
clrsig run failures/fm-007-auth-error-code/scenarios/detect.yaml \
  --env failures/fm-007-auth-error-code/environments/local.yaml
```

Expected: **clrsig exits 1** — both steps receive `200`; assertions expect `401`.

Stop the server:

```bash
# macOS / Linux
fuser -k 19007/tcp

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19007).OwningProcess
```

## Why clrsig catches it

`specFile:` binding alone won't detect this — a `200` body with `{"error":"..."}` can still match the `Token` schema if the field name happens to overlap. The scenario uses explicit `type: status, operator: eq, expected: 401` assertions to enforce the correct HTTP status code on each step.

## Fix

```bash
go run failures/fm-007-auth-error-code/fixed/main.go &
clrsig run failures/fm-007-auth-error-code/scenarios/detect.yaml \
  --env failures/fm-007-auth-error-code/environments/local.yaml
# exits 0
```

**What changed:** `broken/main.go` — both handlers now call `writeJSON(w, http.StatusUnauthorized, ...)` instead of `http.StatusOK` on auth failures.

## GitHub Actions CI setup

```yaml
- { id: "007", slug: "auth-error-code", kind: "server" }
```

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-007-auth-error-code\broken\
& "$env:TEMP\server-broken.exe"
# In a second terminal:
clrsig run failures/fm-007-auth-error-code/scenarios/detect.yaml `
  --env failures/fm-007-auth-error-code/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19007).OwningProcess

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-007-auth-error-code\fixed\
& "$env:TEMP\server-fixed.exe"
# In a second terminal:
clrsig run failures/fm-007-auth-error-code/scenarios/detect.yaml `
  --env failures/fm-007-auth-error-code/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
```

## In the wild

This failure appears in:

- **Legacy "success envelope" APIs** designed before REST conventions were established, where every response is `200` and the body carries a `success: true/false` flag
- **Misconfigured auth middleware** that catches exceptions and re-emits them as 200 responses to avoid client-side error handling
- **Mobile backends** where the client team controls both sides and builds workarounds for the wrong status code, perpetuating the bug
- **Proxies and gateways** that swallow auth errors and return a custom 200 error page instead of forwarding the 401
