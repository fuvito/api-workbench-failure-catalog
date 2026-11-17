# fm-004-02 — Field Name Casing Error

**Difficulty:** Easy  
**Port:** 19009  
**Detection method:** `specFile:` binding

**Primary failure mode:** [FM-004 — Renamed or Removed Response Fields](https://clrsig.com/failures/fm-004/)  

## What the broken server does

The broken server uses camelCase JSON field names (`userId`, `firstName`, `lastName`) instead of the snake_case names declared in the OpenAPI spec (`user_id`, `first_name`, `last_name`). The declared fields are absent and undeclared fields are present.

## Why it matters

Field name casing is a contract detail that every client must match exactly. A Go server that forgets `json:"snake_case"` struct tags (or uses a framework that defaults to camelCase) silently breaks every consumer of the API.

## How the scenario detects it

The `specFile:` binding validates the response against the schema, which declares `required: [user_id, first_name, last_name, email]` and `additionalProperties: false`. The broken server's camelCase fields fail on both: required snake_case fields are missing and undeclared camelCase fields are rejected.
