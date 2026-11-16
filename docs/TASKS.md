# Task List

> Source of truth: [`PRD.md`](PRD.md). This file tracks build progress against the PRD task list.
>
> **Build order:** Easy failure modes first → Medium → Hard (PRD §6.5).

## 7.1 Repository Setup

- [x] Create `github.com/fuvito/api-workbench-failure-catalog` (public)
- [x] Root `README.md` with CI badges and sub-project index
- [x] Root `go.mod` (`module github.com/fuvito/api-workbench-failure-catalog`)
- [x] `internal/testutil/wait.go` HTTP readiness poll helper
- [x] `.gitignore`
- [x] `LICENSE` (MIT)
- [x] `CONTRIBUTING.md` with sub-project template and port convention

## 7.2 CI Scaffolding

- [ ] `.github/workflows/catalog.yml` — push + nightly 06:00 UTC, released clrsig
- [ ] `.github/workflows/catalog-main.yml` — nightly 07:00 UTC only, builds clrsig from api-workbench@main
- [ ] `.github/workflows/update-outputs.yml` — regenerates expected-output blocks in sub-project READMEs
- [x] CI badges in root `README.md`
- [ ] `verify.sh` / `verify.ps1` — cross-platform local verification wrappers
- [ ] (optional) Renovate `regexManagers` rule to auto-update `CLRSIG_VERSION` on new releases

## 7.3 Pre-build Decisions

- [x] `clrsig diff` exits 1 on breaking changes only — confirmed
- [x] Spec-only layout: `v1/openapi.yaml` + `v2/openapi.yaml` adopted
- [x] All 12 sub-projects required to ship with `fixed/` variant

## 7.4 Sub-projects

### Easy (build first — validates the full pipeline)

- [x] **FM-025** — Incorrect HTTP Status and Business-Error Semantics
  - [x] `broken/main.go` — returns 200 for missing products
  - [x] `fixed/main.go` — returns 404 for missing products
  - [x] `openapi.yaml`
  - [x] `scenarios/detect.yaml`
  - [x] `environments/local.yaml`
  - [x] `README.md`
  - [x] CI matrix entry in `catalog.yml` and `catalog-main.yml`

- [x] **FM-001** — Currency, Number, and Precision Formatting Failures
  - [x] `broken/main.go` — monetary fields typed as `string`, returns `"24.00"`
  - [x] `fixed/main.go` — monetary fields typed as `float64`, returns `24.00`
  - [x] `openapi.yaml`
  - [x] `scenarios/detect.yaml`
  - [x] `environments/local.yaml`
  - [x] `README.md`
  - [x] CI matrix entry

- [x] **FM-005** — Incorrect Optional, Nullable, and Required Field Behavior
  - [x] `broken/main.go` — `Email *string`, returns `"email": null` for OAuth users
  - [x] `fixed/main.go` — `Email string`, always returns a non-null string
  - [x] `openapi.yaml`
  - [x] `scenarios/detect.yaml`
  - [x] `environments/local.yaml`
  - [x] `README.md`
  - [x] CI matrix entry

- [x] **FM-010** — OpenAPI Specification and Implementation Drift
  - [x] `broken/main.go` — fields renamed to `product_type`, `available` (spec says `category`, `in_stock`)
  - [x] `fixed/main.go` — fields restored to match spec
  - [x] `openapi.yaml`
  - [x] `scenarios/detect.yaml` — uses `specFile:` binding, single `type: schema` assertion
  - [x] `environments/local.yaml`
  - [x] `README.md`
  - [x] CI matrix entry

### Medium

- [x] **FM-004** — Renamed or Removed Response Fields *(diff-based — `kind: "diff"`)*
  - [x] `v1/openapi.yaml` — `full_name`, `email`, `phone` as required fields
  - [x] `v2/openapi.yaml` — `full_name` renamed to `name`, `phone` removed (2 breaking changes)
  - [x] `scenarios/detect.yaml` — documents the diff command and expected violations
  - [x] `README.md` — includes non-breaking migration pattern
  - [x] CI matrix entry (`kind: "diff"`)
