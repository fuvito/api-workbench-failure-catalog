# fm-034-03 — Non-Idempotent GET

**Difficulty:** Easy  
**Port:** 19033  
**Detection method:** jsonpath assertion (two consecutive GETs)

**Primary failure mode:** [FM-034 — HTTP Method and Operation Semantics Failures](https://clrsig.com/failures/fm-034/)  

## What the broken server does

The broken server mutates ticket state on GET — it marks the ticket `"read"` after the first access. The first call returns the original `"unread"` state; the second reveals the mutation.

## Why it matters

GET is defined as a safe, idempotent method. A GET that changes server state breaks caches, proxies, and clients that re-fetch to confirm state. This pattern is an HTTP semantics violation.

## How the scenario detects it

Two consecutive GET `/tickets/1` calls both assert `$.status == "unread"`. The broken server passes the first assertion but mutates state before returning, so the second GET returns `"read"` and fails.
