# FM-024 — API Version Routing

**Difficulty:** Easy | **Detected by:** `specFile:` binding | **Port:** 19024

## What this failure is

The `/v2/products` endpoint returns a v1-shaped response — missing the `sku` field that v2 adds as required. The router forwards `/v2/` requests to the same handler as `/v1/` without upgrading the response shape.

## Reproduce it

```bash
go run failures/fm-024-version-routing/broken/main.go &
clrsig run failures/fm-024-version-routing/scenarios/detect.yaml \
  --env failures/fm-024-version-routing/environments/local.yaml
```

Expected: **exits 1** — `sku` is required in the v2 spec but absent from every product in the response.

## Why clrsig catches it

`specFile: ../openapi.yaml` declares `sku` as a required field on `Product`. The broken server omits it, so the spec binding reports a missing required property.

## Fix

```bash
go run failures/fm-024-version-routing/fixed/main.go &
clrsig run failures/fm-024-version-routing/scenarios/detect.yaml \
  --env failures/fm-024-version-routing/environments/local.yaml
# exits 0
```

**What changed:** `fixed/main.go` uses a `ProductV2` struct with `SKU string` populated.

## CI

```yaml
- { id: "024", slug: "version-routing", kind: "server" }
```
