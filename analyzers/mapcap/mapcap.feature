# language: en

Feature: GID-183 — make(map) without capacity while filling from range (map-capacity-hint)
  As a developer
  I want to give a capacity hint make(map[K]V, len(src)) when filling a map from range
  So that unnecessary map reallocations are avoided (the standard prealloc covers only slices)

  # Uber rule (perf): map capacity hints.
  # Analyzer gidmapcap, LoadMode TypesInfo (types are needed to tell a
  #   slice/map/string apart from a channel). Generated code (ast.IsGenerated) is skipped.
  #
  # The pattern within ONE function (and one statement block):
  #   1. m := make(map[K]V)        — or var m = make(map[K]V), without a capacity argument;
  #   2. for ... := range src {    — src is a slice/array/map/string (known length);
  #          m[...] = ...          — an unconditional index assignment into m.
  #   → a diagnostic on make.
  #
  # This is a HEURISTIC. Deliberate limitations (conservative, to avoid FPs):
  #   - matched only when m is not used in ANY way between make and the loop; any
  #     mention of m in that span (filling outside the loop, passing into a call,
  #     reading) cancels the diagnostic — the length is no longer known by loop time;
  #   - range over a channel is NOT matched (a channel has no len);
  #   - an assignment m[...] = ... inside an if in the loop body (conditional filling)
  #     is NOT matched — the actual number of inserts is < len(src), the hint may hurt;
  #   - make with a capacity already given, make(map[K]V, n), is correct, not matched;
  #   - the check is local to one statement block (the pattern inside a nested
  #     block is caught separately); cross-block/cross-function flows are left to review.

  # === Class 1: positive (make(map) without cap + unconditional filling in range) ===

  Scenario: positive — filling from range over a slice
    Given "m := make(map[int]int)" and the loop "for _, v := range src { m[v] = v }"
    When the analyzer checks the file
    Then the diagnostic "GID-183: make without capacity while filling from range. Fix: make(map[K]V, len(src))" is reported

  Scenario: positive — the var form of the map declaration
    Given "var m = make(map[string]bool)" and filling in a range over a slice
    When the analyzer checks the file
    Then the diagnostic "GID-183: make without capacity while filling from range. Fix: make(map[K]V, len(src))" is reported

  Scenario: positive — filling from range over a map
    Given "m := make(map[string]int)" and the loop "for k, v := range src { m[k] = v }" over a map
    When the analyzer checks the file
    Then the diagnostic "GID-183: make without capacity while filling from range. Fix: make(map[K]V, len(src))" is reported

  Scenario: positive — filling from range over a string
    Given "m := make(map[rune]int)" and the loop "for _, r := range src { m[r] = 1 }" over a string
    When the analyzer checks the file
    Then the diagnostic "GID-183: make without capacity while filling from range. Fix: make(map[K]V, len(src))" is reported

  # === Class 2: negative ===

  Scenario: negative — make with a capacity given
    Given "m := make(map[int]int, len(src))" and filling in a range
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — filling the map without range
    Given "m := make(map[int]int)" and the assignments "m[1] = 1; m[2] = 2"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — range over a channel
    Given "m := make(map[int]int)" and the loop "for v := range ch { m[v] = v }"
    When the analyzer checks the file
    Then no diagnostic is reported
    # A channel has no len — the size is unknown in advance.

  # === Class 3: boundary (not matched) ===

  Scenario: boundary — conditional filling m[...] inside an if in the loop body
    Given "m := make(map[int]int)" and the loop "for _, v := range src { if v > 0 { m[v] = v } }"
    When the analyzer checks the file
    Then no diagnostic is reported
    # The actual number of inserts is less than len(src) — the hint may hurt.

  Scenario: boundary — m is used between make and the loop (filling outside the loop)
    Given "m := make(map[int]int)", then "m[0] = 0", then a range loop
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — m is passed into a call between make and the loop
    Given "m := make(map[int]int)", then "consume(m)", then a range loop
    When the analyzer checks the file
    Then no diagnostic is reported

  # === Class 4: non-applicability ===

  Scenario: non-applicability — make of a slice, not a map
    Given "s := make([]int, 0)" with append filling in a range
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — a package without a single make(map)
    Given a package without make(map[K]V) calls
    When the analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-183)
#  [x] Layer chosen: go/analysis (analyzer gidmapcap in analyzers/mapcap)
#  [x] Severity and message are defined ("GID-183: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml
