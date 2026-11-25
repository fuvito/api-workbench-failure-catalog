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
cd failures/fm-001-01-number-precision
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
clrsig diff failures/fm-004-01-renamed-removed-field/v1/openapi.yaml \
            failures/fm-004-01-renamed-removed-field/v2/openapi.yaml \
            --breaking-only
```

Each scenario is designed to be lifted: copy `detect.yaml` into your workspace,
point `environments/local.yaml` at your API, and run.

## Failure Mode Index

Each sub-project is an executable example of an official [clrsig.com](https://clrsig.com/failures/) failure mode.
One official failure mode may have multiple concrete examples here.

| # | Official FM | Title | Examples |
|---|---|---|---|
| 1 | [FM-001](https://clrsig.com/failures/fm-001/) | Currency, Number, and Precision Formatting Failures | [fm-001-01-number-precision](failures/fm-001-01-number-precision/) · [fm-001-02-datetime-format](failures/fm-001-02-datetime-format/) |
| 2 | [FM-003](https://clrsig.com/failures/fm-003/) | Pagination Contract Failures | [fm-003-01-pagination-contract](failures/fm-003-01-pagination-contract/) · [fm-003-02-page-overflow](failures/fm-003-02-page-overflow/) · [fm-003-03-page-offset](failures/fm-003-03-page-offset/) · [fm-003-04-missing-pagination-meta](failures/fm-003-04-missing-pagination-meta/) |
| 3 | [FM-004](https://clrsig.com/failures/fm-004/) | Renamed or Removed Response Fields | [fm-004-01-renamed-removed-field](failures/fm-004-01-renamed-removed-field/) · [fm-004-02-field-name-casing](failures/fm-004-02-field-name-casing/) |
| 4 | [FM-005](https://clrsig.com/failures/fm-005/) | Incorrect Optional, Nullable, and Required Field Behavior | [fm-005-01-nullable-required](failures/fm-005-01-nullable-required/) · [fm-005-02-null-collection](failures/fm-005-02-null-collection/) · [fm-005-03-nested-required](failures/fm-005-03-nested-required/) · [fm-005-04-default-not-applied](failures/fm-005-04-default-not-applied/) |
| 5 | [FM-007](https://clrsig.com/failures/fm-007/) | API Versioning and Compatibility Failures | [fm-007-01-version-routing](failures/fm-007-01-version-routing/) |
| 6 | [FM-010](https://clrsig.com/failures/fm-010/) | OpenAPI Specification and Implementation Drift | [fm-010-01-spec-drift](failures/fm-010-01-spec-drift/) · [fm-010-02-error-schema](failures/fm-010-02-error-schema/) · [fm-010-03-bare-array](failures/fm-010-03-bare-array/) · [fm-010-04-object-as-string](failures/fm-010-04-object-as-string/) · [fm-010-05-boolean-type](failures/fm-010-05-boolean-type/) · [fm-010-06-array-item-type](failures/fm-010-06-array-item-type/) · [fm-010-07-minimum-value](failures/fm-010-07-minimum-value/) |
| 7 | [FM-016](https://clrsig.com/failures/fm-016/) | Rate-Limit and Throttling Failures | [fm-016-01-rate-limit](failures/fm-016-01-rate-limit/) |
| 8 | [FM-017](https://clrsig.com/failures/fm-017/) | Stale Cache and Cache-Invalidation Failures | [fm-017-01-stale-read](failures/fm-017-01-stale-read/) |
| 9 | [FM-018](https://clrsig.com/failures/fm-018/) | Idempotency and Duplicate-Operation Failures | [fm-018-01-idempotency](failures/fm-018-01-idempotency/) · [fm-018-02-duplicate-create](failures/fm-018-02-duplicate-create/) |
| 10 | [FM-020](https://clrsig.com/failures/fm-020/) | Distributed Transaction and Partial-Workflow Failures | [fm-020-01-batch-atomicity](failures/fm-020-01-batch-atomicity/) |
| 11 | [FM-022](https://clrsig.com/failures/fm-022/) | Syntactically Valid but Semantically Invalid Payloads | [fm-022-01-semantically-invalid](failures/fm-022-01-semantically-invalid/) · [fm-022-02-filter-ignored](failures/fm-022-02-filter-ignored/) · [fm-022-03-enum-violation](failures/fm-022-03-enum-violation/) · [fm-022-04-sort-order](failures/fm-022-04-sort-order/) · [fm-022-05-limit-ignored](failures/fm-022-05-limit-ignored/) · [fm-022-06-case-insensitive-enum](failures/fm-022-06-case-insensitive-enum/) · [fm-022-07-empty-string-accepted](failures/fm-022-07-empty-string-accepted/) · [fm-022-08-conditional-required](failures/fm-022-08-conditional-required/) · [fm-022-09-id-mismatch](failures/fm-022-09-id-mismatch/) |
| 12 | [FM-023](https://clrsig.com/failures/fm-023/) | Broken Object-Level Authorization and Tenant Isolation | [fm-023-01-object-level-authz](failures/fm-023-01-object-level-authz/) |
| 13 | [FM-025](https://clrsig.com/failures/fm-025/) | Incorrect HTTP Status and Business-Error Semantics | [fm-025-01-incorrect-status](failures/fm-025-01-incorrect-status/) · [fm-025-02-auth-error-code](failures/fm-025-02-auth-error-code/) · [fm-025-03-redirect-instead-of-response](failures/fm-025-03-redirect-instead-of-response/) · [fm-025-04-validation-error-code](failures/fm-025-04-validation-error-code/) · [fm-025-05-empty-result-404](failures/fm-025-05-empty-result-404/) · [fm-025-06-page-out-of-range-500](failures/fm-025-06-page-out-of-range-500/) |
| 14 | [FM-027](https://clrsig.com/failures/fm-027/) | Concurrent Update and Lost-Update Failures | [fm-027-01-lost-update](failures/fm-027-01-lost-update/) |
| 15 | [FM-028](https://clrsig.com/failures/fm-028/) | Long-Running Operation and Asynchronous API Failures | [fm-028-01-async-job](failures/fm-028-01-async-job/) |
| 16 | [FM-034](https://clrsig.com/failures/fm-034/) | HTTP Method and Operation Semantics Failures | [fm-034-01-method-semantics](failures/fm-034-01-method-semantics/) · [fm-034-02-method-not-allowed](failures/fm-034-02-method-not-allowed/) · [fm-034-03-nonidempotent-get](failures/fm-034-03-nonidempotent-get/) · [fm-034-04-patch-semantics](failures/fm-034-04-patch-semantics/) · [fm-034-05-soft-delete](failures/fm-034-05-soft-delete/) · [fm-034-06-bulk-delete](failures/fm-034-06-bulk-delete/) |
| 17 | [FM-038](https://clrsig.com/failures/fm-038/) | Eventual Consistency and Read-After-Write Failures | [fm-038-01-write-not-persisted](failures/fm-038-01-write-not-persisted/) |
| 18 | [FM-044](https://clrsig.com/failures/fm-044/) | Sensitive Data Exposure in API Responses | [fm-044-01-sensitive-data](failures/fm-044-01-sensitive-data/) |

## Failure Modes without Examples

Official [clrsig.com](https://clrsig.com/failures/) failure modes not yet covered in this catalog.
As examples are added, rows move up into the index above.

| # | Official FM | Title |
|---|---|---|
| 1 | [FM-002](https://clrsig.com/failures/fm-002/) | Silent Timeout and Incomplete Response Failures |
| 2 | [FM-006](https://clrsig.com/failures/fm-006/) | Mobile and Older-Client Compatibility Failures |
| 3 | [FM-008](https://clrsig.com/failures/fm-008/) | OAuth Implementation and Token-Validation Failures |
| 4 | [FM-009](https://clrsig.com/failures/fm-009/) | Healthcare Interoperability and HL7 Conformance Failures |
| 5 | [FM-011](https://clrsig.com/failures/fm-011/) | Payment Gateway Certification and Protocol Conformance Failures |
| 6 | [FM-012](https://clrsig.com/failures/fm-012/) | Accessibility Conformance Failures in API-Driven Applications |
| 7 | [FM-013](https://clrsig.com/failures/fm-013/) | Undocumented Dependencies and Accumulated Integration Failures |
| 8 | [FM-014](https://clrsig.com/failures/fm-014/) | Timezone and Daylight-Saving-Time Failures |
| 9 | [FM-015](https://clrsig.com/failures/fm-015/) | Deprecated Endpoint and Lifecycle Failures |
| 10 | [FM-019](https://clrsig.com/failures/fm-019/) | Webhook Delivery and Event-Processing Failures |
| 11 | [FM-021](https://clrsig.com/failures/fm-021/) | File Upload and Downstream Processing Failures |
| 12 | [FM-024](https://clrsig.com/failures/fm-024/) | Localization and Machine-Readable Data Failures |
| 13 | [FM-026](https://clrsig.com/failures/fm-026/) | HTTP Content-Negotiation Failures |
| 14 | [FM-029](https://clrsig.com/failures/fm-029/) | Bulk API Partial-Success Failures |
| 15 | [FM-030](https://clrsig.com/failures/fm-030/) | API Key and Client-Credential Rotation Failures |
| 16 | [FM-031](https://clrsig.com/failures/fm-031/) | CORS and Browser-Based API Access Failures |
| 17 | [FM-032](https://clrsig.com/failures/fm-032/) | TLS/mTLS and Certificate Lifecycle Failures |
| 18 | [FM-033](https://clrsig.com/failures/fm-033/) | API Gateway Routing and Policy Drift Failures |
| 19 | [FM-035](https://clrsig.com/failures/fm-035/) | Request Header Handling and Validation Failures |
| 20 | [FM-036](https://clrsig.com/failures/fm-036/) | Request Size and Resource Exhaustion Failures |
| 21 | [FM-037](https://clrsig.com/failures/fm-037/) | Input Injection and Unsafe Interpretation Failures |
| 22 | [FM-039](https://clrsig.com/failures/fm-039/) | Out-of-Order Event and State Transition Failures |
| 23 | [FM-040](https://clrsig.com/failures/fm-040/) | Database Transaction and Atomicity Failures |
| 24 | [FM-041](https://clrsig.com/failures/fm-041/) | Retry Storm and Cascading Failure |
| 25 | [FM-042](https://clrsig.com/failures/fm-042/) | Dependency Failure and Fault Isolation |
| 26 | [FM-043](https://clrsig.com/failures/fm-043/) | Mass Assignment and Unauthorized Field Modification |
| 27 | [FM-045](https://clrsig.com/failures/fm-045/) | Authentication and Authorization Confusion Across Services |
| 28 | [FM-046](https://clrsig.com/failures/fm-046/) | Security Policy and Secret Leakage Failures |
| 29 | [FM-047](https://clrsig.com/failures/fm-047/) | Character Encoding and Unicode Handling Failures |
| 30 | [FM-048](https://clrsig.com/failures/fm-048/) | Streaming and Large-Response Handling Failures |
| 31 | [FM-049](https://clrsig.com/failures/fm-049/) | API Documentation and Example Drift |
| 32 | [FM-050](https://clrsig.com/failures/fm-050/) | Environment Configuration and Feature-Flag Drift |

## Conventions

- **Ports:** each sub-project has an assigned port in the `190xx` range; see the CI matrix in `.github/workflows/catalog.yml` for the full list.
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
