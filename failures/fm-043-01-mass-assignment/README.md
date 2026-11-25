# fm-043-01 — Mass Assignment

> Full catalog entry: https://clrsig.com/failures/fm-043/

**Difficulty:** Easy
**Detected by:** `clrsig run` with `type: status` and `type: jsonpath` assertions
**Port:** 19052

**Primary failure mode:** [FM-043 — Mass Assignment and Unauthorized Field Modification](https://clrsig.com/failures/fm-043/)  

## What this failure is

A customer may change their `name` and `billing.invoice_email`. The server decodes the `PATCH` body straight into the stored record, so the same request can also set `role`, `tenant_id` and `billing.credit_limit`. The caller is allowed to update this record, but not these properties.

This pattern appears when:
- The request body is bound directly to the database or domain model
- A field is added to the model later and becomes writable without anyone deciding it should be
- Top-level fields are filtered, but nested objects are copied as a whole

The consequence: a customer can make themselves an admin, move into another tenant or raise their own credit limit.

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-043-01-mass-assignment/broken/main.go &
clrsig run failures/fm-043-01-mass-assignment/scenarios/detect.yaml \
  --env failures/fm-043-01-mass-assignment/environments/local.yaml
```

Expected: **clrsig exits 1** — the control step passes and the other four fail. Both attacks return `200`, and the reads show `role: admin`, `tenant_id: tenant-b` and a credit limit of `1000000`.

Stop the server:

```bash
# macOS / Linux
fuser -k 19052/tcp

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19052 -State Listen).OwningProcess
```

## Why clrsig catches it

The `specFile:` binding alone can't detect this. The changed record still matches the `Customer` schema: `admin` is a valid role and `1000000` is a valid credit limit. The spec marks these fields `readOnly` and sets `additionalProperties: false` on the update body, but the server doesn't enforce either one, and a scenario run checks responses, not the request you chose to send. `clrsig conform --negative` does not currently probe unknown fields. The scenario tests the rule itself:

- **Control step.** The editable fields are updated first, so a server that rejects every update fails.
- **Mixed requests.** Each attack combines a protected field with an editable one, the way a real request might hide it.
- **Top level and nested.** `role` and `tenant_id` are top-level; `credit_limit` sits inside `billing`, next to a field the caller may edit.
- **State checks.** A read after each attack confirms the stored values. This matters even when the status is correct: a server can return `400` after it has already saved the change.

To use it on your own API, set your base URL and a test user's token in `environments/local.yaml`, then change the URLs, bodies and expected values to match your record. The broken-server run changes the record, so only point this at disposable test data.

## Fix

```bash
go run failures/fm-043-01-mass-assignment/fixed/main.go &
clrsig run failures/fm-043-01-mass-assignment/scenarios/detect.yaml \
  --env failures/fm-043-01-mass-assignment/environments/local.yaml
# exits 0
```

**What changed:** `fixed/main.go` — `PATCH` decodes into `customerUpdate`, which lists only the writable fields, with `DisallowUnknownFields`. Any other field, at any depth, rejects the whole request with `400`, so the editable fields in it are not applied either.

Some APIs silently ignore protected fields instead of rejecting them. That is also safe if the stored values don't change. For that policy, change each attack step's expected status to `200`, and in the reads expect the editable values the attack sent (`Mallory` and `mallory@example.com`). The protected values stay the same.

## GitHub Actions CI setup

```yaml
- { dir: "fm-043-01-mass-assignment", port: 19052, kind: "server" }
```

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-043-01-mass-assignment\broken\
$server = Start-Process "$env:TEMP\server-broken.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-043-01-mass-assignment/scenarios/detect.yaml `
  --env failures/fm-043-01-mass-assignment/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id $server.Id

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-043-01-mass-assignment\fixed\
$server = Start-Process "$env:TEMP\server-fixed.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-043-01-mass-assignment/scenarios/detect.yaml `
  --env failures/fm-043-01-mass-assignment/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
Stop-Process -Id $server.Id
```

## In the wild

This failure appears in:

- **Frameworks with automatic model binding**, where request fields are copied onto an entity unless each one is excluded
- **ORM-backed update endpoints** that pass the parsed body straight to the update call
- **Profile and settings APIs** where the client only shows the editable fields, so the others are assumed to be unreachable
- **The OWASP API Security Top 10**, which covers this under `API3:2023` Broken Object Property Level Authorization
