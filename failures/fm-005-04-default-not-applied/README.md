# fm-005-04 — Default Value Not Applied

**Difficulty:** Easy  
**Port:** 19019  
**Detection method:** `specFile:` binding

**Primary failure mode:** [FM-005 — Incorrect Optional, Nullable, and Required Field Behavior](https://clrsig.com/failures/fm-005/)  

## What the broken server does

The broken server uses `*bool` for the `published` field. When the client omits the field, the pointer stays `nil` and marshals as JSON `null`. The spec declares `published` as a required `boolean` with a default of `false` — the server must apply that default on create.

## Why it matters

A `null` where a boolean is expected breaks type-safe clients, corrupts conditional logic (`if post.published` is now neither true nor false), and violates the declared required field contract. This commonly arises from using pointer types for optional input fields without a corresponding default-application step.

## How the scenario detects it

`POST /posts` with only `title` (omitting `published`). The `specFile:` binding validates the 201 response body against the schema: `null` does not satisfy `type: boolean`, and the missing required field also fails validation.
