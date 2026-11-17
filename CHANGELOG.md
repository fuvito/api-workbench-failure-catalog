# Changelog

All notable changes to this project will be documented here.

Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
Versions map to GitHub releases; use the release tag when cutting one.

---

## [Unreleased]

### Added

- **FM-014** — Unsupported Method Returns Wrong Status (Easy, port 19014)
  - Broken server accepts any method on a read-only endpoint; DELETE returns 200 instead of 405
  - Detected by status assertion: DELETE expects 405

- **FM-013** — Response ID Mismatch (Easy, port 19013)
  - Broken server ignores path ID and always returns the first record; GET /products/5 returns id=1
  - Detected by jsonpath assertion: `$.id == 5`

- **FM-012** — Off-by-One in Page Offset (Easy, port 19012)
  - Broken server uses `page*perPage` instead of `(page-1)*perPage`, skipping the first page of results
  - Detected by jsonpath assertion: `$.items[0].id == 1` on page 1 and `$.items[0].id == 4` on page 2

- **FM-009** — Field Name Casing Error (Easy, port 19009)
  - Broken server returns camelCase fields (`userId`, `firstName`); spec declares snake_case (`user_id`, `first_name`)
  - Detected by `specFile:` binding — required snake_case fields missing; `additionalProperties: false` rejects camelCase

- **FM-008** — Bare Array Instead of Object Wrapper (Easy, port 19008)
  - Broken server returns `GET /users` as a bare JSON array; spec declares an object with `items` and `count`
  - Detected by `specFile:` binding — array fails `type: object` check

- **FM-041** — Minimum Value Constraint Violated (Easy, port 19041)
  - Broken server returns products with negative `price` and `stock` values
  - Detected by `specFile:` binding — `minimum: 0` constraint in JSON Schema validation

- **FM-036** — Soft Delete Becomes Hard Delete (Easy, port 19036)
  - Broken server permanently removes item on DELETE and returns 204; spec requires 200 with `deleted_at`
  - Detected by status assertion: DELETE expects 200; GET after delete expects 200

- **FM-033** — Non-Idempotent GET (Easy, port 19033)
  - Broken server mutates ticket state on GET — marks it "read" after first access
  - Detected by two consecutive GET assertions: both expect `$.status == "unread"`

- **FM-021** — Nested Object Serialized as String (Easy, port 19021)
  - Broken server returns `metadata` as a JSON-encoded string instead of a nested object
  - Detected by `specFile:` binding — `type: object` fails against actual string value

- **FM-017** — Stale Read After Write (Medium, port 19017)
  - Broken server caches GET responses and never invalidates the cache on PUT
  - Detected by jsonpath assertion: `$.name == "Updated"` after a PUT is ignored by stale cache

- **FM-043** — Validation Error Returns 500 Instead of 422 (Easy, port 19043)
  - Broken server treats missing-field validation the same as runtime errors, returning 500
  - Detected by status assertion: expects 422

- **FM-042** — Duplicate Create Returns 200 Instead of 409 (Easy, port 19042)
  - Broken server inserts duplicate records without checking uniqueness constraint
  - Detected by status assertion: second POST with same email expects 409

- **FM-032** — Nested Required Field Missing (Easy, port 19032)
  - `omitempty` on `street` field silently drops it when empty; spec declares it required
  - Detected by `specFile:` binding — missing required property inside nested Address object

- **FM-031** — Page Overflow Returns Wrong Data (Easy, port 19031)
  - Broken server wraps page number with modulo; page=99 returns page=1 data
  - Detected by jsonpath assertion: `$.count == 0` for out-of-range page

- **FM-011** — Redirect Instead of Response (Easy, port 19011)
  - Broken server returns 301 for all requests due to misconfigured redirect middleware
  - Detected by status assertion: expects 200

- **FM-039** — Bulk Delete No-Op (Easy, port 19039)
  - Broken server returns 204 but never removes items; GET after delete still returns 200
  - Detected by status assertion: step 3 expects 404 on deleted item

- **FM-028** — Filter Parameter Ignored (Easy, port 19028)
  - Broken server ignores `?status=` filter and returns all orders every time
  - Detected by jsonpath assertion on `$.count` for completed and pending filters

- **FM-026** — Async Job Status Never Transitions (Easy, port 19026)
  - Broken server sets job status to "processing" and never advances it to "completed"
  - Detected by jsonpath assertion `$.status == "completed"` on GET /jobs/1

- **FM-024** — API Version Routing (Easy, port 19024)
  - Broken server routes /v2/products to v1 handler; required `sku` field is missing
  - Detected by `specFile:` binding — missing required property in v2 schema

- **FM-020** — Batch Operation Atomicity (Medium, port 19020)
  - Broken server partially creates valid items in a batch containing an invalid one
  - Detected by status assertion: expects 422 (reject all), broken returns 201 (partial success)

- **FM-038** — PATCH Semantics Failure (Medium, port 19038)
  - Broken server replaces entire resource on PATCH instead of merging only provided fields
  - Detected by 3-step read-modify-read scenario; jsonpath assertions catch zeroed-out fields after PATCH

