# fm-034-04 — PATCH Semantics Failure

> Full catalog entry: https://clrsig.com/failures/fm-038/

**Difficulty:** Medium
**Detected by:** `clrsig run` with `type: jsonpath` assertions
**Port:** 19038

**Primary failure mode:** [FM-034 — HTTP Method and Operation Semantics Failures](https://clrsig.com/failures/fm-034/)  

## What this failure is

The server implements PATCH as a full replacement (PUT semantics). Sending `PATCH /users/1` with only `{"email": "new@example.com"}` wipes out the `name` and `role` fields instead of merging the patch into the existing record.

This pattern appears when:
- A developer copies the PUT handler and forgets to implement field-level merging
- An ORM's `save()` method is used instead of `update()`, serializing the whole decoded struct including zero values
- A framework deserializes the request body into an empty struct, leaving unmentioned fields as zero values

The consequence: a partial update silently corrupts the resource. A mobile app updating only a user's email address zeroes out their name. The bug is invisible in a response body check — the returned object looks valid, just with wrong values.

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-034-04-patch-semantics/broken/main.go &
clrsig run failures/fm-034-04-patch-semantics/scenarios/detect.yaml \
  --env failures/fm-034-04-patch-semantics/environments/local.yaml
```

Expected: **clrsig exits 1** — after `PATCH {"email":...}`, step 3 reads back `name: ""` and `role: ""`; assertions expect `"Alice"` and `"admin"`.

Stop the server:

```bash
# macOS / Linux
fuser -k 19038/tcp

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19038).OwningProcess
```

## Why clrsig catches it

The scenario uses a 3-step read-modify-read pattern:

1. **GET** — assert initial `name` and `role`
2. **PATCH** — send only `email`, assert `200`
3. **GET** — assert `name` and `role` are **unchanged**

The broken server zeroes out `name` and `role` in step 2. Step 3's `type: jsonpath` assertions catch the corruption:

```yaml
- type: jsonpath
  path: "$.name"
  operator: eq
  expected: "Alice"
```

## Fix

```bash
go run failures/fm-034-04-patch-semantics/fixed/main.go &
clrsig run failures/fm-034-04-patch-semantics/scenarios/detect.yaml \
  --env failures/fm-034-04-patch-semantics/environments/local.yaml
# exits 0
```

**What changed:** `broken/main.go` — PATCH handler now decodes into `map[string]string` and applies only the keys that were actually provided, leaving all other fields on the existing record untouched.

## GitHub Actions CI setup

```yaml
- { id: "038", slug: "patch-semantics", kind: "server" }
```

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-038-patch-semantics\broken\
& "$env:TEMP\server-broken.exe"
# In a second terminal:
clrsig run failures/fm-034-04-patch-semantics/scenarios/detect.yaml `
  --env failures/fm-034-04-patch-semantics/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19038).OwningProcess

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-038-patch-semantics\fixed\
& "$env:TEMP\server-fixed.exe"
# In a second terminal:
clrsig run failures/fm-034-04-patch-semantics/scenarios/detect.yaml `
  --env failures/fm-034-04-patch-semantics/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
```

## In the wild

This failure appears in:

- **Go APIs** using `json.Decode` into a typed struct — unmentioned fields become zero values, and `save()` persists them
- **Django REST Framework** — using `serializer.save()` without `partial=True` replaces the whole object
- **Rails APIs** — `user.update(params)` with a full `permit` list applies zero values for missing params
- **Copy-paste from PUT handler** — the most common root cause; PATCH handler is a modified PUT that never got proper merge logic
