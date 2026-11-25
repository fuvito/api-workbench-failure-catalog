# Changelog

All notable changes to this project will be documented here.

Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
Versions map to GitHub releases; use the release tag when cutting one.

---

## [Unreleased]

### Added

- **fm-046-01** — Secret Leakage (Medium, port 19058)
  - Broken server leaks internal details through headers, error messages, an extra error field and a debug endpoint
  - Detected by exact error messages, absent-header checks, a strict error schema and an expected 404

- **fm-045-01** — Trusted Identity Headers (Medium, ports 19056 and 19057)
  - Broken gateway keeps identity headers sent by the caller, and the backend trusts them
  - Detected by forged headers sent through the gateway and straight to the backend, with four control steps

- **fm-047-01** — Unicode Round Trip (Easy, port 19055)
  - Broken server cuts names at 20 bytes, replaces 4-byte characters with "?" and removes combining marks
  - Detected by exact name assertions on create and read-back, plus a 21-character rejection check

- **fm-031-01** — CORS Preflight (Medium, port 19054)
  - Broken server leaves Authorization, X-Request-ID and PUT out of the preflight response, omits CORS headers on errors, and allows any origin with credentials
  - Detected by header assertions on preflight and cross-origin requests, with a control step without an Origin header

- **fm-029-01** — Bulk Partial Success (Medium, port 19053)
  - Broken server returns only success and failure counts, and pays successful items again when a batch is resent
  - Detected by `specFile:` binding, per-item result assertions and payment counts

- **fm-043-01** — Mass Assignment (Easy, port 19052)
  - Broken server lets a customer set role, tenant_id and the nested billing.credit_limit through PATCH
  - Detected by status assertions plus re-reads of the stored values, with a control step for editable fields

- **fm-023-01** — Broken Object-Level Authorization (Easy, port 19051)
  - Broken server lets any authenticated user read, update and delete another tenant's invoice by ID
  - Detected by status assertions with two users, plus control steps and owner re-reads

- **fm-025-06** — Out-of-Range Page Returns 500 (Easy, port 19050)
  - Broken server returns 500 when `?page=9999` exceeds the dataset; spec requires 200 with empty list
  - Detected by status assertion: out-of-range page expects 200

- **fm-003-04** — Missing Pagination Metadata Fields (Easy, port 19049)
  - Broken server omits `total_pages` and `has_more` from list responses; both are required
  - Detected by `specFile:` binding — missing required fields fail schema validation

- **fm-025-05** — Empty Result Set Returns 404 (Easy, port 19048)
  - Broken server returns 404 when a filter matches zero records instead of 200 with empty list
  - Detected by status assertion: filtered query with no matches expects 200

- **fm-022-08** — Conditional Required Field Not Enforced (Medium, port 19047)
  - Broken server accepts `status:"shipped"` without `tracking_number`; spec requires it conditionally
  - Detected by status assertion: shipped shipment without tracking_number expects 422

- **fm-038-01** — Write Returns 200 But Doesn't Persist (Medium, port 19046)
  - Broken server decodes PUT body but never writes it to the store; GET after PUT returns stale value
  - Detected by jsonpath: `$.theme == "dark"` after PUT with `{"theme":"dark"}`

- **fm-022-07** — Empty String Accepted for Required Field (Easy, port 19045)
  - Broken server checks only for missing fields; `username: ""` returns 201 instead of 422
  - Detected by status assertion: second POST with empty username expects 422

- **fm-022-05** — Limit Parameter Ignored (Easy, port 19040)
  - Broken server parses `?limit=` but never applies it; always returns the full dataset
  - Detected by jsonpath assertion: `$.count == 2` for `?limit=2` with 5 items in store

- **fm-022-06** — Case-Insensitive Enum Accepted (Easy, port 19037)
  - Broken server normalises input with `strings.ToLower`; `"PENDING"` is accepted when only `"pending"` is valid
  - Detected by status assertion: POST with uppercase enum value expects 422

- **fm-005-04** — Default Value Not Applied (Easy, port 19019)
  - Broken server uses `*bool` for `published`; omitting it in POST returns `null` instead of the declared default `false`
  - Detected by `specFile:` binding — `null` does not satisfy `type: boolean`

- **fm-010-06** — Array Item Type Mismatch (Easy, port 19015)
  - Broken server returns integer `42` in a string-typed array due to a missing type conversion
  - Detected by `specFile:` binding — integer fails `items: { type: string }` validation

- **fm-034-02** — Unsupported Method Returns Wrong Status (Easy, port 19014)
  - Broken server accepts any method on a read-only endpoint; DELETE returns 200 instead of 405
  - Detected by status assertion: DELETE expects 405

- **fm-022-09** — Response ID Mismatch (Easy, port 19013)
  - Broken server ignores path ID and always returns the first record; GET /products/5 returns id=1
  - Detected by jsonpath assertion: `$.id == 5`

