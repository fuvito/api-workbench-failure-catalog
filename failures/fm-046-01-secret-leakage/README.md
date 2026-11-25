# fm-046-01 — Secret Leakage

> Full catalog entry: https://clrsig.com/failures/fm-046/

**Difficulty:** Medium
**Detected by:** `clrsig run` with `type: status`, `type: header` and `type: jsonpath` assertions, plus a strict error schema
**Port:** 19058

**Primary failure mode:** [FM-046 — Security Policy and Secret Leakage Failures](https://clrsig.com/failures/fm-046/)  

## What this failure is

Normal reads work, but errors, headers and a debug endpoint give away internal details. All the values in this demo are fake.

| Where | What the broken server sends |
|---|---|
| Every report response | `X-Backend-Host` and `X-Debug-Query` headers with the database host and SQL |
| `401` | The rejected token |
| A badly formed date | `500` with the database error and connection string, including the password |
| A server error | A safe `error` message, plus a `detail` field with the connection string |
| `GET /debug/config` | The connection string and a signing key, with no login |

This pattern appears when:
- An error value is converted to text and put straight into the response
- Debug headers or endpoints are switched on for troubleshooting and left on
- Input reaches the database before it is validated

The consequence: a caller who triggers an error can learn how to reach internal systems, or get their credentials.

For extra fields in **successful** responses, see [fm-044-01-sensitive-data](../fm-044-01-sensitive-data/).

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-046-01-secret-leakage/broken/main.go &
clrsig run failures/fm-046-01-secret-leakage/scenarios/detect.yaml \
  --env failures/fm-046-01-secret-leakage/environments/local.yaml
```

Expected: **clrsig exits 1** — the control step passes and the other five fail.

Stop the server:

```bash
# macOS / Linux
fuser -k 19058/tcp

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19058 -State Listen).OwningProcess
```

## Why clrsig catches it

clrsig has no "does not contain" check, and you often can't know what a leaked secret will look like. The scenario states what a safe response is instead:

- **Exact error messages.** `eq` on `$.error` fails on anything added to the message.
- **Strict error schema.** `additionalProperties: false` fails on any extra field. The server-error step needs this: its status and message are correct, and only the spec check finds `detail` (`CS-RESP-010`). Without that line in the spec, the step would pass.
- **Absent headers.** `eq ""` asserts that a header was not sent. This checks only the headers you name.
- **Debug paths.** The scenario expects `404` from a path that must not exist. It isn't in the spec, so clrsig reports it as not compared with the spec. List your own debug and admin paths.

**What this doesn't cover:** clrsig only sees responses. It can't check logs, traces or other telemetry, which are part of FM-046 too. The scenario also checks only the errors it triggers. In the fixed server, the log line holds no connection string or token.

### Keep secrets out of your test output

When a check fails, clrsig prints what the server sent. Running this scenario against the broken server prints the fake database password in the output, and in CI that goes into the build log.

- Mark your own credentials `secret: true` in the environment file. clrsig then replaces them with `***`, including when the server echoes them back in a response (checked with text, JSON and JUnit output).
- clrsig can't mask a secret it doesn't know about, such as the server's connection string. Run this against a test environment with fake secrets, and limit who can read the output. If a real secret appears, rotate it.

To use it on your own API, set `baseUrl`, `token`, `reportId` and `failingReportId` in `environments/local.yaml`, and change the paths, expected messages, header names and debug paths. You need a safe way to cause a server error, such as a test record or a test-only fault switch. If you don't have one, remove that step. Test only systems you're allowed to test.

## Fix

```bash
go run failures/fm-046-01-secret-leakage/fixed/main.go &
clrsig run failures/fm-046-01-secret-leakage/scenarios/detect.yaml \
  --env failures/fm-046-01-secret-leakage/environments/local.yaml
# exits 0
```

**What changed:** `fixed/main.go`:

- Error responses are built by `writeError`, which is only given fixed messages, never an error value or request data.
- The date is validated before the lookup, so bad input gets `400`.
- The cause of a server error is logged, not returned.
- The diagnostic headers and `/debug/config` are removed.

The demo's database error has no connection details. Real drivers may include them, so remove secrets from errors before you log them.

## GitHub Actions CI setup

```yaml
- { dir: "fm-046-01-secret-leakage", port: 19058, kind: "server" }
```

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-046-01-secret-leakage\broken\
$server = Start-Process "$env:TEMP\server-broken.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-046-01-secret-leakage/scenarios/detect.yaml `
  --env failures/fm-046-01-secret-leakage/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id $server.Id

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-046-01-secret-leakage\fixed\
$server = Start-Process "$env:TEMP\server-fixed.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-046-01-secret-leakage/scenarios/detect.yaml `
  --env failures/fm-046-01-secret-leakage/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
Stop-Process -Id $server.Id
```

## In the wild

This failure appears in:

- **Framework debug modes** left on outside development, which return stack traces and settings
- **Error handlers that return the exception message**, including database and HTTP client errors with URLs and credentials
- **Management and diagnostic endpoints** exposed without authentication
- **Troubleshooting headers** that name internal hosts, versions or queries
