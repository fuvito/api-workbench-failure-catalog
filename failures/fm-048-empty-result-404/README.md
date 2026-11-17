# FM-048 — Empty Result Set Returns 404

**Difficulty:** Easy  
**Port:** 19048  
**Detection method:** status assertion

## What the broken server does

The broken server conflates "no matching records" with "resource not found". A filtered query that matches zero users returns `404 Not Found` instead of `200 OK` with an empty list.

## Why it matters

A list endpoint returning 404 for empty results forces clients to treat 404 as a normal list response, breaking the semantic contract. Clients that correctly interpret 404 as "this endpoint doesn't exist" will treat the empty-results case as an error.

## How the scenario detects it

Two steps: a known role asserts 200 (passes on both), then `?role=unknown` (no matches) asserts 200. The broken server returns 404, failing the second assertion.
