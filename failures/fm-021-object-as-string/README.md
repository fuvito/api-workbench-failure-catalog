# FM-021 — Nested Object Serialized as String

**Difficulty:** Easy  
**Port:** 19021  
**Detection method:** `specFile:` binding

## What the broken server does

The broken server stores the `metadata` field as a Go `string` and marshals it as a JSON string literal — `"{\"author\":\"Alice\",\"version\":1}"` — instead of a nested JSON object `{"author":"Alice","version":1}`.

## Why it matters

A client expecting an object receives a string and must double-decode it, which is undocumented, fragile, and violates the declared schema. This often happens when a database column stores JSON text and the Go struct uses `string` instead of `json.RawMessage` or a struct type.

## How the scenario detects it

The `specFile:` binding validates the GET response against the OpenAPI schema, which declares `metadata` as `type: object`. The broken server returns a `type: string` value, which fails JSON Schema type validation.
