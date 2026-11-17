# fm-010-06 — Array Item Type Mismatch

**Difficulty:** Easy  
**Port:** 19015  
**Detection method:** `specFile:` binding

**Primary failure mode:** [FM-010 — OpenAPI Specification and Implementation Drift](https://clrsig.com/failures/fm-010/)  

## What the broken server does

The broken server returns a tags list where one element is an integer (`42`) instead of a string. The spec declares `items: { type: string }` for the array.

## Why it matters

Heterogeneous arrays break clients that assume uniform types — deserializers throw at runtime, type-safe languages cannot unmarshal, and processing pipelines silently corrupt data. This commonly arises when a numeric database ID leaks into a string-typed array due to a missing type conversion.

## How the scenario detects it

The `specFile:` binding validates each array item against `type: string`. The integer `42` fails the item schema, causing clrsig to exit 1.
