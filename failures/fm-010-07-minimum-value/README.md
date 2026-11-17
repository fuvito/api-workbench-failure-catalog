# fm-010-07 — Minimum Value Constraint Violated

**Difficulty:** Easy  
**Port:** 19041  
**Detection method:** `specFile:` binding

**Primary failure mode:** [FM-010 — OpenAPI Specification and Implementation Drift](https://clrsig.com/failures/fm-010/)  

## What the broken server does

The broken server stores and returns products with negative `price` and `stock` values. The OpenAPI schema declares `minimum: 0` on both fields, but no runtime validation enforces this constraint.

## Why it matters

Negative prices or inventory counts are semantically invalid and can corrupt billing, inventory, and reporting logic downstream. Schema constraints like `minimum` codify domain invariants that the server must uphold.

## How the scenario detects it

The `specFile:` binding validates the GET `/products` response against the OpenAPI schema. The broken server returns `price: -5.00` and `stock: -3`, both of which fail the `minimum: 0` constraint in JSON Schema validation.
