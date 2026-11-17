# FM-037 — Case-Insensitive Enum Accepted

**Difficulty:** Easy  
**Port:** 19037  
**Detection method:** status assertion

## What the broken server does

The broken server normalises the input with `strings.ToLower` before comparing against the enum, making validation case-insensitive. `"PENDING"` is accepted when only `"pending"` is a valid enum value.

## Why it matters

JSON Schema enum validation is exact and case-sensitive. A server that silently accepts `"PENDING"` will store that value and later return it, causing downstream consumers and schema validators to reject the data. The source of truth is the spec, not the server's lenient parser.

## How the scenario detects it

Two steps: a valid lowercase `"pending"` asserts 201 (passes on both servers), then `"PENDING"` asserts 422. The broken server returns 201 for the uppercase input, failing the second assertion.
