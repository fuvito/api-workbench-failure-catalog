# fm-037-01 — Query Injection

> Full catalog entry: https://clrsig.com/failures/fm-037/

**Difficulty:** Medium
**Detected by:** `clrsig run` with injection test strings and `type: jsonpath` result-count assertions
**Port:** 19064

**Primary failure mode:** [FM-037 — Input Injection and Unsafe Interpretation Failures](https://clrsig.com/failures/fm-037/)  

## What this failure is

Alice may search only her own products. The broken server builds a filter expression by pasting her search text into it, then parses the result. Searching for `zzz OR owner=bob` returns Bob's product. It also sorts by any column name, including `cost`, which is never returned.

The demo uses a tiny query language in place of SQL. It can't run commands or touch anything outside the server.

This pattern appears when:
- A query is built with string concatenation or formatting
- Values are parameterized but column or sort names are not
- A search feature passes user text to a query parser

The consequence: a caller can change what a query does and read data outside their scope.

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-037-01-query-injection/broken/main.go &
clrsig run failures/fm-037-01-query-injection/scenarios/detect.yaml \
  --env failures/fm-037-01-query-injection/environments/local.yaml
```

Expected: **clrsig exits 1** — the two control steps pass and the other three fail.

Stop the server:

```bash
# macOS / Linux
fuser -k 19064/tcp

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19064 -State Listen).OwningProcess
```

## Why clrsig catches it

Leaked rows still match the response schema, so the scenario checks what the search returned:

- **Result counts.** An injection string is not a product name, so it must match nothing.
- **Legitimate punctuation.** A name with an apostrophe must still be found. Blocking characters is not a fix.
- **Identifiers.** Sorting by a column that isn't on the allowed list must get `400`.

**What this doesn't cover:** these strings fit the demo's query language. Real SQL, NoSQL and command injection need strings that fit your interpreter, and a passing run doesn't show that an API is free of injection. Use dedicated security testing for that.

To use it on your own API, use read-only test strings suited to your data store, and change the paths and expected counts. Run it only against an isolated environment you're allowed to test. Don't send strings that modify or delete data.

## Fix

```bash
go run failures/fm-037-01-query-injection/fixed/main.go &
clrsig run failures/fm-037-01-query-injection/scenarios/detect.yaml \
  --env failures/fm-037-01-query-injection/environments/local.yaml
# exits 0
```

**What changed:** `fixed/main.go` — the search text is passed as a value and never parsed, which is what a parameterized query does. The owner condition comes from the caller's identity. `sort` must be a key in `sortable`.

## GitHub Actions CI setup

```yaml
- { dir: "fm-037-01-query-injection", port: 19064, kind: "server" }
```

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-037-01-query-injection\broken\
$server = Start-Process "$env:TEMP\server-broken.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-037-01-query-injection/scenarios/detect.yaml `
  --env failures/fm-037-01-query-injection/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id $server.Id

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-037-01-query-injection\fixed\
$server = Start-Process "$env:TEMP\server-fixed.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-037-01-query-injection/scenarios/detect.yaml `
  --env failures/fm-037-01-query-injection/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
Stop-Process -Id $server.Id
```

## In the wild

This failure appears in:

- **SQL built with string formatting**, often in reporting and search endpoints
- **`ORDER BY` clauses** built from a request parameter, because placeholders don't work there
- **Filter objects** passed straight from JSON to a document database
