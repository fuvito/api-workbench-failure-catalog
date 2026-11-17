# FM-047 — Conditional Required Field Not Enforced

**Difficulty:** Medium  
**Port:** 19047  
**Detection method:** status assertion

## What the broken server does

The broken server validates the `status` enum but ignores the conditional constraint: when `status` is `"shipped"`, `tracking_number` is required. A shipment with `status:"shipped"` and no `tracking_number` returns `201 Created` instead of `422`.

## Why it matters

Conditional validation rules enforce domain-level invariants that JSON Schema alone cannot express. A shipped order without a tracking number is semantically corrupt — downstream fulfillment and notification systems will fail when they look up a tracking number that doesn't exist.

## How the scenario detects it

Two steps: a valid shipped shipment with `tracking_number` asserts 201 (passes on both), then a shipped shipment without `tracking_number` asserts 422. The broken server returns 201, failing the second assertion.
