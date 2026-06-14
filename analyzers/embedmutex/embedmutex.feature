# language: en
# Specification of rule GID-178 (no-embed-mutex).
# Linter: gidembedmutex (go/analysis, LoadModeTypesInfo).

Feature: GID-178 — ban on embedding sync.Mutex/sync.RWMutex in structs
  As a backend-go service developer
  I want the mutex to be stored as a named unexported field (mu sync.Mutex)
  So that the Lock/Unlock methods are not promoted via embedding into the type's public API

  # Embedding (an anonymous field) of sync.Mutex / sync.RWMutex (and pointers to them)
  # promotes Lock/Unlock outward: external code can lock someone else's mutex.
  # Detection is via go/types (pass.TypesInfo), not by selector text, so it works
  # reliably with import aliases of the sync package: an anonymous field whose
  # type, after dereferencing the pointer, is the named type Mutex/RWMutex from
  # the "sync" package. A named field of any kind is OK. Embedding into interfaces
  # cannot happen — not affected. Generated code (ast.IsGenerated) is skipped.

  # --- Positive cases (the violation is caught) ---

  Scenario: positive — embedded sync.Mutex
    Given a struct with the anonymous field "sync.Mutex"
    When the analyzer checks the file
    Then the diagnostic "GID-178: sync.Mutex is embedded in the struct. Fix: use a named mutex field (mu sync.Mutex), otherwise Lock/Unlock leak into the type's API" is reported on the embedded field

  Scenario: positive — embedded pointer *sync.RWMutex
    Given a struct with the anonymous field "*sync.RWMutex"
    When the analyzer checks the file
    Then the diagnostic "GID-178: sync.RWMutex is embedded in the struct. Fix: use a named mutex field (mu sync.Mutex), otherwise Lock/Unlock leak into the type's API" is reported on the embedded field

  Scenario: positive — embedding via an import alias of the sync package
    Given the import "syncalias \"sync\"" and a struct with the anonymous field "syncalias.Mutex"
    When the analyzer checks the file
    Then the diagnostic "GID-178: sync.Mutex is embedded in the struct …" is reported on the embedded field
    # (Matched by type via go/types, not by selector text.)

  # --- Negative cases (clean code passes) ---

  Scenario: negative — a named unexported field mu sync.Mutex
    Given a struct with the field "mu sync.Mutex"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — a named pointer field mu *sync.RWMutex
    Given a struct with the field "mu *sync.RWMutex"
    When the analyzer checks the file
    Then no diagnostic is reported

  # --- Boundary cases (looks similar but is not matched) ---

  Scenario: boundary — an embedded own Mutex type (not from sync)
    Given the declaration "type Mutex struct{}" and a struct with the anonymous field "Mutex"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — embedded sync.WaitGroup (another type from sync, not a mutex)
    Given a struct with the anonymous field "sync.WaitGroup"
    When the analyzer checks the file
    Then no diagnostic is reported

  # --- Non-applicability (the rule does not apply) ---

  Scenario: non-applicability — the package has no structs
    Given a package without struct declarations (only functions and interfaces)
    When the analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-178)
#  [x] Layer chosen: go/analysis (types needed — Mutex/RWMutex from sync via go/types)
#  [x] Severity and message are defined ("GID-178: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml
