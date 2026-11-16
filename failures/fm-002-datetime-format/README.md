# FM-002 — Date/Time Format Failures

> Full catalog entry: https://clrsig.com/failures/fm-002/

**Difficulty:** Easy
**Detected by:** `clrsig run` with `specFile:` binding
**Port:** 19002

## What this failure is

The server returns `created_at` as a Unix timestamp integer (`1731196800`) instead of an ISO 8601 string (`"2024-11-10T08:00:00Z"`). The OpenAPI spec declares this field as `type: string, format: date-time`, so the schema validator sees a type mismatch immediately.

This pattern appears when:
- The database stores timestamps as `BIGINT` and the ORM/serializer exposes them directly
- A legacy internal service uses epoch seconds and the API layer forgets to convert
- A refactor changes the backing type from `time.Time` to `int64` without updating the spec

The consequence: consumer code that expects to parse an ISO 8601 string (`new Date(event.created_at)`) receives an integer and silently produces the wrong result or throws.

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-002-datetime-format/broken/main.go &
clrsig run failures/fm-002-datetime-format/scenarios/detect.yaml \
  --env failures/fm-002-datetime-format/scenarios/env.yaml
```

Expected: **clrsig exits 1** — `created_at` has type `integer`; spec requires `type: string`.

Stop the server before running the fixed version:

```bash
# macOS / Linux
fuser -k 19002/tcp

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19002).OwningProcess
```

## Why clrsig catches it

The scenario (`scenarios/detect.yaml`) sends `GET /events/1` and `GET /events/2`. With `specFile: ../openapi.yaml` declared, clrsig automatically validates every response body against the declared schema. The broken server returns:

```json
{"id": 1, "name": "API Summit", "created_at": 1731196800, "scheduled_for": "12/15/2026"}
```

clrsig reports:
```
✗ schema: Error at "/created_at": value must be a string
```

No assertions needed beyond `assertions: []` — the `specFile:` binding does the work.

## Fix

```bash
go run failures/fm-002-datetime-format/fixed/main.go &
clrsig run failures/fm-002-datetime-format/scenarios/detect.yaml \
  --env failures/fm-002-datetime-format/scenarios/env.yaml
# exits 0
```

**What changed:** `broken/main.go` — `CreatedAt` field type changed from `int64` to `string`, value changed from Unix timestamp to ISO 8601. `ScheduledFor` changed from `MM/DD/YYYY` to `YYYY-MM-DD`.

## Adapt this to your API

Replace `baseUrl` in `scenarios/env.yaml` with your API's address:

```yaml
variables:
  - name: baseUrl
    value: https://your-api.example.com
```

Update `scenarios/detect.yaml` to target your endpoint and date fields:

```yaml
specFile: ../openapi.yaml
steps:
  - label: "created_at must be a string"
    request:
      method: GET
      url: "{{baseUrl}}/your-endpoint/1"
    assertions: []
```

As long as your OpenAPI spec declares date fields as `type: string, format: date-time`, the `specFile:` binding catches any server that returns an integer.

## GitHub Actions CI setup

This sub-project is included in the CI matrix in both workflow files. Push the repo to GitHub and it runs automatically.

If you need to add it back after a reset, add to both `catalog.yml` and `catalog-main.yml`:

```yaml
- { id: "002", slug: "datetime-format", kind: "server" }
```

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-002-datetime-format\broken\
& "$env:TEMP\server-broken.exe"
# In a second terminal:
clrsig run failures/fm-002-datetime-format/scenarios/detect.yaml `
  --env failures/fm-002-datetime-format/scenarios/env.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19002).OwningProcess

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-002-datetime-format\fixed\
& "$env:TEMP\server-fixed.exe"
# In a second terminal:
clrsig run failures/fm-002-datetime-format/scenarios/detect.yaml `
  --env failures/fm-002-datetime-format/scenarios/env.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
```

## In the wild

This failure appears in:

- **Unix-epoch APIs** where backend engineers think in timestamps and forget consumers expect ISO 8601
- **ORM leakage** where a `BIGINT` timestamp column is mapped directly to the response without conversion
- **Cross-language serialization** — Go's `time.Time` marshals to RFC 3339 by default, but `int64` does not; a type change in a struct field silently changes the wire format
- **Legacy migrations** where an older internal service used epoch seconds and a new public-facing wrapper was supposed to convert but doesn't
