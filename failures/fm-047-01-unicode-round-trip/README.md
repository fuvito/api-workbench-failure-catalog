# fm-047-01 — Unicode Round Trip

> Full catalog entry: https://clrsig.com/failures/fm-047/

**Difficulty:** Easy
**Detected by:** `clrsig run` with `type: jsonpath` assertions on names written and read back
**Port:** 19055

**Primary failure mode:** [FM-047 — Character Encoding and Unicode Handling Failures](https://clrsig.com/failures/fm-047/)  

## What this failure is

A contact name may have up to 20 characters. The broken server damages names before it stores them, and still returns `201`:

| Mistake | Sent | Stored |
|---|---|---|
| The limit cuts at 20 **bytes**, which can split a character | `Анна Константинопуло` | `Анна Конст�` |
| Characters above U+FFFF are replaced | `Zoë 🚀` | `Zoë ?` |
| Combining marks are removed | `Jose` + U+0301 (shown as `José`) | `Jose` |
| A name over the limit is cut instead of rejected | 21 characters | `201` with a shorter name |

This pattern appears when:
- A length limit uses the size in bytes, for a column or buffer measured in bytes
- Storage or a conversion step can't hold 4-byte characters, such as MySQL's `utf8mb3` character set (some setups reject them instead of replacing them)
- An input filter removes every combining mark to "strip accents" or to block unusual text

The consequence: customer names and addresses are changed without any error, and the original can't be recovered.

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-047-01-unicode-round-trip/broken/main.go &
clrsig run failures/fm-047-01-unicode-round-trip/scenarios/detect.yaml \
  --env failures/fm-047-01-unicode-round-trip/environments/local.yaml
```

Expected: **clrsig exits 1** — 7 steps fail and 4 pass. The ASCII and the accented Turkish name survive, so a test that only used those would miss the bug.

Stop the server:

```bash
# macOS / Linux
fuser -k 19055/tcp

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19055 -State Listen).OwningProcess
```

## Why clrsig catches it

The `specFile:` binding alone doesn't detect this: a damaged name is still a string within `maxLength`. The scenario compares each name with exactly what it sent:

- **Create and read back.** Each name is checked in the create response and again in a separate read, because an API can echo the request correctly and still store something else.
- **One cause per name.** The Cyrillic name is within the character limit but takes 39 bytes. The emoji name is short in bytes, so it fails only because of the 4-byte character.
- **Exact comparison.** `eq` compares the strings as they are, without normalization. `Jose` followed by U+0301 and the single character `é` (U+00E9) look the same but are not equal. The scenario writes the combining mark as `́` so it stays visible and an editor can't silently combine it.
- **Over the limit.** A 21-character name must get `400`.

**What this doesn't cover:** it checks that one API preserves text, not what happens in queues or downstream systems. It also doesn't test normalization-based comparison, such as whether two forms of the same username count as duplicates. If your API normalizes text (for example to NFC), that can be correct: set the expected value to the normalized form your API documents.

To use it on your own API, change the URLs, body and JSON paths, and set the length cases to your field's real limit. The scenario creates records, so only point it at disposable test data. Save the file as UTF-8.

## Fix

```bash
go run failures/fm-047-01-unicode-round-trip/fixed/main.go &
clrsig run failures/fm-047-01-unicode-round-trip/scenarios/detect.yaml \
  --env failures/fm-047-01-unicode-round-trip/environments/local.yaml
# exits 0
```

**What changed:** `fixed/main.go` — the name is stored as sent. The limit uses `utf8.RuneCountInString`, which counts code points as `maxLength` does, and a longer name is rejected with `400`.

A code point is not always what a reader sees as one character: `é` written with a combining mark is two, and some emoji are several. Say which one your limit counts.

## GitHub Actions CI setup

```yaml
- { dir: "fm-047-01-unicode-round-trip", port: 19055, kind: "server" }
```

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-047-01-unicode-round-trip\broken\
$server = Start-Process "$env:TEMP\server-broken.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-047-01-unicode-round-trip/scenarios/detect.yaml `
  --env failures/fm-047-01-unicode-round-trip/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id $server.Id

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-047-01-unicode-round-trip\fixed\
$server = Start-Process "$env:TEMP\server-fixed.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-047-01-unicode-round-trip/scenarios/detect.yaml `
  --env failures/fm-047-01-unicode-round-trip/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
Stop-Process -Id $server.Id
```

The Windows console may show some characters as boxes or `?` even when the data is correct. The pass or fail result is what counts.

## In the wild

This failure appears in:

- **Customer and address records** moved through a system that assumes one byte per character
- **Databases set up with `utf8mb3`**, where emoji and some CJK characters can't be stored
- **Names with accents sent in decomposed form**, which some clients and file systems produce
- **Fields with a byte-based limit** behind an API that documents the limit in characters
