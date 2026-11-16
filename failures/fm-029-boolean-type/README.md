# FM-029 — Boolean Field Type Confusion

> Full catalog entry: https://clrsig.com/failures/fm-029/

**Difficulty:** Easy
**Detected by:** `clrsig run` with `specFile:` binding
**Port:** 19029

## What this failure is

The server returns boolean fields as strings (`"true"`, `"false"`) or integers (`1`, `0`) instead of JSON booleans (`true`, `false`). The spec declares these fields as `type: boolean`.

This pattern appears when:
- A database driver maps `BIT` or `TINYINT(1)` columns to `int64` and the serializer emits the raw number
- A PHP or Ruby serializer quotes all scalar values, turning `true` into `"true"`
- A Go struct field is typed as `string` or `int` instead of `bool` — the value is correct but the wire type is wrong
- A legacy system stored booleans as `"Y"`/`"N"` or `1`/`0` and the migration left the serialization layer unchanged

The consequence: consumer code doing `if (user.active)` evaluates `"false"` as truthy in JavaScript, `1` as truthy in Python without an explicit `== True` check, and so on. The bug is silent and only manifests in edge cases where the value is `false`/`0`.

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-029-boolean-type/broken/main.go &
clrsig run failures/fm-029-boolean-type/scenarios/detect.yaml \
  --env failures/fm-029-boolean-type/environments/local.yaml
```

Expected: **clrsig exits 1** — `active` and `email_verified` have type `string`; feature flags have type `integer`; spec requires `type: boolean` for all.

Stop the server:

```bash
# macOS / Linux
fuser -k 19029/tcp

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19029).OwningProcess
```

## Why clrsig catches it

With `specFile: ../openapi.yaml`, clrsig validates every response body against the declared schema. `"true"` (string) and `1` (integer) do not satisfy `type: boolean` in JSON Schema:

```
✗ schema: Error at "/active": value must be a boolean
✗ schema: Error at "/dark_mode": value must be a boolean
```

No explicit assertions needed beyond `assertions: []`.

## Fix

```bash
go run failures/fm-029-boolean-type/fixed/main.go &
clrsig run failures/fm-029-boolean-type/scenarios/detect.yaml \
  --env failures/fm-029-boolean-type/environments/local.yaml
# exits 0
```

**What changed:** `broken/main.go` — `UserBroken.Active` and `UserBroken.EmailVerified` changed from `string` to `bool`; `FeatureFlagsBroken` fields changed from `int` to `bool`.

## GitHub Actions CI setup

```yaml
- { id: "029", slug: "boolean-type", kind: "server" }
```

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-029-boolean-type\broken\
& "$env:TEMP\server-broken.exe"
# In a second terminal:
clrsig run failures/fm-029-boolean-type/scenarios/detect.yaml `
  --env failures/fm-029-boolean-type/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19029).OwningProcess

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-029-boolean-type\fixed\
& "$env:TEMP\server-fixed.exe"
# In a second terminal:
clrsig run failures/fm-029-boolean-type/scenarios/detect.yaml `
  --env failures/fm-029-boolean-type/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
```

## In the wild

This failure appears in:

- **MySQL-backed Go APIs** — `database/sql` returns `TINYINT(1)` as `int64`; developers forget to convert to `bool` before serializing
- **PHP APIs** — older versions of `json_encode` quote booleans when they come from an associative array with string keys
- **Rails APIs** — `ActiveRecord` attributes are typed but `as_json` overrides can drop the type conversion
- **Column migrations** — a boolean column was added as `TINYINT` for compatibility, and the serialization layer was never updated to coerce the type
