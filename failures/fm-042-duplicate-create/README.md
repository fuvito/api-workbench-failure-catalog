# FM-042 — Duplicate Create Returns 200 Instead of 409

**Difficulty:** Easy | **Detected by:** `type: status` assertion | **Port:** 19042

## What this failure is

Creating a user with an email that already exists returns `201 Created` and inserts a duplicate record instead of returning `409 Conflict`. The uniqueness constraint is never checked.

## Reproduce it

```bash
go run failures/fm-042-duplicate-create/broken/main.go &
clrsig run failures/fm-042-duplicate-create/scenarios/detect.yaml \
  --env failures/fm-042-duplicate-create/environments/local.yaml
```

Expected: **exits 1** — second POST with same email returns `201`; assertion expects `409`.

## Fix

```bash
go run failures/fm-042-duplicate-create/fixed/main.go &
clrsig run failures/fm-042-duplicate-create/scenarios/detect.yaml \
  --env failures/fm-042-duplicate-create/environments/local.yaml
# exits 0
```

**What changed:** `fixed/main.go` — scans existing users for email collision and returns `409` before inserting.

## CI

```yaml
- { id: "042", slug: "duplicate-create", kind: "server" }
```
