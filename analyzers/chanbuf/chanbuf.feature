# language: en

Feature: GID-179 — channel buffer size only 0 or 1 (chan-buffer-size)
  As a developer
  I want make(chan T, N) to use a buffer of 0 or 1
  So that a larger buffer does not mask a synchronization problem without an explicit justification

  # Uber rule: "channel size is one or none".
  # Analyzer gidchanbuf, LoadMode TypesInfo. The size is computed via
  #   pass.TypesInfo.Types[expr].Value (constant.Int):
  #   - a constant > 1 (literal, named const, const expression) — matched;
  #   - 0 and 1 (or make without a size) — OK;
  #   - not a constant (variable, call) — NOT matched (justified at runtime, review decides).
  # Only make for channels is matched: make([]T, N) and make(map[K]V, N) are skipped.
  # Generated code (ast.IsGenerated) is skipped.
  # Targeted suppression — the standard //nolint:gidchanbuf.

  # === Class 1: positive (constant buffer > 1) ===

  Scenario: positive — literal size greater than 1
    Given the expression "make(chan int, 2)"
    When the analyzer checks the file
    Then the diagnostic "GID-179: channel buffer 2 is not allowed (only 0 or 1). Fix: use an unbuffered channel or buffer 1, or justify a larger buffer with //nolint:gidchanbuf." is reported

  Scenario: positive — size from a named const (maxWorkers = 10)
    Given the const "maxWorkers = 10" and the expression "make(chan int, maxWorkers)"
    When the analyzer checks the file
    Then the diagnostic "GID-179: channel buffer 10 is not allowed (only 0 or 1). Fix: use an unbuffered channel or buffer 1, or justify a larger buffer with //nolint:gidchanbuf." is reported

  Scenario: positive — size from a constant expression
    Given the expression "make(chan string, 2*3)"
    When the analyzer checks the file
    Then the diagnostic "GID-179: channel buffer 6 is not allowed (only 0 or 1). Fix: use an unbuffered channel or buffer 1, or justify a larger buffer with //nolint:gidchanbuf." is reported

  # === Class 2: negative (buffer 0 or 1, or no size) ===

  Scenario: negative — channel without a size
    Given the expression "make(chan int)"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — buffer 0
    Given the expression "make(chan int, 0)"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — buffer 1
    Given the expression "make(chan int, 1)"
    When the analyzer checks the file
    Then no diagnostic is reported

  # === Class 3: boundary (not matched) ===

  Scenario: boundary — size given by a variable
    Given the variable "n int" and the expression "make(chan int, n)"
    When the analyzer checks the file
    Then no diagnostic is reported
    # The size is not a constant — justified at runtime, review decides.

  Scenario: boundary — size given by a function call
    Given the expression "make(chan int, size())"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — make of a slice with size > 1
    Given the expression "make([]int, 5)"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — make of a map with size > 1
    Given the expression "make(map[string]int, 5)"
    When the analyzer checks the file
    Then no diagnostic is reported

  # === Class 4: non-applicability ===

  Scenario: non-applicability — a file without make
    Given a package without a single make call
    When the analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-179)
#  [x] Layer chosen: go/analysis (analyzer gidchanbuf in analyzers/chanbuf)
#  [x] Severity and message are defined ("GID-179: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml
