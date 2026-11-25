# fm-014-01 — Daylight-Saving Schedule

> Full catalog entry: https://clrsig.com/failures/fm-014/

**Difficulty:** Medium
**Detected by:** `clrsig run` with fixed dates around clock changes and exact `type: jsonpath` assertions
**Port:** 19080

**Primary failure mode:** [FM-014 — Timezone and Daylight-Saving-Time Failures](https://clrsig.com/failures/fm-014/)  

## What this failure is

A daily schedule should run at the same local time. The broken server converts the first run to an instant and adds 24 hours for each following day. After a clock change, a 09:00 job in New York runs at 10:00, and one in Berlin at 08:00. It also gives a time that is skipped by the clock change a run **before** the change.

This pattern appears when:
- Schedules are stored as UTC instants instead of a local time plus a time zone
- "Next day" means "24 hours later"
- Nobody decided what happens to repeated or skipped times

The consequence: jobs run an hour early or late for months, and some near the change run at the wrong moment.

## Prerequisites

- **Go 1.22+** — verify with `go version`
- **clrsig** — install from [clrsig.com/docs/install](https://clrsig.com/docs/install/)

## Reproduce it

Run from the **repo root**:

```bash
go run failures/fm-014-01-dst-schedule/broken/main.go &
clrsig run failures/fm-014-01-dst-schedule/scenarios/detect.yaml \
  --env failures/fm-014-01-dst-schedule/environments/local.yaml
```

Expected: **clrsig exits 1** — the ordinary-week control step passes and the other four fail.

Stop the server:

```bash
# macOS / Linux
fuser -k 19080/tcp

# Windows (PowerShell)
Stop-Process -Id (Get-NetTCPConnection -LocalPort 19080 -State Listen).OwningProcess
```

## Why clrsig catches it

Every timestamp the broken server returns is a valid `date-time`, so the spec check passes it. The scenario expects exact local times and offsets:

- **Fixed dates.** The requests name 2026 and 2027 dates, so the test gives the same result on any day.
- **Two directions, two regions.** US and European clocks change on different dates.
- **Edge policy.** The repeated and skipped times follow the policy in the spec.

The expected values depend on the time zone rules. If a government changes them, the data and expectations change. The demo embeds Go's `time/tzdata` as a fallback for machines without time zone files; a system's own time zone data takes precedence when present.

To use it on your own API, you need an endpoint that shows computed run times, such as a preview. Change the paths, zones, dates and expected values to match your documented policy for repeated and skipped times.

**What this doesn't cover:** it checks computed times, not that a scheduler actually fires then, or runs only once.

## Fix

```bash
go run failures/fm-014-01-dst-schedule/fixed/main.go &
clrsig run failures/fm-014-01-dst-schedule/scenarios/detect.yaml \
  --env failures/fm-014-01-dst-schedule/environments/local.yaml
# exits 0
```

**What changed:** `fixed/main.go` — each run is computed from its calendar date and local time. `resolve` picks the first occurrence of a repeated time and moves a skipped time past the change.

## GitHub Actions CI setup

```yaml
- { dir: "fm-014-01-dst-schedule", port: 19080, kind: "server" }
```

### Verify locally — Windows (PowerShell)

```powershell
# Broken — expect exit 1
go build -o "$env:TEMP\server-broken.exe" .\failures\fm-014-01-dst-schedule\broken\
$server = Start-Process "$env:TEMP\server-broken.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-014-01-dst-schedule/scenarios/detect.yaml `
  --env failures/fm-014-01-dst-schedule/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 1
Stop-Process -Id $server.Id

# Fixed — expect exit 0
go build -o "$env:TEMP\server-fixed.exe" .\failures\fm-014-01-dst-schedule\fixed\
$server = Start-Process "$env:TEMP\server-fixed.exe" -PassThru
Start-Sleep -Seconds 1
clrsig run failures/fm-014-01-dst-schedule/scenarios/detect.yaml `
  --env failures/fm-014-01-dst-schedule/environments/local.yaml
Write-Host "Exit: $LASTEXITCODE"   # expect: 0
Stop-Process -Id $server.Id
```

## In the wild

This failure appears in:

- **Daily reports and billing jobs** that drift by an hour twice a year
- **Calendar APIs** that store only UTC for recurring events
- **Jobs scheduled around 02:00** that are skipped or run twice on the change day
