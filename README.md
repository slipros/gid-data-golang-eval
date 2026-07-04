# gid-data-golang-eval

A custom golangci-lint plugin that turns the internal style guide
(skill `go-styleguide`) into a deterministic linter for **local development**.

## Status: the linter fully replaces the style part of the `go-styleguide` skill

Verified on 2026-06-07 by a full cross-check of all skill docs (31 files) against the registry:

- every deterministically checkable style-guide rule is implemented (GID-001…GID-234,
  all ✅ with a mandatory eval) or covered by standard golangci-lint linters
  (layer 3, GID-201…GID-209);
- heuristics that cannot be ported are explicitly listed in [RULES.md](RULES.md)
  ("Not portable") and [FINDINGS.md](FINDINGS.md) §2.4/2.5 — they deliberately stay
  on code review;
- on top of the skill, Uber/Google best-practice rules were added (GID-178…GID-197)
  that the skill never checked — the resulting style control is stricter than a
  manual skill-based review.
- `make eval` — all analysistest suites green; `make lint-fast` on this repository —
  0 issues.

The skill remains the source of code templates and the project documentation format
(task specs, README indexes); it no longer performs code style checking —
the linter does that deterministically.

- **[RULES.md](RULES.md)** — rule registry with statuses; every rule must have an eval
- `analyzers/` — go/analysis analyzers (one rule or a group of related GID-IDs = one linter)
- `analyzers/patterns/` — simple AST pattern rules (GID-001…008), layer 1
- `.golangci.yml` — reference config with all linters and settings examples;
  based on the production config of consent-api (UDMP/backend-go) with GID layers on top

## Quick start

Requires golangci-lint **v2.12.2** (pinned in `.custom-gcl.yml`).

```sh
make build         # build the bin/custom-gcl binary
make eval          # run evals for all rules (go test ./...)
make lint-fast     # lint this repository with the built binary
make install-hook  # git pre-commit hook with the local check
```

## Using it in your service

`gid*` linters are golangci-lint module plugins: a regular `golangci-lint run`
does **not** see them — they are compiled into a separate `custom-gcl` binary
(full golangci-lint v2.12.2 + our linters). You use the built binary exactly like
regular golangci-lint — standard and `gid*` linters run in a single pass over a
single `.golangci.yml`. Build the binary in one of two ways.

### Option A — `go install` (recommended)

The binary is installed directly, no golangci-lint clone needed:

```sh
go install github.com/slipros/gid-data-golang-eval/cmd/custom-gcl@latest
```

`custom-gcl` lands in `$(go env GOPATH)/bin` (add it to `PATH`). To upgrade, rerun
`go install` with a newer tag. A service only needs its `.golangci.yml` —
nothing else to clone or copy.

### Option B — `golangci-lint custom` (.custom-gcl.yml)

A local binary inside the project (requires golangci-lint v2.12.2 installed):

```yaml
# .custom-gcl.yml
version: v2.12.2
name: custom-gcl
destination: ./bin
plugins:
  - module: 'github.com/slipros/gid-data-golang-eval'
    version: vX.Y.Z          # latest release tag (see Releases); or path: /local/path for development
```

Build: `golangci-lint custom` → `./bin/custom-gcl`.

### Next (for both options)

1. Start from the reference [.golangci.yml](.golangci.yml) — enable the `gid*`
   linters you need, configure exceptions (`settings.exclude`, `settings.tree`,
   `settings.tags`, …); drop the repo-specific bits (exclusions for `testdata`,
   `giddirtree.settings.tree`).
2. Run: `custom-gcl run ./...` (option A) or `./bin/custom-gcl run ./...` (option B).

## IDE

For diagnostics to show up right in the editor, the IDE must invoke `custom-gcl`
instead of regular golangci-lint. The path is `$(go env GOPATH)/bin/custom-gcl`
for `go install` (option A) or `${workspaceFolder}/bin/custom-gcl` for an
in-project build (option B):

- **VS Code** (`settings.json`):

  ```json
  {
    "go.lintTool": "golangci-lint-v2",
    "go.alternateTools": { "golangci-lint-v2": "custom-gcl" }
  }
  ```

  (`custom-gcl` from `PATH` with `go install`; otherwise the absolute path to the binary.)

- **GoLand**: Settings → Tools → Go Linter (golangci-lint plugin) → point it
  at `custom-gcl`.

## Rule exceptions

Two levels (details in [RULES.md](RULES.md)):

- targeted — `//nolint:<linter>` with a justification comment;
- centralized — the linter's `settings` in `.golangci.yml`
  (e.g. `gidcreateupdate.settings.exclude`, `giddbtags.settings.tags`,
  `giddirtree.settings.tree`).

## Adding a new rule

The process is at the end of [RULES.md](RULES.md): registry row → `.feature` spec →
implementation → **mandatory eval** (analysistest, 4 case classes) → enable in the config.
