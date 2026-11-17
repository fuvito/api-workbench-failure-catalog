# FM-046 — Write Returns 200 But Doesn't Persist

**Difficulty:** Medium  
**Port:** 19046  
**Detection method:** jsonpath assertion (write-then-read)

## What the broken server does

The broken server decodes the PUT body into a local variable but never writes it to the shared config store. It returns `200 OK` with the stale config, making the write appear successful while silently discarding the change.

## Why it matters

A no-op write that returns 200 is one of the hardest bugs to notice — the client believes the update succeeded and may not re-verify. It commonly arises when a handler is stubbed out and the persistence call is forgotten or accidentally guarded behind a dead branch.

## How the scenario detects it

PUT `/config` with `{"theme":"dark"}` asserts 200 (passes on both). The follow-up GET asserts `$.theme == "dark"`. The broken server returns `"light"`, failing the assertion.