- **fm-003-03** — Off-by-One in Page Offset (Easy, port 19012)
  - Broken server uses `page*perPage` instead of `(page-1)*perPage`, skipping the first page of results
  - Detected by jsonpath assertion: `$.items[0].id == 1` on page 1 and `$.items[0].id == 4` on page 2

- **fm-004-02** — Field Name Casing Error (Easy, port 19009)
  - Broken server returns camelCase fields (`userId`, `firstName`); spec declares snake_case (`user_id`, `first_name`)
  - Detected by `specFile:` binding — required snake_case fields missing; `additionalProperties: false` rejects camelCase

- **fm-010-03** — Bare Array Instead of Object Wrapper (Easy, port 19008)
  - Broken server returns `GET /users` as a bare JSON array; spec declares an object with `items` and `count`
  - Detected by `specFile:` binding — array fails `type: object` check

- **fm-010-07** — Minimum Value Constraint Violated (Easy, port 19041)
  - Broken server returns products with negative `price` and `stock` values
  - Detected by `specFile:` binding — `minimum: 0` constraint in JSON Schema validation

- **fm-034-05** — Soft Delete Becomes Hard Delete (Easy, port 19036)
  - Broken server permanently removes item on DELETE and returns 204; spec requires 200 with `deleted_at`
  - Detected by status assertion: DELETE expects 200; GET after delete expects 200

- **fm-034-03** — Non-Idempotent GET (Easy, port 19033)
  - Broken server mutates ticket state on GET — marks it "read" after first access
  - Detected by two consecutive GET assertions: both expect `$.status == "unread"`

- **fm-010-04** — Nested Object Serialized as String (Easy, port 19021)
  - Broken server returns `metadata` as a JSON-encoded string instead of a nested object
  - Detected by `specFile:` binding — `type: object` fails against actual string value

- **fm-017-01** — Stale Read After Write (Medium, port 19017)
  - Broken server caches GET responses and never invalidates the cache on PUT
  - Detected by jsonpath assertion: `$.name == "Updated"` after a PUT is ignored by stale cache

- **fm-025-04** — Validation Error Returns 500 Instead of 422 (Easy, port 19043)
  - Broken server treats missing-field validation the same as runtime errors, returning 500
  - Detected by status assertion: expects 422

- **fm-018-02** — Duplicate Create Returns 200 Instead of 409 (Easy, port 19042)
  - Broken server inserts duplicate records without checking uniqueness constraint
  - Detected by status assertion: second POST with same email expects 409

- **fm-005-03** — Nested Required Field Missing (Easy, port 19032)
  - `omitempty` on `street` field silently drops it when empty; spec declares it required
  - Detected by `specFile:` binding — missing required property inside nested Address object

- **fm-003-02** — Page Overflow Returns Wrong Data (Easy, port 19031)
  - Broken server wraps page number with modulo; page=99 returns page=1 data
  - Detected by jsonpath assertion: `$.count == 0` for out-of-range page

- **fm-025-03** — Redirect Instead of Response (Easy, port 19011)
  - Broken server returns 301 for all requests due to misconfigured redirect middleware
  - Detected by status assertion: expects 200

- **fm-034-06** — Bulk Delete No-Op (Easy, port 19039)
  - Broken server returns 204 but never removes items; GET after delete still returns 200
  - Detected by status assertion: step 3 expects 404 on deleted item

- **fm-022-02** — Filter Parameter Ignored (Easy, port 19028)
  - Broken server ignores `?status=` filter and returns all orders every time
  - Detected by jsonpath assertion on `$.count` for completed and pending filters

- **fm-028-01** — Async Job Status Never Transitions (Easy, port 19026)
  - Broken server sets job status to "processing" and never advances it to "completed"
  - Detected by jsonpath assertion `$.status == "completed"` on GET /jobs/1

- **fm-007-01** — API Version Routing (Easy, port 19024)
  - Broken server routes /v2/products to v1 handler; required `sku` field is missing
  - Detected by `specFile:` binding — missing required property in v2 schema

- **fm-020-01** — Batch Operation Atomicity (Medium, port 19020)
  - Broken server partially creates valid items in a batch containing an invalid one
  - Detected by status assertion: expects 422 (reject all), broken returns 201 (partial success)

- **fm-034-04** — PATCH Semantics Failure (Medium, port 19038)
  - Broken server replaces entire resource on PATCH instead of merging only provided fields
  - Detected by 3-step read-modify-read scenario; jsonpath assertions catch zeroed-out fields after PATCH

- **fm-022-04** — Sort Order Contract (Easy, port 19035)
  - Broken server ignores `sort` query parameter and always returns insertion order
  - Detected by jsonpath assertions on `$.items[0].price` for ascending and descending sorts

