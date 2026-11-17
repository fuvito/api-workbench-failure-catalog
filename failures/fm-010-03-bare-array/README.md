# fm-010-03 — Bare Array Instead of Object Wrapper

**Difficulty:** Easy  
**Port:** 19008  
**Detection method:** `specFile:` binding

**Primary failure mode:** [FM-010 — OpenAPI Specification and Implementation Drift](https://clrsig.com/failures/fm-010/)  

## What the broken server does

The broken server returns `GET /users` as a bare JSON array `[{...},{...}]`. The OpenAPI spec declares the response as an object `{"items":[...],"count":N}`.

## Why it matters

Bare array responses cannot carry metadata (total count, pagination cursors, links) without breaking clients. Wrapping in an object is a standard extensibility practice, and switching from array to object later is a breaking change.

## How the scenario detects it

The `specFile:` binding validates the response against the declared schema, which requires `type: object` with `required: [items, count]`. A bare array fails the `type: object` check.
