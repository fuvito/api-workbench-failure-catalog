# Contributing

Thanks for helping grow the clrsig failure catalog. This document covers how to
add a new failure-mode sub-project. [`docs/PRD.md`](docs/PRD.md) records the
original design and isn't kept up to date; where they differ, follow this guide.

## What a sub-project is

Each directory under `failures/` is a self-contained reproduction of one API
contract failure mode from the [clrsig failure-mode
catalog](https://clrsig.com/failures/). It exists to prove that `clrsig` catches
the failure in CI and to give users a scenario they can copy into their own
workspace.

## Naming and ports

- Directory name: `failures/fm-NNN-EE-slug/` where `NNN` is the zero-padded
  failure-mode number, `EE` numbers the examples of that failure mode from `01`,
  and `slug` is a short kebab-case description
  (e.g. `fm-001-01-number-precision`).
- Port: give each server-based sub-project an unused port in the `190xx` range.
  The port isn't derived from the failure-mode number. Some sub-projects use
  more than one port, and the CI matrix lists only the one it polls on
  `/health`, so search the server code too. For example, this finds whether
  `19090` is taken:

  ```bash
  git grep -n 19090 -- failures .github/workflows
  ```

  No output means no tracked file uses it. Keep `environments/local.yaml`
  static — do not add port discovery.

## Layouts

### Server-based (most failure modes)

```
failures/fm-NNN-EE-slug/
├── README.md           # see "README template" below
├── broken/
│   └── main.go         # server that exhibits the failure
├── fixed/
│   └── main.go         # corrected server — same spec, all assertions pass
├── openapi.yaml        # the declared spec
├── scenarios/
│   └── detect.yaml     # clrsig scenario that catches the failure
└── environments/
    └── local.yaml      # baseUrl: http://localhost:190xx
```

### Spec-only (diff-based failure modes)

```
failures/fm-NNN-EE-slug/
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
4. **README template.** Follow the [README template](#readme-template) below.
   CI doesn't run the commands in your README: the workflows have their own build
   and `clrsig` steps. Run the README commands yourself and make sure they work
   as written.
5. **Register the matrix entry.** Add your sub-project to the `fm:` matrix in
   `.github/workflows/catalog.yml` and `.github/workflows/catalog-main.yml`
   with the correct `kind` (`server` or `diff`).
6. **Link the catalog page.** The README header links to
   `https://clrsig.com/failures/fm-NNN/`.

## README template

[`fm-035-01-request-headers`](failures/fm-035-01-request-headers/README.md) is a
complete example. Some older READMEs use other layouts; use this one for new
sub-projects.

The header has the title (`# fm-NNN-EE — Name`), the catalog link, **Difficulty**,
**Detected by**, **Port** and **Primary failure mode**. Then these sections:

| Section | Contents |
|---|---|
| What this failure is | What breaks, why it happens in practice and what a consumer experiences |
| Prerequisites | Go version and clrsig |
| Reproduce it | Commands to run from the repo root, the expected result and how to stop the server |
| Why clrsig catches it | Which checks fail and why, **what the example doesn't cover**, and how to adapt it to a real API, with any warning about data it changes |
| Fix | Commands for the fixed server and what changed |
| GitHub Actions CI setup | The matrix entry, and how to verify locally on Windows |
| In the wild | Where this pattern appears in real projects |

Spec-only sub-projects have no server, so they leave out the server commands.
Don't paste expected output: nothing keeps it up to date.

## Proposing a new failure mode

There is no fixed limit on the number of sub-projects, and one failure mode can
have several examples. The [README](README.md#failure-mode-index) lists what is
covered, and [which failure modes have no example and
why](README.md#failure-modes-without-examples). To propose another example, or
a way to genuinely detect one of the missing modes, open a pull request with the
sub-project implemented per this guide.

## Verify locally

```bash
go run failures/fm-NNN-EE-slug/broken/main.go &
clrsig run failures/fm-NNN-EE-slug/scenarios/detect.yaml \
  --env failures/fm-NNN-EE-slug/environments/local.yaml
# expect: exit 1 (failure caught)

go run failures/fm-NNN-EE-slug/fixed/main.go &
clrsig run failures/fm-NNN-EE-slug/scenarios/detect.yaml \
  --env failures/fm-NNN-EE-slug/environments/local.yaml
# expect: exit 0
```
