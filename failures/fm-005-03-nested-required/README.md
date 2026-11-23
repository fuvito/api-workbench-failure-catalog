# fm-005-03 — Nested Required Field Missing

**Difficulty:** Easy | **Detected by:** `specFile:` binding | **Port:** 19032

**Primary failure mode:** [FM-005 — Incorrect Optional, Nullable, and Required Field Behavior](https://clrsig.com/failures/fm-005/)  

## What this failure is

`shipping_address.street` is omitted when it is an empty string. Go's `json:",omitempty"` tag silently drops the field when empty, but the spec declares `street` as required inside the `Address` schema.

## Reproduce it

```bash
go run failures/fm-005-03-nested-required/broken/main.go &
clrsig run failures/fm-005-03-nested-required/scenarios/detect.yaml \
  --env failures/fm-005-03-nested-required/environments/local.yaml
```

Expected: **exits 1** — order 2's `shipping_address` is missing the required `street` property.

## Fix

```bash
go run failures/fm-005-03-nested-required/fixed/main.go &
clrsig run failures/fm-005-03-nested-required/scenarios/detect.yaml \
  --env failures/fm-005-03-nested-required/environments/local.yaml
# exits 0
```

**What changed:** `fixed/main.go` — `street` field uses plain `json:"street"` (no `omitempty`); data has a real value.

## CI

```yaml
- { id: "032", slug: "nested-required", kind: "server" }
```
