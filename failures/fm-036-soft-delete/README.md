# FM-036 — Soft Delete Becomes Hard Delete

**Difficulty:** Easy  
**Port:** 19036  
**Detection method:** status assertion

## What the broken server does

The broken server implements DELETE as a hard delete: it removes the item from the store and returns `204 No Content`. The spec requires a soft delete: return `200 OK` with the item body including a `deleted_at` timestamp, leaving the record queryable.

## Why it matters

Soft delete is a common audit-trail pattern. Accidentally implementing hard delete permanently destroys records and breaks downstream reads that expect the item to remain accessible after deletion.

## How the scenario detects it

1. DELETE `/items/1` — asserts status 200; the broken server returns 204 → detected
2. GET `/items/1` — asserts status 200; the broken server returns 404 (item was removed) → also detected