- [x] **FM-016** — Rate-Limit and Throttling Failures
  - [x] `broken/main.go` — throttles at 2 requests (global counter), returns undeclared 429
  - [x] `fixed/main.go` — limit raised to 1000, scenario never triggers throttling
  - [x] `openapi.yaml` — declares 200 and 404 only (429 intentionally absent)
  - [x] `scenarios/detect.yaml` — 3 steps with `specFile:` binding, step 3 hits the limit
  - [x] `environments/local.yaml`
  - [x] `README.md` — includes production fix pattern (declare 429 + Retry-After)
  - [x] CI matrix entry
- [x] **FM-022** — Syntactically Valid but Semantically Invalid Payloads
  - [x] `broken/main.go` — `total` set to `unit_price` instead of `unit_price × quantity`
  - [x] `fixed/main.go` — correct totals (30.00 and 49.98)
  - [x] `openapi.yaml`
  - [x] `scenarios/detect.yaml` — jsonpath assertions on `$.total`
  - [x] `environments/local.yaml`
  - [x] `README.md`
  - [x] CI matrix entry
- [x] **FM-034** — HTTP Method and Operation Semantics Failures
  - [x] `broken/main.go` — POST returns 200 (should be 201); DELETE returns 200+body (should be 204)
  - [x] `fixed/main.go` — correct status codes and Location header
  - [x] `openapi.yaml`
  - [x] `scenarios/detect.yaml` — specFile binding, type: schema on POST and DELETE
  - [x] `environments/local.yaml`
  - [x] `README.md`
  - [x] CI matrix entry
- [x] **FM-044** — Sensitive Data Exposure in API Responses
  - [x] `broken/main.go` — full User struct serialised, leaks password_hash, api_key, role
  - [x] `fixed/main.go` — explicit UserResponse struct, sensitive fields never serialised
  - [x] `openapi.yaml` — User schema with additionalProperties: false
  - [x] `scenarios/detect.yaml` — specFile binding, type: schema on two users
  - [x] `environments/local.yaml`
  - [x] `README.md`
  - [x] CI matrix entry

### Hard

- [x] **FM-003** — Pagination Contract Failures
  - [x] `broken/main.go` — has_more always true, traps clients in infinite loop
  - [x] `fixed/main.go` — has_more computed as end < len(products)
  - [x] `openapi.yaml`
  - [x] `scenarios/detect.yaml` — 3 steps across pages 1, 2, 3; step 3 is the discriminator
  - [x] `environments/local.yaml`
  - [x] `README.md`
  - [x] CI matrix entry
- [x] **FM-018** — Idempotency and Duplicate-Operation Failures
  - [x] `broken/main.go` — no duplicate check, every POST creates a new payment
  - [x] `fixed/main.go` — order_id lookup before insert, returns existing payment on retry
  - [x] `openapi.yaml`
  - [x] `scenarios/detect.yaml` — 3 steps: submit, retry, count; step 3 is the discriminator
  - [x] `environments/local.yaml`
  - [x] `README.md`
  - [x] CI matrix entry
- [x] **FM-027** — Concurrent Update and Lost-Update Failures
  - [x] `broken/main.go` — PUT ignores version field, always overwrites (silent lost update)
  - [x] `fixed/main.go` — version check before write, 409 on stale version
  - [x] `openapi.yaml`
  - [x] `scenarios/detect.yaml` — 3 steps: read, first write, stale write; step 3 is the discriminator
  - [x] `environments/local.yaml`
  - [x] `README.md`
  - [x] CI matrix entry

## 7.5 Website Integration

- [ ] Verify `clrsig.com/failures/fm-NNN/` page structure can link to GitHub directory
- [ ] Sub-project READMEs reviewed as source material for product pages
- [ ] Each live sub-project linked from its `clrsig.com/failures/fm-NNN/` page
- [ ] Catalog CI badge added to `clrsig.com` landing page

## 7.6 Release Gate

- [ ] All 12 sub-projects passing CI (broken exits 1, fixed exits 0)
- [ ] Root README index complete with correct FM links and status
- [ ] Repository announced on product site and in release notes
