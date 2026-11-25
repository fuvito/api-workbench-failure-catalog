# fm-029-01 — Bulk Partial Success

> Full catalog entry: https://clrsig.com/failures/fm-029/

**Difficulty:** Medium
**Detected by:** `clrsig run` with `specFile:` binding and `type: jsonpath` assertions
**Port:** 19053

**Primary failure mode:** [FM-029 — Bulk API Partial-Success Failures](https://clrsig.com/failures/fm-029/)  

## What this failure is

A batch of payments is processed item by item, so some can succeed while others fail. The server returns `200` with only `{"succeeded":2,"failed":1}`. The client can't tell which payment failed. If it sends the whole batch again, the two successful payments are made a second time.

This pattern appears when:
- The batch handler loops over the items and keeps only counters
- Item errors are logged on the server but not returned
- Retrying was designed for single requests, so nothing stops a repeated item

The consequence: the client either leaves the failed item unpaid or pays the others twice.

Partial success is not the bug here. This API documents that items are independent. If your API promises that a batch is all-or-nothing, see [fm-020-01-batch-atomicity](../fm-020-01-batch-atomicity/) instead.

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-029-01-bulk-partial-success/broken/main.go &
clrsig run failures/fm-029-01-bulk-partial-success/scenarios/detect.yaml \
  --env failures/fm-029-01-bulk-partial-success/environments/local.yaml
```

Expected: **clrsig exits 1** — the three batch responses have no `results`, and the payment count is `4` after the batch is sent again, instead of `2`. By the last step it is `5` instead of `3`.

Stop the server:

```bash
# macOS / Linux
fuser -k 19053/tcp

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19053 -State Listen).OwningProcess
```

## Why clrsig catches it

There are two failures, and they are caught in different ways:

- **Missing item results.** The spec marks `results` as required, so the `specFile:` binding fails every batch response without it (`CS-RESP-005`).
- **Wrong correlation.** A schema can't say that the second result belongs to the second item. The scenario checks the `ref`, `status`, error code and `retryable` flag of each position.
- **Repeated payments.** No response check can see these. The scenario sends the same batch again, then counts the payments.
- **Safe recovery.** The corrected item is sent on its own, and the count shows that all three items are paid once.

The payment counts start at zero, so restart the server before running the scenario again.

To use it on your own API, change the URLs, item bodies and result paths to match your batch format. Include at least one valid and one invalid item, and count the side effects through whatever endpoint shows them. The scenario makes real writes, so only point it at disposable test data.

## Fix

```bash
go run failures/fm-029-01-bulk-partial-success/fixed/main.go &
clrsig run failures/fm-029-01-bulk-partial-success/scenarios/detect.yaml \
  --env failures/fm-029-01-bulk-partial-success/environments/local.yaml
# exits 0
```

**What changed:** `fixed/main.go`:

- Returns one result per item, in request order, with the item's `ref`. A failed result has an error `code`, a `message` and `retryable`. `retryable: false` means the same item will fail again until it is changed.
- Pays each `ref` at most once. Sending a paid item again returns the original `payment_id`. The same `ref` with a different amount fails with `ref_conflict`.
- Rejects a batch with `400`, before paying anything, if any item has no `ref`.

This demo keeps payments in memory. A real service needs to store each `ref` durably, with a unique constraint, in the same transaction as the payment.

The fixed server returns `200` for a mixed batch. Some APIs use `207 Multi-Status` instead; if yours does, change the expected status in the scenario and the spec.

## GitHub Actions CI setup

```yaml
- { dir: "fm-029-01-bulk-partial-success", port: 19053, kind: "server" }
```

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-029-01-bulk-partial-success\broken\
$server = Start-Process "$env:TEMP\server-broken.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-029-01-bulk-partial-success/scenarios/detect.yaml `
  --env failures/fm-029-01-bulk-partial-success/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id $server.Id

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-029-01-bulk-partial-success\fixed\
$server = Start-Process "$env:TEMP\server-fixed.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-029-01-bulk-partial-success/scenarios/detect.yaml `
  --env failures/fm-029-01-bulk-partial-success/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
Stop-Process -Id $server.Id
```

## In the wild

This failure appears in:

- **Payout and invoicing APIs**, where resending a whole batch repeats money movements
- **Import endpoints** that report "95 of 100 imported" without saying which 5 failed
- **Messaging and notification batches**, where recipients get the same message twice after a retry
