# FM-050 — Out-of-Range Page Returns 500

**Difficulty:** Easy  
**Port:** 19050  
**Detection method:** status assertion

## What the broken server does

The broken server does not guard against a `start` offset beyond the dataset length. When `page=9999`, `start` far exceeds `len(allItems)` and the server explicitly returns `500 Internal Server Error` (the production equivalent is an unrecovered index-out-of-bounds panic).

## Why it matters

Clients paginating through a large dataset will eventually request a page that is beyond the last page. The correct response is `200 OK` with an empty list — not a 500 that signals infrastructure failure and may trigger alerts or retries.

## How the scenario detects it

Two steps: page 1 asserts 200 (passes on both), then `page=9999` asserts 200. The broken server returns 500, failing the second assertion.
