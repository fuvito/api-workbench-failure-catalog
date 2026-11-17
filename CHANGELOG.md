# Changelog

All notable changes to this project will be documented here.

Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
Versions map to GitHub releases; use the release tag when cutting one.

---

## [Unreleased]

### Added

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
