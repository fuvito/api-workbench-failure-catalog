# fm-025-04 — Validation Error Returns 500 Instead of 422

**Difficulty:** Easy | **Detected by:** `type: status` assertion | **Port:** 19043

**Primary failure mode:** [FM-025 — Incorrect HTTP Status and Business-Error Semantics](https://clrsig.com/failures/fm-025/)  

## What this failure is

When required fields are missing from the request body, the server returns `500 Internal Server Error` instead of `422 Unprocessable Entity`. The handler treats missing-field validation failures the same as unexpected runtime errors.

## Reproduce it

```bash
go run failures/fm-043-validation-error-code/broken/main.go &
clrsig run failures/fm-043-validation-error-code/scenarios/detect.yaml \
  --env failures/fm-043-validation-error-code/environments/local.yaml
```

Expected: **exits 1** — both steps receive `500`; assertions expect `422`.

## Fix

```bash
go run failures/fm-043-validation-error-code/fixed/main.go &
clrsig run failures/fm-043-validation-error-code/scenarios/detect.yaml \
  --env failures/fm-043-validation-error-code/environments/local.yaml
# exits 0
```

**What changed:** `fixed/main.go` — validation checks return `422` with `{"error":"...","fields":[...]}`.

## CI

```yaml
- { id: "043", slug: "validation-error-code", kind: "server" }
```
