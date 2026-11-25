# fm-031-01 — CORS Preflight

> Full catalog entry: https://clrsig.com/failures/fm-031/

**Difficulty:** Medium
**Detected by:** `clrsig run` with `type: header` assertions on preflight and cross-origin requests
**Port:** 19054

**Primary failure mode:** [FM-031 — CORS and Browser-Based API Access Failures](https://clrsig.com/failures/fm-031/)  

## What this failure is

The API works from `curl`, Postman and server-side tests, but a web app at `https://portal.example.test` can't use it. The broken server has five CORS mistakes:

| Mistake | Effect in a browser |
|---|---|
| The preflight doesn't allow `Authorization` or `X-Request-ID` | The real request is never sent |
| The preflight doesn't allow `PUT` | Updates are blocked |
| Error responses have no CORS headers | The page can't read the `401`; it only sees a network error |
| `X-Request-ID` is not exposed | The page can't read it |
| Any origin is allowed, with credentials | Pages on other sites can read responses to requests made with the user's credentials |

It also leaves out `Vary: Origin`, so a shared cache could serve one origin's CORS headers to another.

This pattern appears when:
- CORS was set up for a simple `GET`, then the app added authentication, custom headers or new methods
- CORS headers are added after authentication, so error responses never get them
- "Echo the request's origin" is used to get around the rule that `*` can't be combined with credentials

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-031-01-cors-preflight/broken/main.go &
clrsig run failures/fm-031-01-cors-preflight/scenarios/detect.yaml \
  --env failures/fm-031-01-cors-preflight/environments/local.yaml
```

Expected: **clrsig exits 1** — the control step passes and the other five fail.

Stop the server:

```bash
# macOS / Linux
fuser -k 19054/tcp

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19054 -State Listen).OwningProcess
```

## Why clrsig catches it

CORS is enforced by the browser, so a client that isn't a browser gets a normal response. The control step shows this: it passes against the broken server. The `specFile:` binding doesn't help either, because the status codes and bodies are correct. The scenario sends what a browser would send and checks the response headers:

- **Preflight.** An `OPTIONS` request with `Origin`, `Access-Control-Request-Method` and `Access-Control-Request-Headers`, and no credentials.
- **List headers.** `matches` finds one entry in a comma-separated list, in any order. Header names ignore case; method names don't.
- **Errors.** A request with a bad token checks that the `401` still has CORS headers.
- **Other origins.** `eq ""` asserts that a header was not sent.

**What this doesn't prove:** clrsig is not a browser. It checks the headers the server sends, not whether a particular browser accepts them. It tests one URL, so check each route or gateway path whose CORS settings can differ.

CORS also isn't access control. It limits what pages in a browser can read. It does not stop other clients, and your API must still authenticate and authorize every request.

To use it on your own API, set `baseUrl`, `token`, `allowedOrigin` and `otherOrigin` in `environments/local.yaml`. Then change the URL, method and header names to match what your web app sends. If your API doesn't allow credentialed requests, remove the `Access-Control-Allow-Credentials` assertions for the allowed origin. Some servers answer a preflight with `200` instead of `204`, which browsers also accept; change the status and the spec to match.

## Fix

```bash
go run failures/fm-031-01-cors-preflight/fixed/main.go &
clrsig run failures/fm-031-01-cors-preflight/scenarios/detect.yaml \
  --env failures/fm-031-01-cors-preflight/environments/local.yaml
# exits 0
```

**What changed:** `fixed/main.go` — all CORS handling moved into `withCORS`, which wraps every route:

- Only origins in `allowedOrigins`, matched exactly, get CORS headers.
- Preflights are answered before authentication, with `PUT`, `Authorization` and `X-Request-ID` allowed.
- The headers are set before the handler runs, so error responses have them too.
- `X-Request-ID` is exposed, and every response has `Vary: Origin`.

## GitHub Actions CI setup

```yaml
- { dir: "fm-031-01-cors-preflight", port: 19054, kind: "server" }
```

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-031-01-cors-preflight\broken\
$server = Start-Process "$env:TEMP\server-broken.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-031-01-cors-preflight/scenarios/detect.yaml `
  --env failures/fm-031-01-cors-preflight/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id $server.Id

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-031-01-cors-preflight\fixed\
$server = Start-Process "$env:TEMP\server-fixed.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-031-01-cors-preflight/scenarios/detect.yaml `
  --env failures/fm-031-01-cors-preflight/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
Stop-Process -Id $server.Id
```

## In the wild

This failure appears in:

- **API gateways** that require authentication on `OPTIONS`, so every preflight fails with `401`
- **New client headers**, such as tracing or version headers, added to the web app but not to the allowed list
- **Error pages from proxies or rate limiters** that have no CORS headers, which hides the real status from the app
- **"Reflect any origin" snippets** copied to silence a CORS error during development
