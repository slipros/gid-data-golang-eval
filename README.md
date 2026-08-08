# gid-data-golang-eval

A custom golangci-lint plugin that turns the internal style guide
(skill `go-styleguide`) into a deterministic linter for **local development**.

## Status: the linter fully replaces the style part of the `go-styleguide` skill

Verified on 2026-06-07 by a full cross-check of all skill docs (31 files) against the registry:

- every deterministically checkable style-guide rule is implemented (GID-001…GID-246,
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
- `.golangci.yml` — this repo's own config (self-lint): all linters plus
  repo-specific settings (its own `giddirtree` tree, testdata exclusions). The
  binary does not read it by itself, so `make lint-fast` passes `--config`;
- `gid-golangci.yml` — **distributable config for services**: the same rule
  set with the canonical gid.team `internal/` tree (app, client, dal, domain,
  event, job, metric, schedule, server, validate) and no repo-specific
  exclusions. It is **embedded into the custom-gcl binary** (`config.go`,
  `internal/defaultconfig`) and is what a plain run uses — nothing to copy,
  nothing to keep in sync;
  based on the production config of consent-api (UDMP/backend-go) with GID layers on top
- `gid-golangci-rules.yml` — the second embedded config: the `gid*` rules
  **only**, no stock linters and no formatter, selected with `--gid-rules-only`
  (see [Running alongside your own golangci-lint](#running-alongside-your-own-golangci-lint))

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
single config, and that config ships inside the binary. Build it in one of two
ways.

### Option A — `go install` (recommended)

The binary is installed directly, no golangci-lint clone needed:

```sh
go install github.com/slipros/gid-data-golang-eval/cmd/custom-gcl@latest
```

`custom-gcl` lands in `$(go env GOPATH)/bin` (add it to `PATH`). To upgrade, rerun
`go install` with a newer tag. A service needs nothing else — no config to clone
or copy, the binary carries its own.

> **After upgrading the binary run `custom-gcl cache clean`.** The golangci-lint
> result cache keys on the config and the checked sources, not on the gid rules
> embedded in the binary — without cleaning it replays stale diagnostics from
> the previous revision.

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

Just run it: `custom-gcl run ./...` (option A) or `./bin/custom-gcl run ./...`
(option B). **Nothing to configure** — custom-gcl is the gid ruleset, and it
runs with the config it carries.

The distributable [gid-golangci.yml](gid-golangci.yml) is compiled into the
binary. On every run it is written to
`~/.cache/gid-golangci/gid-golangci-<content hash>.yml` and passed as `--config`
— the run says so on stderr. The name is the hash of the config itself, so each
config gets its own file and an upgrade simply writes a new one; it is a run
detail, not something to edit or point at. The binary and its ruleset are
therefore always the same revision.

A `.golangci.yml` lying in the repository is **not** picked up by itself —
unlike regular golangci-lint, which would read it and then run without a single
gid linter. To lint by that config instead, name it:

```sh
custom-gcl run --config .golangci.yml ./...   # the repository ruleset
custom-gcl run --no-config ./...              # the stock golangci-lint set
```

The notice on stderr names the repository config it stepped over, so an ignored
one never goes unnoticed.

To start a repository config from the shipped ruleset:

```sh
custom-gcl gid-config > .golangci.yml     # prints the built-in config
```

Then enable the `gid*` linters you need and configure exceptions
(`settings.exclude`, `settings.tree`, `settings.tags`, …) — and run with
`--config`.

### Running alongside your own golangci-lint

A service that already runs a stock `golangci-lint` with its own
`.golangci.yml` ends up running most of the work twice: the distributable
config enables ~40 stock linters on top of the gid rules, and in such a
repository nearly all of them are enabled a second time — `staticcheck`,
`gocritic`, `gosec`, `revive`, `unparam` and the `goimports` formatter among
them. `--gid-rules-only` selects the second embedded config: the `gid*` rules
and nothing else.

```sh
custom-gcl run --gid-rules-only ./...        # only the gid rules
custom-gcl gid-config --gid-rules-only       # print that config
```

Measured on a 1681-file service: **41.7 s for the full config against 1.33 s
for this one**, with byte-identical gid diagnostics (2717 of them).

It keeps three stock linters a service config almost never enables —
`depguard` (GID-137, the uuid-fork ban), `musttag` and `perfsprint` (GID-208) —
and drops three that carry gid tuning: `staticcheck` `checks: [all]`
(GID-206), the `revive` rule list (GID-207) and the `ireturn` allow-list
(GID-208). Move those into the repository config, or lose the rules.

> **Check what your repository config actually reports before switching.** The
> saving is only real if the stock run covers those linters, and the two
> settings that routinely stop it are `run.tests: false` (the gid rules judge
> tests, the stock linters then would not) and `issues.max-issues-per-linter` /
> `max-same-issues` capping the output. On the service measured above those two
> turned 2461 findings into 3.

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
implementation → **mandatory eval** (analysistest, 4 case classes) → enable it in all
three configs (`.golangci.yml`, `gid-golangci.yml`, `gid-golangci-rules.yml`).
