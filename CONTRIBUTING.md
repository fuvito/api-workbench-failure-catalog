# Contributing

Thanks for helping grow the clrsig failure catalog. This document covers how to
add a new failure-mode sub-project. For the full design, see
[`docs/PRD.md`](docs/PRD.md).

## What a sub-project is

Each directory under `failures/` is a self-contained reproduction of one API
contract failure mode from the [clrsig failure-mode
catalog](https://clrsig.com/failures/). It exists to prove that `clrsig` catches
the failure in CI and to give users a scenario they can copy into their own
workspace.

## Naming and ports

- Directory name: `failures/fm-NNN-slug/` where `NNN` is the zero-padded
  failure-mode number and `slug` is a short kebab-case description
  (e.g. `fm-001-number-precision`).
- Port: every server-based sub-project listens on `19000 + NNN` (FM-001 →
  `19001`, FM-044 → `19044`). Keep `environments/local.yaml` static — do not
  add port discovery.

## Layouts

### Server-based (most failure modes)

```
failures/fm-NNN-slug/
├── README.md           # follows the template in docs/PRD.md §6.3
├── broken/
│   └── main.go         # server that exhibits the failure
├── fixed/
│   └── main.go         # corrected server — same spec, all assertions pass
├── openapi.yaml        # the declared spec
├── scenarios/
│   └── detect.yaml     # clrsig scenario that catches the failure
└── environments/
    └── local.yaml      # baseUrl: http://localhost:19NNN
```

### Spec-only (diff-based failure modes)

```
failures/fm-NNN-slug/
├── README.md
├── v1/
│   └── openapi.yaml    # original spec
├── v2/
│   └── openapi.yaml    # changed spec that introduces the breaking change
└── scenarios/
    └── detect.yaml     # clrsig diff command and expected violations
```

## Hard requirements

1. **Standard library only.** Every server is a single `main.go` file with no
   external dependencies — no frameworks, no databases.
2. **`/health` endpoint.** Every server-based sub-project must expose
   `GET /health` returning HTTP 200 (empty body is fine). CI polls this
   endpoint for readiness before running `clrsig`.
3. **Exit-code contract.** `clrsig` must exit `1` against `broken/` and exit
   `0` against `fixed/`. Both legs are required — the fix is what proves the
   detection fired for the right reason.
4. **README template.** Use the template in [`docs/PRD.md`
   §6.3](docs/PRD.md). The reproduction commands in the README are executed
   literally in CI, so they must work as written.
5. **Register the matrix entry.** Add your sub-project to the `fm:` matrix in
   `.github/workflows/catalog.yml` and `.github/workflows/catalog-main.yml`
   with the correct `kind` (`server` or `diff`).
6. **Link the catalog page.** The README header links to
   `https://clrsig.com/failures/fm-NNN/`.

## Proposing a new failure mode

The v1 scope is fixed at 12 sub-projects (see `docs/PRD.md` §6.5). To propose
one of the backlog modes (`docs/PRD.md` §8), open a pull request with the
sub-project implemented per this guide. Modes rated **Ready** in the backlog
are the best candidates.

## Verify locally

```bash
go run failures/fm-NNN-slug/broken/main.go &
clrsig run failures/fm-NNN-slug/scenarios/detect.yaml \
  --env failures/fm-NNN-slug/environments/local.yaml
# expect: exit 1 (failure caught)

go run failures/fm-NNN-slug/fixed/main.go &
clrsig run failures/fm-NNN-slug/scenarios/detect.yaml \
  --env failures/fm-NNN-slug/environments/local.yaml
# expect: exit 0
```