- **FM-035** — Sort Order Contract (Easy, port 19035)
  - Broken server ignores `sort` query parameter and always returns insertion order
  - Detected by jsonpath assertions on `$.items[0].price` for ascending and descending sorts

- **FM-030** — Enum Value Violation (Easy, port 19030)
  - Broken server returns wrong-case (`"PENDING"`) and undeclared (`"SHIPPED"`) enum values
  - Detected by `specFile:` binding — JSON Schema enum validation is exact and case-sensitive

- **FM-029** — Boolean Field Type Confusion (Easy, port 19029)
  - Broken server returns booleans as strings (`"true"`) or integers (`1`/`0`) instead of JSON booleans
  - Detected by `specFile:` binding — string/integer types do not satisfy `type: boolean`

- **FM-023** — Empty Collection vs Null (Easy, port 19023)
  - Broken server returns `null` for empty arrays; Go nil slices marshal to null instead of `[]`
  - Detected by `specFile:` binding — `null` does not satisfy `type: array` in OpenAPI 3.0

- **FM-007** — Authentication Error Code Confusion (Easy, port 19007)
  - Broken server returns 200 with `{"error":"..."}` instead of 401 on bad credentials and missing token
  - Detected by explicit `type: status` assertions — specFile binding alone cannot catch wrong status codes

- **FM-006** — Inconsistent Error Response Schema (Easy, port 19006)
  - Broken server returns different error shapes per endpoint (`message`, `code+msg`, `error`)
  - Detected by `specFile:` binding — non-conforming error bodies fail against the declared `Error` schema

- **FM-002** — Date/Time Format Failures (Easy, port 19002)
  - Broken server returns `created_at` as a Unix timestamp integer; spec declares `type: string, format: date-time`
  - Detected by `specFile:` binding — integer type mismatch against declared `string` fails automatically

- **FM-027** — Concurrent Update and Lost-Update Failures (Hard, port 19027)
  - Broken server ignores `version` field on PUT; stale writes silently overwrite current data
  - Detected by a 3-step scenario; step 3 (stale PUT) expects 409, broken returns 200

- **FM-018** — Idempotency and Duplicate-Operation Failures (Hard, port 19018)
  - Broken server processes every POST as a new payment; retry creates a duplicate charge
  - Detected by a 3-step scenario: submit, retry, then assert `$.count == 1`

- **FM-003** — Pagination Contract Failures (Hard, port 19003)
  - Broken server hardcodes `has_more: true` on every page including the last
  - Detected by a 3-step jsonpath scenario; step 3 (last page) is the discriminator

- **FM-044** — Sensitive Data Exposure in API Responses (Medium, port 19044)
  - Broken server serialises internal `User` struct, leaking `password_hash`, `api_key`, `role`
  - Detected by `additionalProperties: false` in the OpenAPI schema + `specFile:` binding

- **FM-034** — HTTP Method and Operation Semantics Failures (Medium, port 19034)
  - Broken server returns `200 OK` for `POST` (should be `201 Created`) and `200 OK` with body for `DELETE` (should be `204 No Content`)
  - Detected by `specFile:` binding: both status codes are absent from the declared responses

- **FM-022** — Syntactically Valid but Semantically Invalid Payloads (Medium, port 19022)
  - Broken server sets `total` to `unit_price` instead of `unit_price × quantity`
  - Detected by `jsonpath` value assertions on `$.total`

- **FM-016** — Rate-Limit and Throttling Failures (Medium, port 19016)
  - Broken server returns undeclared `429 Too Many Requests` after 2 requests
  - Detected by `specFile:` binding: `429` is not declared in the spec

- **FM-004** — Renamed or Removed Response Fields (Medium, diff-based)
  - `v2/openapi.yaml` renames `full_name` → `name` and removes `phone` (2 breaking changes)
  - Detected by `clrsig diff v1 v2 --breaking-only`

- **FM-010** — OpenAPI Specification and Implementation Drift (Easy, port 19010)
  - Broken server emits `product_type`/`available`; spec declares `category`/`in_stock`
  - Detected by `specFile:` binding + `type: schema`

- **FM-005** — Incorrect Optional, Nullable, and Required Field Behavior (Easy, port 19005)
  - Broken server returns `"email": null` for OAuth users; spec declares `email` as required string
  - Detected by inline schema assertion

- **FM-001** — Currency, Number, and Precision Formatting Failures (Easy, port 19001)
  - Broken server returns monetary values as strings (`"24.00"`); spec declares `number`
  - Detected by inline schema assertion

- **FM-025** — Incorrect HTTP Status and Business-Error Semantics (Easy, port 19025)
  - Broken server returns `200 OK` for missing products; spec declares `404`
  - Detected by `type: status` assertion

- Two CI workflows: `catalog.yml` (push + nightly, pinned clrsig release) and `catalog-main.yml` (nightly, builds clrsig from `api-workbench@main`)
- Root `README.md` with failure mode index and CI badges
- `CONTRIBUTING.md` with sub-project template and port convention
