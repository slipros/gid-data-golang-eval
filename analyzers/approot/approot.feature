# language: en

Feature: GID-246 — a struct named *Adapter is needless indirection (approot)
  As a developer
  I want structs whose name carries "adapter" to be flagged
  So that agents stop spawning pointless adapter wrappers (esp. in app/api/wiring)
     and the adaptation lives inline where the dependency is actually consumed

  # One analyzer, linter gidapproot, LoadModeSyntax is enough (names/structure only).
  # The rule fires repository-wide — an adapter is a smell wherever it appears, not
  # only in the composition root. Legitimate infrastructure adapters are exempted by
  # directory (settings.exclude-paths) or by type name (settings.exclude).
  # Generated code and _test.go files (mocks, stubs) are skipped.
  #
  # Motivation (incident 2026-07-15): govorun-server's app/api/wiring filled up with
  # DedupAdapter/…Adapter structs that added nothing. An adapter only converts data
  # between layers — and that is the job of the layer's convert subpackage
  # (domain/service/convert, dal/repository/convert), not a standalone adapter type.
  #
  # Match (name-based, deliberately broad):
  #   - a struct type declaration (not an interface, alias, or func type)
  #   - whose name contains "adapter" (case-insensitive substring),
  #   - whose package path is not under settings.exclude-paths,
  #   - whose name is not in settings.exclude.
  # Fix: drop the adapter and move the mapping into <layer>/convert.

  # --- Class 1: positive (the violation is caught) ---

  Scenario: positive — an adapter struct in app/api/wiring
    Given a package in "/internal/app/api/wiring" with "type DedupAdapter struct{...}"
    When the gidapproot analyzer checks the file
    Then the diagnostic "GID-246: \"DedupAdapter\" is an adapter struct …" is reported on "DedupAdapter"

  Scenario: positive — "adapter" as a lowercase substring, not a suffix
    Given a package with "type adapterCache struct{...}"
    When the gidapproot analyzer checks the file
    Then the diagnostic "GID-246: \"adapterCache\" is an adapter struct …" is reported

  Scenario: positive — an adapter struct OUTSIDE the app layer (repository-wide)
    Given a package in "/internal/client/dedup" with "type Adapter struct{...}"
    When the gidapproot analyzer checks the file
    Then the diagnostic "GID-246 …" is reported on "Adapter"
    # The rule is not scoped to app — an adapter is a smell wherever it appears.

  # --- Class 2: negative (clean code passes) ---

  Scenario: negative — a struct without "adapter" in its name, even with methods
    Given a package with "type MetricsObserver struct{...}; func (m *MetricsObserver) Observe(...) {...}"
    When the gidapproot analyzer checks the file
    Then no diagnostic is reported
    # The rule looks only at the name; it does not care about methods.

  Scenario: negative — a func type named AdapterFunc
    Given a package with "type AdapterFunc func(ctx context.Context) error"
    When the gidapproot analyzer checks the file
    Then no diagnostic is reported
    # Only struct type declarations are scoped.

  # --- Class 3: boundary (looks like a violation but is allowed) ---

  Scenario: boundary — an interface named CacheAdapter is a consumer-side port
    Given a package with "type CacheAdapter interface{...}"
    When the gidapproot analyzer checks the file
    Then no diagnostic is reported
    # Interfaces named *Adapter are ports declared on the consumer side — legitimate.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — settings.exclude exempts a specific type name
    Given settings.exclude contains "LegacyAdapter" and that struct exists
    When the gidapproot analyzer checks the file
    Then no diagnostic is reported for "LegacyAdapter" (others still flagged)

  Scenario: non-applicability — settings.exclude-paths exempts a directory
    Given settings.exclude-paths contains "client" and an Adapter struct sits under /internal/client
    When the gidapproot analyzer checks the file
    Then no diagnostic is reported
    # Legitimate infrastructure adapters live under an excluded path.

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-246)
#  [x] Layer chosen: go/analysis (package approot: gidapproot)
#  [x] Message is defined ("GID-246: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml
