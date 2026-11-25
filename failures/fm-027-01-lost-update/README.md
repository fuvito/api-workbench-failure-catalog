# fm-027-01 — Concurrent Update and Lost-Update Failures

> Full catalog entry: https://clrsig.com/failures/fm-027/

**Difficulty:** Hard
**Detected by:** `clrsig run` multi-step scenario (status assertions)
**Port:** 19027

**Primary failure mode:** [FM-027 — Concurrent Update and Lost-Update Failures](https://clrsig.com/failures/fm-027/)  

## What this failure is

The server accepts any PUT regardless of whether the client's data is current. When two clients both read version 1 of an item, make independent changes, and then both write, the second write silently overwrites the first. The first client's update is lost with no error, no warning, and no way to know it happened.

The spec requires optimistic locking: each update must include the `version` the client last read. If the server's current version is different — meaning another writer got there first — the server must return `409 Conflict` so the client can re-read, merge, and retry. The broken server ignores the submitted `version` and always accepts the write.

## Scenario flow

| Step | Request | Broken | Fixed |
|------|---------|--------|-------|
| 1 | GET /items/1 | version=1 ✓ | version=1 ✓ |
| 2 | PUT /items/1 version=1 (first write) | 200 — updated to version=2 ✓ | 200 — updated to version=2 ✓ |
| 3 | PUT /items/1 version=1 (stale write) | **200 — silent lost update** ✗ | 409 — rejected ✓ |

Step 3 is the discriminator. After step 2 the server is at version 2; a write with version=1 is stale. The broken server accepts it and clobbers the step-2 update. The fixed server rejects it.

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-027-01-lost-update/broken/main.go &
clrsig run failures/fm-027-01-lost-update/scenarios/detect.yaml \
  --env failures/fm-027-01-lost-update/environments/local.yaml
```

Expected: **clrsig exits 1** — step 3 returns `200 OK`; the assertion expects `409 Conflict`.

Stop the server before running the fixed version:

```bash
# macOS / Linux
pkill -f 'broken/main.go'

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19027).OwningProcess
```

## Why clrsig catches it

The scenario sequences three requests and uses a `type: status` assertion on the third:

```yaml
- name: "stale write — must be rejected with 409"
  request:
    method: PUT
    path: /items/1
    body:
      value: 99
      version: 1
  assertions:
    - type: status
      expected: 409
```

The broken server returns `200` here — it never checks whether version 1 is still current after step 2 advanced the item to version 2. The fixed server checks and rejects.

## Fix

```bash
go run failures/fm-027-01-lost-update/fixed/main.go &
clrsig run failures/fm-027-01-lost-update/scenarios/detect.yaml \
  --env failures/fm-027-01-lost-update/environments/local.yaml
# exits 0
```

**What the demo does:** adds a version check before writing:

```go
// broken — version never checked
item.Value = update.Value
item.Version++

// fixed — stale writes rejected
if update.Version != item.Version {
    writeJSON(w, http.StatusConflict, ErrorResponse{
        Error: "version conflict",
        Code:  "VERSION_CONFLICT",
    })
    return
}
item.Value = update.Value
item.Version++
```

**What the production fix looks like** — use a `WHERE` clause on the version column so the database enforces the check atomically:

```sql
UPDATE items
SET value = $1, version = version + 1
WHERE id = $2 AND version = $3
```

If `rows_affected == 0`, the version didn't match — return `409 Conflict`. This is atomic and correct under concurrent load, unlike an application-layer read-then-write.

Alternatively, use HTTP ETags: the server sets `ETag: "version-2"` on GET responses; clients send `If-Match: "version-1"` on PUT requests; the server returns `412 Precondition Failed` if the ETag doesn't match. Same semantics, standard HTTP mechanism.

## Adapt this to your API

Replace `baseUrl` in `environments/local.yaml` and add three steps for any resource that supports concurrent writes:

```yaml
steps:
  - name: read current version
    request:
      method: GET
      path: /your-resource/1
    assertions:
      - type: jsonpath
        path: $.version
        expected: 1

  - name: first write succeeds
    request:
      method: PUT
      path: /your-resource/1
      body:
        field: "new value"
        version: 1
    assertions:
      - type: status
        expected: 200

  - name: stale write must be rejected
    request:
      method: PUT
      path: /your-resource/1
      body:
        field: "overwrite attempt"
        version: 1
    assertions:
      - type: status
        expected: 409
```

If clrsig exits 1, the server accepted a stale write — data loss is possible under concurrent load.

## GitHub Actions CI setup

This sub-project uses `kind: "server"`.

### Matrix entry

```yaml
- { id: "027", slug: "lost-update", kind: "server" }
```

Add to both `catalog.yml` and `catalog-main.yml`.

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-027-lost-update\broken\
& "$env:TEMP\server-broken.exe"
# In a second terminal:
clrsig run failures/fm-027-01-lost-update/scenarios/detect.yaml `
  --env failures/fm-027-01-lost-update/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19027).OwningProcess

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-027-lost-update\fixed\
& "$env:TEMP\server-fixed.exe"
# In a second terminal:
clrsig run failures/fm-027-01-lost-update/scenarios/detect.yaml `
  --env failures/fm-027-01-lost-update/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
```

## In the wild

Lost-update failures appear when:

- **No version tracking at all** — a REST API exposes PUT without any version or ETag mechanism; concurrent editors silently overwrite each other; this goes unnoticed until a user reports that their change "disappeared"
- **Version checked in application code, not in the database** — the application reads the current version, compares it in memory, then writes; a second concurrent request passes the check between the read and the write
- **ETag generated from timestamp** — two writes in the same millisecond produce identical ETags; the conflict check passes for both
- **PATCH instead of PUT, but no field-level conflict detection** — a PATCH endpoint merges updates but doesn't track which version each field was read from; a slow client overwrites a fast client's changes to unrelated fields
- **Soft-delete race** — a record is "deleted" (soft) while another request is in-flight updating it; the update succeeds and un-deletes the record
