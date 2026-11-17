# FM-039 — Bulk Delete No-Op

**Difficulty:** Easy | **Detected by:** `type: status` assertion | **Port:** 19039

## What this failure is

`POST /items/bulk-delete` returns `204 No Content` but performs no deletion. The handler decodes the request body and discards the IDs without touching the store. Items that should be gone are still retrievable.

## Reproduce it

```bash
go run failures/fm-039-bulk-delete/broken/main.go &
clrsig run failures/fm-039-bulk-delete/scenarios/detect.yaml \
  --env failures/fm-039-bulk-delete/environments/local.yaml
```

Expected: **exits 1** — after bulk-delete, `GET /items/1` returns `200` (item still exists); assertion expects `404`.

## Why clrsig catches it

Step 3 asserts `status 404` on `GET /items/1` after the bulk delete. The broken server never actually deleted item 1, so it returns `200`, failing the assertion.

## Fix

```bash
go run failures/fm-039-bulk-delete/fixed/main.go &
clrsig run failures/fm-039-bulk-delete/scenarios/detect.yaml \
  --env failures/fm-039-bulk-delete/environments/local.yaml
# exits 0
```

**What changed:** `fixed/main.go` calls `delete(items, id)` for each ID in the request body.

## CI

```yaml
- { id: "039", slug: "bulk-delete", kind: "server" }
```
