# FM-011 — Redirect Instead of Response

**Difficulty:** Easy | **Detected by:** `type: status` assertion | **Port:** 19011

## What this failure is

`GET /posts/{id}` returns `301 Moved Permanently` instead of `200` with the post. A redirect middleware was added for a legacy URL migration but fires for all requests.

## Reproduce it

```bash
go run failures/fm-011-redirect-instead-of-response/broken/main.go &
clrsig run failures/fm-011-redirect-instead-of-response/scenarios/detect.yaml \
  --env failures/fm-011-redirect-instead-of-response/environments/local.yaml
```

Expected: **exits 1** — both steps receive `301`; assertions expect `200`.

## Fix

```bash
go run failures/fm-011-redirect-instead-of-response/fixed/main.go &
clrsig run failures/fm-011-redirect-instead-of-response/scenarios/detect.yaml \
  --env failures/fm-011-redirect-instead-of-response/environments/local.yaml
# exits 0
```

**What changed:** `fixed/main.go` — the redirect call removed; handler returns the post directly.

## CI

```yaml
- { id: "011", slug: "redirect-instead-of-response", kind: "server" }
```
