# FM-045 — Empty String Accepted for Required Field

**Difficulty:** Easy  
**Port:** 19045  
**Detection method:** status assertion

## What the broken server does

The broken server checks only for missing fields — it validates that `username` and `email` are both absent simultaneously, but accepts an empty string `""` as a valid value. `POST /users` with `{"username":"","email":"alice@example.com"}` returns 201 instead of 422.

## Why it matters

An empty string is semantically equivalent to no value for most string fields. Accepting it creates blank usernames or emails in the database, breaks display logic, and violates the `minLength: 1` constraint the spec declares on required string fields.

## How the scenario detects it

Two steps: a valid POST asserts 201 (passes on both), then a POST with `username: ""` asserts 422. The broken server returns 201, failing the second assertion.
