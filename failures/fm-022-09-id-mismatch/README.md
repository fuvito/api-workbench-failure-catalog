# fm-022-09 — Response ID Mismatch

**Difficulty:** Easy  
**Port:** 19013  
**Detection method:** jsonpath assertion

**Primary failure mode:** [FM-022 — Syntactically Valid but Semantically Invalid Payloads](https://clrsig.com/failures/fm-022/)  

## What the broken server does

The broken server ignores the path ID parameter and always returns the same hardcoded record. `GET /products/5` returns a product with `id: 1`.

## Why it matters

An ID mismatch means the client cannot trust that the response corresponds to the requested resource. This pattern surfaces when ID parsing is added to a stub handler but the lookup is never wired up, or when query construction has a silent bug.

## How the scenario detects it

`GET /products/5` asserts `$.id == 5`. The broken server returns `id: 1`, failing the assertion.
