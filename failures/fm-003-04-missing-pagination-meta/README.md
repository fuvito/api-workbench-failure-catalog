# fm-003-04 — Missing Pagination Metadata Fields

**Difficulty:** Easy  
**Port:** 19049  
**Detection method:** `specFile:` binding

**Primary failure mode:** [FM-003 — Pagination Contract Failures](https://clrsig.com/failures/fm-003/)  

## What the broken server does

The broken server returns `items` and `count` but omits `total_pages` and `has_more` from the response. The spec declares all four fields as `required`.

## Why it matters

Pagination metadata is the contract between server and client for navigating large datasets. Missing `has_more` or `total_pages` forces clients to guess whether more pages exist, leading to either missed data or unnecessary extra requests.

## How the scenario detects it

The `specFile:` binding validates the response against the schema, which requires `total_pages` and `has_more`. The broken server omits both, failing the `required` check.
