# fm-034-02 — Unsupported Method Returns Wrong Status

**Difficulty:** Easy  
**Port:** 19014  
**Detection method:** status assertion

**Primary failure mode:** [FM-034 — HTTP Method and Operation Semantics Failures](https://clrsig.com/failures/fm-034/)  

## What the broken server does

The broken server accepts any HTTP method on `GET /settings` without a method guard. `DELETE /settings` returns `200 OK` with the settings body instead of `405 Method Not Allowed`.

## Why it matters

A missing method guard means clients can accidentally destructive-call read-only endpoints and receive a success response. It also makes the endpoint's contract ambiguous — a consumer cannot rely on the spec to know which methods are safe.

## How the scenario detects it

Two steps: a valid `GET /settings` asserts 200 (passes on both servers), then `DELETE /settings` asserts 405. The broken server returns 200 on the DELETE, failing the second assertion.
