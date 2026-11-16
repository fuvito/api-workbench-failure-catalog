# FM-006 — Inconsistent Error Response Schema

> Full catalog entry: https://clrsig.com/failures/fm-006/

**Difficulty:** Easy
**Detected by:** `clrsig run` with `specFile:` binding
**Port:** 19006

## What this failure is

Each endpoint returns a different error shape on 404. The spec declares a single `Error` schema (`{"error": string}`) for every 4xx response across all paths, but the broken server returns:

- `GET /users/999` → `{"message": "user not found"}` — uses `message` instead of `error`
- `GET /orders/999` → `{"code": 404, "msg": "not found"}` — uses numeric `code` + `msg`
- `GET /products/999` → `{"error": "not found"}` — accidentally correct

This pattern appears when:
- Different engineers wrote different handlers without checking the shared error schema
- Error handling was copy-pasted from an older service with a different shape
- The spec was written after the fact and the team missed a few endpoints

The consequence: consumer error-handling code that reads `response.error` works for some endpoints and silently gets `undefined` for others, causing missed error messages or uncaught exceptions.

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-006-error-schema/broken/main.go &
clrsig run failures/fm-006-error-schema/scenarios/detect.yaml \
  --env failures/fm-006-error-schema/environments/local.yaml
```

Expected: **clrsig exits 1** — `/users/999` returns `{"message":...}` and `/orders/999` returns `{"code":...,"msg":...}`; both fail the `Error` schema.

Stop the server before running the fixed version:

```bash
# macOS / Linux
fuser -k 19006/tcp

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19006).OwningProcess
```

## Why clrsig catches it

The scenario sends 404-triggering requests to three endpoints. With `specFile: ../openapi.yaml` declared, clrsig validates every response body — including error responses — against the schema declared for that status code. The broken server's `/users/` and `/orders/` handlers return shapes that don't match `{required: [error], properties: {error: {type: string}}}`.

## Fix

```bash
go run failures/fm-006-error-schema/fixed/main.go &
clrsig run failures/fm-006-error-schema/scenarios/detect.yaml \
  --env failures/fm-006-error-schema/environments/local.yaml
# exits 0
```

**What changed:** `broken/main.go` — all three handlers now call a shared `errNotFound` helper that always returns `{"error": "..."}`.

## GitHub Actions CI setup

This sub-project is included in the CI matrix. If you need to add it back:

```yaml
- { id: "006", slug: "error-schema", kind: "server" }
```

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-006-error-schema\broken\
& "$env:TEMP\server-broken.exe"
# In a second terminal:
clrsig run failures/fm-006-error-schema/scenarios/detect.yaml `
  --env failures/fm-006-error-schema/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19006).OwningProcess

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-006-error-schema\fixed\
& "$env:TEMP\server-fixed.exe"
# In a second terminal:
clrsig run failures/fm-006-error-schema/scenarios/detect.yaml `
  --env failures/fm-006-error-schema/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
```

## In the wild

This failure appears in:

- **Microservice APIs** where error middleware was added to some routers but not others
- **Gradual spec adoption** where the team wrote the OpenAPI spec after the fact and normalized the error shape in the spec without updating all handlers
- **Framework migrations** where the old framework returned `{"message": ...}` and the new one returns `{"error": ...}` but not all handlers were updated
- **Third-party integrations** where one team's SDK reads `response.error` and another reads `response.message`, and both work in their own tests but break each other's
