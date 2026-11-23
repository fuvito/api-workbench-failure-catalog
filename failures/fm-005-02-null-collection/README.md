# fm-005-02 — Empty Collection vs Null

> Full catalog entry: https://clrsig.com/failures/fm-023/

**Difficulty:** Easy
**Detected by:** `clrsig run` with `specFile:` binding
**Port:** 19023

**Primary failure mode:** [FM-005 — Incorrect Optional, Nullable, and Required Field Behavior](https://clrsig.com/failures/fm-005/)  

## What this failure is

The server returns `null` for empty lists instead of `[]`. The spec declares collection fields as `type: array`, which does not permit `null` without an explicit `nullable: true`. The broken server uses uninitialized Go slices that marshal to `null`.

This pattern appears when:
- A Go struct field is `[]T` and never initialized — `nil` marshals to `null`, not `[]`
- An ORM returns `NULL` from a LEFT JOIN on an empty relation and the serializer passes it through
- A developer initializes the field only when there are results: `if len(items) > 0 { resp.Items = items }`

The consequence: consumer code that iterates `response.members.forEach(...)` throws a null pointer exception or `TypeError: cannot read property 'forEach' of null` on empty-but-valid results, treating a normal business case as an error.

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-005-02-null-collection/broken/main.go &
clrsig run failures/fm-005-02-null-collection/scenarios/detect.yaml \
  --env failures/fm-005-02-null-collection/environments/local.yaml
```

Expected: **clrsig exits 1** — team 2 returns `{"members":null}` and project 2 returns `{"tags":null}`; spec requires `type: array`.

Stop the server:

```bash
# macOS / Linux
fuser -k 19023/tcp

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19023).OwningProcess
```

## Why clrsig catches it

With `specFile: ../openapi.yaml`, clrsig validates every response body against the declared schema. `null` does not satisfy `type: array` in JSON Schema (OpenAPI 3.0), so the validator rejects it:

```
✗ schema: Error at "/members": value must be an array
```

No explicit assertions needed — `assertions: []` is sufficient.

## Fix

```bash
go run failures/fm-005-02-null-collection/fixed/main.go &
clrsig run failures/fm-005-02-null-collection/scenarios/detect.yaml \
  --env failures/fm-005-02-null-collection/environments/local.yaml
# exits 0
```

**What changed:** `broken/main.go` — empty entries in `teamMembers` and `projectTags` maps changed from `nil` to `{}` (initialized slices), so Go's JSON encoder emits `[]` instead of `null`.

## GitHub Actions CI setup

```yaml
- { id: "023", slug: "null-collection", kind: "server" }
```

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-023-null-collection\broken\
& "$env:TEMP\server-broken.exe"
# In a second terminal:
clrsig run failures/fm-005-02-null-collection/scenarios/detect.yaml `
  --env failures/fm-005-02-null-collection/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19023).OwningProcess

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-023-null-collection\fixed\
& "$env:TEMP\server-fixed.exe"
# In a second terminal:
clrsig run failures/fm-005-02-null-collection/scenarios/detect.yaml `
  --env failures/fm-005-02-null-collection/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
```

## In the wild

This failure appears in:

- **Go APIs** — the most common source: a nil `[]T` field is left uninitialized and marshals to `null`; the fix is `make([]T, 0)` or `[]T{}`
- **ORM-backed list endpoints** — `db.Where(...).Find(&items)` on an empty result set may leave `items` as nil depending on the ORM version
- **Conditional population** — `if len(results) > 0 { resp.Items = results }` accidentally omits initialization for the empty case
- **Legacy APIs evolving toward arrays** — a field that used to be a single object was changed to an array, and the `null` sentinel from the old shape was carried forward
