# CLAUDE.md

A custom golangci-lint plugin (module plugin system) that ports the rules of the internal
gid.team styleguide (skill `go-styleguide`) into a deterministic linter. Every rule
has an ID `GID-NNN` and is registered in [RULES.md](RULES.md).

## Commands

```bash
make build         # build the bin/custom-gcl binary (golangci-lint custom)
make eval          # run the eval of all rules (go test ./...)
make lint-fast     # check the repository code with the built binary
go test ./analyzers/<slug>/...   # eval of a single rule
```

The build requires golangci-lint **v2.9.0** — the version is pinned in `.custom-gcl.yml`.
Dependency versions are pinned to golangci v2.9.0 — do not upgrade without verifying the build.

## Structure

- `analyzers/<slug>/` — go/analysis analyzers: one rule (or a group of related GID-IDs) = one linter `gid<slug>`
- `analyzers/patterns/` — simple AST patterns (GID-001…008), layer 1
- `plugin.go` — registration of all analyzers in the plugin system
- `internal/pathseg` — matching layers by path segments (`/domain/model`, `/dal/entity`, …)
- `internal/exclude` — parsing of `settings.exclude` (`Method` | `Type.Method`)
- `.golangci.yml` — the reference config: each linter with a `desc` and example settings
- `RULES.md` — the rule registry with statuses; the single source of truth on rules

## Hard requirements

1. **Every rule must have an eval.** A rule is not considered done without
   `analysistest` + `testdata/src/...` with `// want`, covering 4 case classes:
   positive, negative, boundary, non-applicability (template — `rule_template.feature`).
2. The process for adding a rule (end of RULES.md): registry row → `.feature` spec →
   implementation → eval → enable in `.golangci.yml` → update the status in RULES.md.
3. UUID — only `github.com/gofrs/uuid` (we enforce this ourselves with rule GID-137).
4. Errors — only `github.com/pkg/errors` (GID-146).
5. eval fixtures in `testdata/` deliberately violate the rules — do not "fix" them.

## Analyzer conventions

- Linter name: `gid<slug>` without hyphens (`gidnogetprefix`).
- Settings via `settings` in `.golangci.yml`; pinpoint exclusions via
  `//nolint:<linter>`; centralized ones — `settings.exclude` / `settings.tree` /
  `settings.tags`, etc.
- The package layer is determined by path segments through `internal/pathseg`,
  not by a string `strings.Contains`.
- Diagnostics and `description`/`Doc` are formulated **in English** in the format
  `<problem>. Fix: <example>.` — each message contains a valid fix
  example. Accordingly, the `// want` comments in testdata are written in English.
