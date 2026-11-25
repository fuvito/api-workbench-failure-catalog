# fm-023-01 — Broken Object-Level Authorization

> Full catalog entry: https://clrsig.com/failures/fm-023/

**Difficulty:** Easy
**Detected by:** `clrsig run` with two users and explicit `type: status` assertions
**Port:** 19051

**Primary failure mode:** [FM-023 — Broken Object-Level Authorization and Tenant Isolation](https://clrsig.com/failures/fm-023/)  

## What this failure is

The server checks that the caller is authenticated, but not that the caller owns the invoice named in the URL. Alice (`tenant-a`) can change `/invoices/inv-1001` to `/invoices/inv-1002` and read, update or delete Bob's (`tenant-b`) invoice. This is also called BOLA, or insecure direct object reference (IDOR).

This pattern appears when:
- Authentication is handled once in middleware, and each handler is assumed to be covered by it
- Records are looked up by primary key alone, without the tenant in the query
- The ownership check is added to the read path but forgotten on update or delete

The consequence: any customer with a valid login can reach other customers' data by supplying their IDs.

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-023-01-object-level-authz/broken/main.go &
clrsig run failures/fm-023-01-object-level-authz/scenarios/detect.yaml \
  --env failures/fm-023-01-object-level-authz/environments/local.yaml
```

Expected: **clrsig exits 1** — both control steps pass and the other five fail. Alice reads Bob's invoice (`200`), changes its memo and deletes it (`204`).

Stop the server:

```bash
# macOS / Linux
fuser -k 19051/tcp

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19051 -State Listen).OwningProcess
```

## Why clrsig catches it

The `specFile:` binding alone can't detect this: Bob's invoice is a valid `Invoice`, and `200` is a declared status. A spec says what a response looks like, not who may receive it. The scenario states that rule itself:

- **Two users.** Each step chooses whose token to send, so Alice's token can be paired with Bob's invoice ID.
- **Control steps.** Each owner reads their own invoice first. Without these, a server that returns `404` to everyone would pass.
- **State checks.** After the update and delete attempts, Bob reads the invoice again to confirm it was not changed or removed. A status code alone doesn't prove that nothing happened.

To use it on your own API, create two test users in separate tenants and set their tokens and one resource ID each in `environments/local.yaml`. Then change the URLs, request body and expected values in the scenario. The broken-server run deletes the target record, so only point this at disposable test data.

## Fix

```bash
go run failures/fm-023-01-object-level-authz/fixed/main.go &
clrsig run failures/fm-023-01-object-level-authz/scenarios/detect.yaml \
  --env failures/fm-023-01-object-level-authz/environments/local.yaml
# exits 0
```

**What changed:** `fixed/main.go` — the read, update and delete handlers all look up invoices through `findOwned`, which requires the invoice's tenant to match the caller's. Another tenant's invoice returns the same `404` as a missing one, so the response does not reveal that the ID exists. If your API documents `403` for this case instead, change the expected status in the scenario and declare `403` in the spec.

## GitHub Actions CI setup

```yaml
- { dir: "fm-023-01-object-level-authz", port: 19051, kind: "server" }
```

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-023-01-object-level-authz\broken\
$server = Start-Process "$env:TEMP\server-broken.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-023-01-object-level-authz/scenarios/detect.yaml `
  --env failures/fm-023-01-object-level-authz/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id $server.Id

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-023-01-object-level-authz\fixed\
$server = Start-Process "$env:TEMP\server-fixed.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-023-01-object-level-authz/scenarios/detect.yaml `
  --env failures/fm-023-01-object-level-authz/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
Stop-Process -Id $server.Id
```

## In the wild

This failure appears in:

- **Multi-tenant SaaS APIs** where invoices, orders or documents are fetched by ID without a tenant filter
- **Nested and secondary routes** such as exports or attachments that skip the check the main route performs
- **Mobile and single-page app backends** that rely on the client only showing a user their own IDs
- **The OWASP API Security Top 10**, which lists Broken Object Level Authorization as `API1:2023`
