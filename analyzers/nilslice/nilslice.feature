# language: en

Feature: GID-185 — a nil slice is valid, an empty literal []T{} is redundant (nil-slice-style)
  As a developer
  I want to return and declare a nil slice instead of an empty literal []T{}
  So that needless allocations are avoided — a nil slice iterates, supports append, len(nil) == 0

  # Uber/Google rule: "nil is a valid slice".
  # Analyzer gidnilslice, LoadMode TypesInfo (types are needed to tell a
  #   slice apart from an array [N]T and a map map[K]V).
  # We match an empty composite literal []T{} (len(Elts) == 0, type — *types.Slice):
  #   - in a return statement          → "return nil instead of an empty slice";
  #   - in := and in var = (variable initialization) → "declare var s []T".
  # NOT matched: non-empty literals; []T{} as a call argument or a struct field
  #   value (emptiness there can be semantics, e.g. json [] vs null);
  #   arrays [N]T{}; map literals; make([]T, ...) (the domain of prealloc).
  # Generated code (ast.IsGenerated) is skipped.
  # Targeted suppression — the standard //nolint:gidnilslice.

  # === Class 1: positive (an empty slice literal) ===

  Scenario: positive — return of an empty slice literal
    Given a function with "return []int{}"
    When the analyzer checks the file
    Then the diagnostic "GID-185: return nil instead of an empty slice. Fix: a nil slice is valid" is reported

  Scenario: positive — initialization via := with an empty literal
    Given the expression "s := []string{}"
    When the analyzer checks the file
    Then the diagnostic "GID-185: declare a zero-value slice. Fix: var s []T" is reported

  Scenario: positive — initialization via var = with an empty literal
    Given the expression "var s = []byte{}"
    When the analyzer checks the file
    Then the diagnostic "GID-185: declare a zero-value slice. Fix: var s []T" is reported

  # === Class 2: negative (correct code) ===

  Scenario: negative — return nil
    Given a function with "return nil"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — a zero-value slice declaration
    Given the expression "var s []int"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — a non-empty slice literal
    Given a function with "return []int{1}"
    When the analyzer checks the file
    Then no diagnostic is reported
    # A non-empty literal is data, not "emptiness".

  # === Class 3: boundary (not matched) ===

  Scenario: boundary — []T{} as a call argument
    Given the expression "f([]int{})"
    When the analyzer checks the file
    Then no diagnostic is reported
    # An empty non-nil slice can be semantics (json [] vs null).

  Scenario: boundary — []T{} as a struct field value
    Given the expression "T{X: []int{}}"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — the array [0]int{}
    Given the expression "[0]int{}"
    When the analyzer checks the file
    Then no diagnostic is reported
    # An array is not a slice, filtered out via TypesInfo.

  Scenario: boundary — the map literal map[string]int{}
    Given the expression "map[string]int{}"
    When the analyzer checks the file
    Then no diagnostic is reported

  # === Class 4: non-applicability ===

  Scenario: non-applicability — a file without slice literals
    Given a package without a single slice literal
    When the analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-185)
#  [x] Layer chosen: go/analysis (analyzer gidnilslice in analyzers/nilslice)
#  [x] Severity and message are defined ("GID-185: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml
