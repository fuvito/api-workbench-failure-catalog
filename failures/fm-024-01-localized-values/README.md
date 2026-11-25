# fm-024-01 — Localized Values

> Full catalog entry: https://clrsig.com/failures/fm-024/

**Difficulty:** Easy
**Detected by:** `clrsig run` with `Accept-Language` variations, `type: jsonpath` assertions and spec enums
**Port:** 19061

**Primary failure mode:** [FM-024 — Localization and Machine-Readable Data Failures](https://clrsig.com/failures/fm-024/)  

## What this failure is

With `Accept-Language: fr`, the broken server translates or reformats values that programs read:

| Field | English | French (broken) |
|---|---|---|
| `status` | `in_progress` | `en_cours` |
| `total` | `1234.5` | `"1 234,50"` |
| `created_at` | `2026-03-04T09:30:00Z` | `04/03/2026 09:30` |
| error `code` | `ORDER_NOT_FOUND` | `COMMANDE_INTROUVABLE` |

It also looks only at the start of the header, and sends no `Content-Language` or `Vary`.

This pattern appears when:
- A translation layer runs over the whole response
- Numbers and dates are formatted on the server for display
- Error codes are generated from the error message

The consequence: a client that works in English breaks for French users.

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-024-01-localized-values/broken/main.go &
clrsig run failures/fm-024-01-localized-values/scenarios/detect.yaml \
  --env failures/fm-024-01-localized-values/environments/local.yaml
```

Expected: **clrsig exits 1** — the English control step passes and the other four fail.

Stop the server:

```bash
# macOS / Linux
fuser -k 19061/tcp

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19061 -State Listen).OwningProcess
```

## Why clrsig catches it

- **Same values, different language.** The French steps expect exactly the same `status`, `total`, `created_at` and `code` as English.
- **Translation still has to happen.** The label and message must be French, so ignoring `Accept-Language` also fails.
- **Spec check.** The `status` and `code` enums, `type: number` and `format: date-time` make the `specFile:` check report translated values too.
- **Fallback and preference.** An unsupported language must fall back to English, and a supported language listed after an unsupported one must be used.

To use it on your own API, change the paths, languages, expected labels and the fields that must stay stable.

## Fix

```bash
go run failures/fm-024-01-localized-values/fixed/main.go &
clrsig run failures/fm-024-01-localized-values/scenarios/detect.yaml \
  --env failures/fm-024-01-localized-values/environments/local.yaml
# exits 0
```

**What changed:** `fixed/main.go` — only `status_label` and the error `message` come from the translation table. `language` reads every entry and its quality value, matches on the primary subtag and falls back to English. Responses send `Content-Language` and `Vary: Accept-Language`. The parser is deliberately small and doesn't handle `*`.

## GitHub Actions CI setup

```yaml
- { dir: "fm-024-01-localized-values", port: 19061, kind: "server" }
```

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-024-01-localized-values\broken\
$server = Start-Process "$env:TEMP\server-broken.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-024-01-localized-values/scenarios/detect.yaml `
  --env failures/fm-024-01-localized-values/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id $server.Id

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-024-01-localized-values\fixed\
$server = Start-Process "$env:TEMP\server-fixed.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-024-01-localized-values/scenarios/detect.yaml `
  --env failures/fm-024-01-localized-values/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
Stop-Process -Id $server.Id
```

## In the wild

This failure appears in:

- **Status values translated for the web page**, then used by mobile or partner code
- **Amounts returned with a decimal comma** for some locales
- **Error codes built from translated messages**, so client error handling only works in one language
