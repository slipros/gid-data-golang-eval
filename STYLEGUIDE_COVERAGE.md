# Coverage of styleguide best practices by golangci-lint v2 linters

Sources: [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md),
[Google Go Style Guide](https://google.github.io/styleguide/go/guide),
[Google Go Best Practices](https://google.github.io/styleguide/go/best-practices).

Buckets:
- **DEFAULT** — caught by the default golangci-lint v2 set (`errcheck`, `govet`, `ineffassign`, `staticcheck`, `unused`);
- **OPT-IN** — a ready-made linter/formatter exists; it needs to be enabled and configured;
- **CUSTOM** — no ready-made linter; requires our go/analysis or ruleguard;
- **REVIEW** — cannot be automated, stays on code review.

> ⚠️ **staticcheck nuance in v2:** with the `staticcheck` linter enabled, the default
> is `all` **minus** `ST1000` (package comment), `ST1003` (mixed caps /
> initialisms), `ST1016` (consistent receiver), `ST1020-22` (doc comments).
> That is, SA (bugs), S (simple), QF, and most ST checks (including ST1005
> error strings, ST1012 error naming, ST1006 ban on this/self) — are active.
> The excluded ST checks are enabled via `settings.staticcheck.checks`.

> ⚠️ **Our reference `.golangci.yml` is currently `default: none`** — of the standard
> five, only `errcheck` is enabled. `govet`, `staticcheck`, `ineffassign`,
> `unused` are not running. This is the first candidate for fixing.

---

## 1. What the default golangci-lint v2 five covers

| Rule from the guides | Linter/check |
|---|---|
| Do not ignore errors, `_ = err` (Google: handle errors) | `errcheck` (+ `check-blank: true` — already GID-202) |
| Do not copy mutexes/locks by value (Uber: zero-value mutex; Google: copying) | `govet/copylocks` |
| Field names in composite literals of foreign packages (Uber/Google) | `govet/composites` |
| Printf: format matches arguments, `...f` naming (Uber) | `govet/printf` |
| Leaked context cancel (Google: goroutine lifetimes, partially) | `govet/lostcancel` |
| Loop variable capture (Uber: parallel tests; before Go 1.22) | `govet/loopclosure` |
| Context key is not a basic type (Google: context keys) | `staticcheck SA1029` |
| Typed nil in interface comparison (Google: return error interface, partially) | `staticcheck SA4023` |
| `fmt.Errorf` without formatting → `errors.New` (Uber) | `staticcheck S1028` |
| Error strings: lowercase, no trailing period (Uber/Google) | `staticcheck ST1005` |
| Error naming `ErrX`/`errX` (Uber) | `staticcheck ST1012` |
| Receiver is not `this`/`self`/`me` (Google) | `staticcheck ST1006` |
| Redundant `break`, redundant nil checks, and other simple cases | `staticcheck S1023`, `S1009`, `S1021`, … |
| Unused code / useless assignments | `unused`, `ineffassign` |

**Conclusion:** the default covers the "correctness + basic error hygiene" layer, but
almost nothing from the structural/stylistic layer of the guides.

## 2. OPT-IN: ready-made linters worth discussing for enablement

### High value, no conflicts with our styleguide

| Linter | What it covers from the guides | Note |
|---|---|---|
| `govet` + `staticcheck` + `unused` + `ineffassign` | all of section 1 | currently disabled because of `default: none` |
| `revive` (selected rules) | `indent-error-flow`, `early-return`, `superfluous-else` (Uber: nesting/else), `deep-exit` (Uber: exit in main; Google: log.Fatal), `dot-imports`, `blank-imports` (Google), `use-any`, `exported` (doc comments, Google) | enable only the needed rules |
| `staticcheck ST1003, ST1016, ST1000` | mixed caps, initialisms (URL not Url), consistent receiver, package comment (Google) | add to `settings.staticcheck.checks` |
| `gocritic` (broader than ruleguard) | dozens of small Uber rules: `ifElseChain`, `exitAfterDefer`, `ptrToRefParam` (pointer to interface), etc. | the scaffolding is already wired up for ruleguard |
| `predeclared` | Uber: do not shadow builtin names | — |
| `gochecknoinits` | Uber: avoid `init()` | bootstrap in `internal/app` — exceptions |
| `nakedret` | Google: naked return only in short functions | — |
| `forcetypeassert` *or* `errcheck.check-type-assertions: true` | Uber: comma-ok on type assertion | the errcheck option is off by default |
| `prealloc` | Uber: slice capacity hints | does not cover map hints |
| `perfsprint` | Uber: strconv vs fmt, Sprintf→concatenation (Google: string concatenation) | — |
| `musttag` | Uber: field tags in marshaled structs | complements our GID-125/168 |
| `nestif`, `gocognit`/`gocyclo` | Uber: complexity of table subtests, nesting | thresholds to discuss |
| `importas` | Google: aliases of proto/`pb` imports | — |
| `thelper`, `testpackage`, `paralleltest`, `tparallel` | Google tests: `t.Helper()`, `_test` packages, parallelism | align with the go-testing skill |
| `containedctx` | Google: do not store Context in a struct | — |
| `interfacebloat` | Google: small interfaces | method threshold |
| `ireturn` | Google: accept interfaces, return concrete types | configure an allow-list (error, generics) |
| `grouper` | Uber: grouping const/var/import into blocks | complements GID-130 |
| `wrapcheck` | **our rule "errors from outside are always Wrapped"** (≈GID-140/141) | checks that an err from a foreign package is wrapped; configurable for pkg/errors — a candidate to replace / serve as the basis of a custom rule |

### Conflict with our styleguide — must not be enabled, or only with caveats

| Linter | Conflict |
|---|---|
| `errorlint` (`%w`, `errors.Is`) | the guides are built on std wrapping with `%w`; we have GID-146 — `pkg/errors` only, `fmt.Errorf` is forbidden. Only the `comparison`/`asserts` part is useful (`errors.Is`/`As` instead of `==`) — std `errors.Is/As` are allowed for us |
| `gochecknoglobals` | our package-level `var Default*Options` (GID-126) and `var ErrX` are legitimate. Needs exceptions, or do not enable |
| `err113` | same: pushes toward std errors; resolve the overlap with GID-136 in favor of our rule |
| `depguard` | could replace GID-137/146 (banning uuid forks, testify per Google) — but our custom linters give better messages; Google forbids assert libraries, we use testify (require/assert) — a deliberate deviation |
| `lll` | Google explicitly calls line-length an "invalid local style"; our GID-201 (120) is a deliberate deviation, we keep it |

## 3. CUSTOM: candidates for new GID rules

Deterministic (AST/types), no ready-made linter, no overlap with our rules:

| Candidate | Source | Comment |
|---|---|---|
| Ban embedding `sync.Mutex`/`RWMutex` in a struct | Uber | trivial AST check |
| Channel buffer only 0 or 1 (`make(chan T, N)`, N>1 — diagnostic) | Uber | literals; exception settings |
| Ban goroutines in `init()` / I/O in `init()` | Uber | complements gochecknoinits if we allow init in app |
| `os.Exit`/`log.Fatal` at most once in `main` | Uber | call counter |
| `[]byte("literal")` inside a loop → hoist it out | Uber perf | — |
| Map capacity hint (`make(map, n)` when size is known) | Uber perf | prealloc cannot do this |
| Ban `"failed to ..."` in error messages | Uber | string literal in Wrap/WithMessage |
| `return []T{}` → `return nil`; `var s []T` vs `s := []T{}` | Uber/Google | — |
| `new(T)` → `&T{}`; `T{}` → `var x T` for zero values | Uber | ruleguard level |
| Format string — `const`/literal, not a variable | Uber | complements govet/printf |
| Ban package names `util`/`common`/`helper`/`shared` | Google | extension of GID-158 (dirtree) or standalone |
| Ban custom context types (not `context.Context` in the ctx position) | Google | "no exceptions" per Google |
| Channel direction in signatures (`<-chan`/`chan<-`) | Google | types check of parameters |
| `error` — the last return parameter; do not return a concrete error type | Google | typed-nil trap |
| Yoda conditions (`"foo" == x`) | Google | ruleguard |
| `%q` instead of manual `\"%s\"` | Google | ruleguard |
| `reflect.DeepEqual` in tests → cmp/require | Google | ruleguard |
| Subtest names in `t.Run` without spaces/slashes | Google | literals |
| `flag.*` only in `package main`; flag name in snake_case | Google | if CLIs appear |
| Symbol does not repeat the package name (`widget.NewWidget`) | Google | complements GID-104 |

Discarded as contradicting our styleguide: the `_` prefix for private globals
(Uber-specific), enum start at one (we use string enums, GID-123), interface
compliance verification via `var _` (judgment), go.uber.org/atomic (its own ecosystem).

## 4. REVIEW: not automatable

Clarity/Simplicity/Consistency, choosing the error type by the matrix, in-band errors,
interface design ("the consumer defines it" — partially closed for us by GID-134/173),
goroutine lifecycle, time semantics, functional options vs option struct,
got-before-want in tests, documentation completeness, global state litmus tests.

---

## Summary

| | Uber | Google guide+BP | Unique total |
|---|---|---|---|
| DEFAULT (once the five are enabled) | ~7 | ~9 | ~14 |
| OPT-IN | ~18 | ~25 | ~25 linters/rules |
| CUSTOM deterministic | ~25 | ~15 | ~20 candidates |
| REVIEW | ~12 | ~30 | — |

Discussion priority:
1. Enable the standard five in the reference config (`govet`, `staticcheck`, `unused`, `ineffassign`).
2. Pick up the cheap OPT-INs: `revive` (selectively), ST1003/ST1016/ST1000, `predeclared`, `nakedret`, `forcetypeassert`, `prealloc`, `perfsprint`, `musttag`.
3. `wrapcheck` — as the basis/replacement for the future GID-140/141 (errors from outside → Wrap).
4. Pick the first wave from the CUSTOM candidates (I suggest: embed mutex, channel size, exit-once, util packages, custom context, error last param, DeepEqual in tests).
