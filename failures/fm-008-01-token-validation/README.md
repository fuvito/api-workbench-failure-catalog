# fm-008-01 — Token Validation

> Full catalog entry: https://clrsig.com/failures/fm-008/

**Difficulty:** Medium
**Detected by:** `clrsig run` with prepared test tokens and `type: status` assertions
**Port:** 19078

**Primary failure mode:** [FM-008 — OAuth Implementation and Token-Validation Failures](https://clrsig.com/failures/fm-008/)  

## What this failure is

A payments API accepts JWT access tokens. The broken server checks the signature and little else:

| Token | Broken server |
|---|---|
| Expired | Accepted |
| Issued for `reporting-api` | Accepted |
| From another issuer | Accepted |
| `alg: none`, no signature | Accepted |
| Valid, but only `payments:read` | Can create a refund |

This pattern appears when:
- A library verifies the signature, and claim checks are assumed
- The verifier takes the algorithm from the token
- Any valid token is treated as permission for every route

The consequence: old or unrelated tokens, or made-up ones, give access to payments.

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-008-01-token-validation/broken/main.go &
clrsig run failures/fm-008-01-token-validation/scenarios/detect.yaml \
  --env failures/fm-008-01-token-validation/environments/local.yaml
```

Expected: **clrsig exits 1** — the valid-token and bad-signature steps pass and the other five fail.

Stop the server:

```bash
# macOS / Linux
fuser -k 19078/tcp

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19078 -State Listen).OwningProcess
```

## Why clrsig catches it

Each test token is wrong in exactly one way, so a failing step names the missing check. The two control steps show that a valid token works and that signature checking exists at all.

The tokens are synthetic, signed in advance with the demo's fake HS256 secret. The valid one expires in 2100.

**What this doesn't cover:** real OAuth setups usually use asymmetric keys. It doesn't test key rotation (JWKS), revocation, refresh tokens, token type or the OAuth flows themselves.

To use it on your own API, get test tokens from your authorization server or a test signing key, each wrong in one way, and put them in the environment file with `secret: true`. Don't commit tokens that grant real access. Expiring tokens must be reissued, so put the valid one in through `--var` in CI.

## Fix

```bash
go run failures/fm-008-01-token-validation/fixed/main.go &
clrsig run failures/fm-008-01-token-validation/scenarios/detect.yaml \
  --env failures/fm-008-01-token-validation/environments/local.yaml
# exits 0
```

**What changed:** `fixed/main.go` — only `HS256` is accepted, whatever the token claims. After the signature, `exp`, `iss` and `aud` are checked, and each route requires its scope. The demo parser is for illustration; use a maintained JWT library in production.

## GitHub Actions CI setup

```yaml
- { dir: "fm-008-01-token-validation", port: 19078, kind: "server" }
```

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-008-01-token-validation\broken\
$server = Start-Process "$env:TEMP\server-broken.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-008-01-token-validation/scenarios/detect.yaml `
  --env failures/fm-008-01-token-validation/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id $server.Id

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-008-01-token-validation\fixed\
$server = Start-Process "$env:TEMP\server-fixed.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-008-01-token-validation/scenarios/detect.yaml `
  --env failures/fm-008-01-token-validation/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
Stop-Process -Id $server.Id
```

## In the wild

This failure appears in:

- **Shared identity providers** where one service's token works on another
- **Custom JWT code** that honours `alg: none`
- **Scopes defined in the identity provider** but never checked by the API
