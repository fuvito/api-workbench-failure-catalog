# fm-010-01 — OpenAPI Specification and Implementation Drift

> Full catalog entry: https://clrsig.com/failures/fm-010/

**Difficulty:** Easy
**Detected by:** `clrsig run` + `specFile:` schema assertions
**Port:** 19010

**Primary failure mode:** [FM-010 — OpenAPI Specification and Implementation Drift](https://clrsig.com/failures/fm-010/)  

## What this failure is

The OpenAPI spec and the server implementation have drifted. During a domain-model refactor, a developer renamed two fields — `category` → `product_type` and `in_stock` → `available` — without updating the spec. The server still responds successfully with valid JSON, so the drift goes undetected in normal integration tests.

The difference from FM-001 and FM-005: this scenario uses `specFile:` binding in `detect.yaml`. clrsig loads the OpenAPI spec and validates the entire response body against the declared schema automatically — no field-by-field assertions needed. Any drift between spec and server surfaces without knowing in advance what drifted.

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-010-01-spec-drift/broken/main.go &
clrsig run failures/fm-010-01-spec-drift/scenarios/detect.yaml \
  --env failures/fm-010-01-spec-drift/environments/local.yaml
```

Expected: **clrsig exits 1** — `category` and `in_stock` are missing from the response; `product_type` and `available` are present but not in the spec.

Stop the server before running the fixed version:

```bash
# macOS / Linux
pkill -f 'broken/main.go'

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19010).OwningProcess
```

## Why clrsig catches it

The scenario binds the OpenAPI spec at the top level:

```yaml
specFile: ./openapi.yaml
```

The single step asserts `type: schema` with no inline schema. clrsig looks up the declared schema for `GET /items/{id}` in the spec and validates the full response body against it. The broken server:

- **Missing required fields:** `category` and `in_stock` are in the spec's `required` list but absent from the response
- **Unknown fields:** `product_type` and `available` are in the response but not declared in the spec

Both violations fail the spec-bound schema assertion, and clrsig exits 1.

This is the difference from FM-001/FM-005: you do not need to predict which fields will drift. The spec itself is the validator.

See [specFile binding](https://clrsig.com/docs/format-spec/) in the Clear Signal format spec.

## Fix

```bash
go run failures/fm-010-01-spec-drift/fixed/main.go &
clrsig run failures/fm-010-01-spec-drift/scenarios/detect.yaml \
  --env failures/fm-010-01-spec-drift/environments/local.yaml
# exits 0
```

**What changed:** `broken/main.go` — `ProductType string` → `Category string` and `Available bool` → `InStock bool`, with JSON tags updated to match the spec (`category`, `in_stock`).

## Adapt this to your API

Replace `baseUrl` in `environments/local.yaml` and point `specFile` at your own OpenAPI document:

```yaml
# scenarios/detect.yaml
specFile: ./your-api.yaml
name: drift detection
steps:
  - name: response matches spec
    request:
      method: GET
      path: /your-endpoint/known-id
    assertions:
      - type: schema
```

This pattern scales: add one step per endpoint you want to monitor for drift. No assertions to maintain as the spec evolves — the spec file is the single source of truth.

## GitHub Actions CI setup

This sub-project uses `kind: "server"` in the CI matrix — same as FM-001 and FM-005.

### Matrix entry

```yaml
- { id: "010", slug: "spec-drift", kind: "server" }
```

Add to both `catalog.yml` and `catalog-main.yml`.

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-010-spec-drift\broken\
& "$env:TEMP\server-broken.exe"
# In a second terminal:
clrsig run failures/fm-010-01-spec-drift/scenarios/detect.yaml `
  --env failures/fm-010-01-spec-drift/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19010).OwningProcess

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-010-spec-drift\fixed\
& "$env:TEMP\server-fixed.exe"
# In a second terminal:
clrsig run failures/fm-010-01-spec-drift/scenarios/detect.yaml `
  --env failures/fm-010-01-spec-drift/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
```

## In the wild

Spec/implementation drift accumulates when:

- **Domain model refactors** rename fields for internal clarity (`isAvailable` → `in_stock`) without touching the JSON serialization layer or the spec
- **Multiple teams own the spec and the code** — the backend team renames a field for a migration, the API team updates the spec separately, and the two changes diverge
- **Code-first API development** where the spec is auto-generated from the code at one point in time, then the code continues to evolve while the spec stays frozen
- **Microservice decomposition** — an endpoint moves to a new service that returns a slightly different field structure; the spec for the old service was never updated and the new service was never spec'd
