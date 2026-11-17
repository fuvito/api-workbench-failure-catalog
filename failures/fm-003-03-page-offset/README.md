# fm-003-03 — Off-by-One in Page Offset

**Difficulty:** Easy  
**Port:** 19012  
**Detection method:** jsonpath assertion

**Primary failure mode:** [FM-003 — Pagination Contract Failures](https://clrsig.com/failures/fm-003/)  

## What the broken server does

The broken server computes the slice start as `page * perPage` instead of `(page - 1) * perPage`. For page 1 this skips the first three items; for page 2 it overshoots past the end of the dataset, returning an empty list.

## Why it matters

Off-by-one pagination bugs silently drop the first page of results for every caller and return empty pages earlier than expected. They are easy to introduce and hard to notice without an explicit contract test against expected item IDs.

## How the scenario detects it

Two paginated requests assert the first item's `id` on each page. The broken server's miscalculated offset returns wrong items on page 1 and an empty array on page 2.
