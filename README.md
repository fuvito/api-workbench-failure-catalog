# api-workbench-failure-catalog

Runnable reproductions of the API contract failure modes that
[clrsig](https://clrsig.com) detects. Each sub-project under `failures/` is a
real, broken (and fixed) server that CI exercises on every commit — a live
regression suite for the detection engine and a copy-pasteable scenario library
for your own workspace.

[![Failure Catalog (released)](https://github.com/fuvito/api-workbench-failure-catalog/actions/workflows/catalog.yml/badge.svg)](https://github.com/fuvito/api-workbench-failure-catalog/actions/workflows/catalog.yml)
[![Failure Catalog (main)](https://github.com/fuvito/api-workbench-failure-catalog/actions/workflows/catalog-main.yml/badge.svg)](https://github.com/fuvito/api-workbench-failure-catalog/actions/workflows/catalog-main.yml)

> **Status:** pre-build. The design source of truth is
> [`docs/PRD.md`](docs/PRD.md). Sub-projects land incrementally; the index
> below marks each as it ships.

## How it works

Each server-based sub-project ships three things:

- `broken/` — a single-file Go server (standard library only) that exhibits the failure
- `fixed/` — the corrected server, same spec, all assertions pass
- `scenarios/detect.yaml` — the clrsig scenario that catches the failure

Run one locally in under two minutes:

```bash
cd failures/fm-001-number-precision
go run broken/main.go &
clrsig run scenarios/detect.yaml --env environments/local.yaml
# exit 1 — failure caught

go run fixed/main.go &
clrsig run scenarios/detect.yaml --env environments/local.yaml
# exit 0 — fix verified
```

Diff-based sub-projects compare two OpenAPI spec versions instead — no server
needed:

```bash
clrsig diff failures/fm-004-renamed-removed-field/v1/openapi.yaml \
            failures/fm-004-renamed-removed-field/v2/openapi.yaml \
            --breaking-only
```

Each scenario is designed to be lifted: copy `detect.yaml` into your workspace,
point `environments/local.yaml` at your API, and run.

## Failure modes

| FM | Failure mode | Difficulty | Detection | Port | Status |
|---|---|---|---|---|---|
| [FM-001](https://clrsig.com/failures/fm-001/) | Currency, Number, and Precision Formatting Failures | Easy | `clrsig run` schema assertion | 19001 | pending |
| [FM-003](https://clrsig.com/failures/fm-003/) | Pagination Contract Failures | Hard | `clrsig run` multi-step scenario | 19003 | pending |
| [FM-004](https://clrsig.com/failures/fm-004/) | Renamed or Removed Response Fields | Medium | `clrsig diff` v1 vs v2 | — | pending |
| [FM-005](https://clrsig.com/failures/fm-005/) | Incorrect Optional, Nullable, and Required Field Behavior | Easy | `clrsig run` schema assertion | 19005 | pending |
| [FM-010](https://clrsig.com/failures/fm-010/) | OpenAPI Specification and Implementation Drift | Easy | `clrsig run` + `specFile:` schema assertions | 19010 | pending |
| [FM-016](https://clrsig.com/failures/fm-016/) | Rate-Limit and Throttling Failures | Medium | `clrsig run` undeclared 429 | 19016 | pending |
| [FM-018](https://clrsig.com/failures/fm-018/) | Idempotency and Duplicate-Operation Failures | Hard | `clrsig run` multi-step scenario | 19018 | pending |
| [FM-022](https://clrsig.com/failures/fm-022/) | Syntactically Valid but Semantically Invalid Payloads | Medium | `clrsig run` jsonpath assertions | 19022 | pending |
| [FM-025](https://clrsig.com/failures/fm-025/) | Incorrect HTTP Status and Business-Error Semantics | Easy | `clrsig run` status assertion | 19025 | pending |
| [FM-027](https://clrsig.com/failures/fm-027/) | Concurrent Update and Lost-Update Failures | Hard | `clrsig run` multi-step scenario | 19027 | pending |
| [FM-034](https://clrsig.com/failures/fm-034/) | HTTP Method and Operation Semantics Failures | Medium | `clrsig run` + `specFile:` schema assertions | 19034 | pending |
| [FM-044](https://clrsig.com/failures/fm-044/) | Sensitive Data Exposure in API Responses | Medium | `clrsig run` schema (`additionalProperties: false`) | 19044 | pending |

The full catalog of all 50 failure modes lives at
[clrsig.com/failures](https://clrsig.com/failures/). The 38 modes not in v1 are
tracked in the backlog (`docs/PRD.md` §8).

## Conventions

- **Ports:** `19000 + FM number` — FM-001 → `19001`, FM-044 → `19044`.
- **Health check:** every server-based sub-project exposes `GET /health` → 200; CI polls it for readiness.
- **Exit codes:** `clrsig` exits `1` on `broken/` (failure caught) and `0` on `fixed/` (fix verified). CI asserts both legs.
- **Dependencies:** servers import only the Go standard library.

## CI

Two independent workflows keep the catalog honest:

- **`catalog.yml`** — runs the matrix against the latest pinned clrsig release on every push and nightly at 06:00 UTC.
- **`catalog-main.yml`** — builds clrsig from `api-workbench@main` nightly at 07:00 UTC to catch engine regressions before they ship.
- **`update-outputs.yml`** — regenerates the expected-output blocks in sub-project READMEs so docs can't drift from real output.

## Contributing

See [`CONTRIBUTING.md`](CONTRIBUTING.md) for the sub-project template, port
convention, `/health` contract, and how to propose a failure mode from the
backlog.

## License

[MIT](LICENSE)