- **fm-022-03** — Enum Value Violation (Easy, port 19030)
  - Broken server returns wrong-case (`"PENDING"`) and undeclared (`"SHIPPED"`) enum values
  - Detected by `specFile:` binding — JSON Schema enum validation is exact and case-sensitive

- **fm-010-05** — Boolean Field Type Confusion (Easy, port 19029)
  - Broken server returns booleans as strings (`"true"`) or integers (`1`/`0`) instead of JSON booleans
  - Detected by `specFile:` binding — string/integer types do not satisfy `type: boolean`

- **fm-005-02** — Empty Collection vs Null (Easy, port 19023)
  - Broken server returns `null` for empty arrays; Go nil slices marshal to null instead of `[]`
  - Detected by `specFile:` binding — `null` does not satisfy `type: array` in OpenAPI 3.0

- **fm-025-02** — Authentication Error Code Confusion (Easy, port 19007)
  - Broken server returns 200 with `{"error":"..."}` instead of 401 on bad credentials and missing token
  - Detected by explicit `type: status` assertions — specFile binding alone cannot catch wrong status codes

- **fm-010-02** — Inconsistent Error Response Schema (Easy, port 19006)
  - Broken server returns different error shapes per endpoint (`message`, `code+msg`, `error`)
  - Detected by `specFile:` binding — non-conforming error bodies fail against the declared `Error` schema

- **fm-001-02** — Date/Time Format Failures (Easy, port 19002)
  - Broken server returns `created_at` as a Unix timestamp integer; spec declares `type: string, format: date-time`
  - Detected by `specFile:` binding — integer type mismatch against declared `string` fails automatically

- **fm-027-01** — Concurrent Update and Lost-Update Failures (Hard, port 19027)
  - Broken server ignores `version` field on PUT; stale writes silently overwrite current data
  - Detected by a 3-step scenario; step 3 (stale PUT) expects 409, broken returns 200

- **fm-018-01** — Idempotency and Duplicate-Operation Failures (Hard, port 19018)
  - Broken server processes every POST as a new payment; retry creates a duplicate charge
  - Detected by a 3-step scenario: submit, retry, then assert `$.count == 1`

- **fm-003-01** — Pagination Contract Failures (Hard, port 19003)
  - Broken server hardcodes `has_more: true` on every page including the last
  - Detected by a 3-step jsonpath scenario; step 3 (last page) is the discriminator

- **fm-044-01** — Sensitive Data Exposure in API Responses (Medium, port 19044)
  - Broken server serialises internal `User` struct, leaking `password_hash`, `api_key`, `role`
  - Detected by `additionalProperties: false` in the OpenAPI schema + `specFile:` binding

- **fm-034-01** — HTTP Method and Operation Semantics Failures (Medium, port 19034)
  - Broken server returns `200 OK` for `POST` (should be `201 Created`) and `200 OK` with body for `DELETE` (should be `204 No Content`)
  - Detected by `specFile:` binding: both status codes are absent from the declared responses

- **fm-022-01** — Syntactically Valid but Semantically Invalid Payloads (Medium, port 19022)
  - Broken server sets `total` to `unit_price` instead of `unit_price × quantity`
  - Detected by `jsonpath` value assertions on `$.total`

- **fm-016-01** — Rate-Limit and Throttling Failures (Medium, port 19016)
  - Broken server returns undeclared `429 Too Many Requests` after 2 requests
  - Detected by `specFile:` binding: `429` is not declared in the spec

- **fm-004-01** — Renamed or Removed Response Fields (Medium, diff-based)
  - `v2/openapi.yaml` renames `full_name` → `name` and removes `phone` (2 breaking changes)
  - Detected by `clrsig diff v1 v2 --breaking-only`

- **fm-010-01** — OpenAPI Specification and Implementation Drift (Easy, port 19010)
  - Broken server emits `product_type`/`available`; spec declares `category`/`in_stock`
  - Detected by `specFile:` binding + `type: schema`

- **fm-005-01** — Incorrect Optional, Nullable, and Required Field Behavior (Easy, port 19005)
  - Broken server returns `"email": null` for OAuth users; spec declares `email` as required string
  - Detected by inline schema assertion

- **fm-001-01** — Currency, Number, and Precision Formatting Failures (Easy, port 19001)
  - Broken server returns monetary values as strings (`"24.00"`); spec declares `number`
  - Detected by inline schema assertion

- **fm-025-01** — Incorrect HTTP Status and Business-Error Semantics (Easy, port 19025)
  - Broken server returns `200 OK` for missing products; spec declares `404`
  - Detected by `type: status` assertion

- Two CI workflows: `catalog.yml` (push + nightly, pinned clrsig release) and `catalog-main.yml` (nightly, builds clrsig from `api-workbench@main`)
- Root `README.md` with failure mode index and CI badges
- `CONTRIBUTING.md` with sub-project template and port convention
