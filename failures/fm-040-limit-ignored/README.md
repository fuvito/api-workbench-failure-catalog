# FM-040 — Limit Parameter Ignored

**Difficulty:** Easy  
**Port:** 19040  
**Detection method:** jsonpath assertion

## What the broken server does

The broken server parses the `limit` query parameter but never applies it — the full dataset is always returned. `GET /items?limit=2` returns all 5 items with `count: 5`.

## Why it matters

An ignored `limit` parameter silently breaks pagination, overwhelms clients expecting bounded responses, and defeats the contract clients rely on to control response size. This commonly occurs when the query param is added to the handler signature but not wired to the slice operation.

## How the scenario detects it

`GET /items?limit=2` asserts `$.count == 2`. The broken server returns `count: 5`, failing the assertion.
