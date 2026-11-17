# FM-017 — Stale Read After Write

**Difficulty:** Medium  
**Port:** 19017  
**Detection method:** jsonpath assertion

## What the broken server does

The broken server caches GET responses in memory but never invalidates the cache on a PUT. After a client updates an item, subsequent GETs still return the pre-update state from the cache.

## Why it matters

Stale cache bugs produce correct-looking responses that silently diverge from the actual data. A client updating a resource and immediately reading it back cannot trust the response.

## How the scenario detects it

1. PUT `/items/1` with `{"name":"Updated"}` — succeeds on both servers
2. GET `/items/1` — asserts `$.name == "Updated"`; the broken server returns `"Original"` from its stale cache
