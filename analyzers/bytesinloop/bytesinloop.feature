# language: en

Feature: GID-182 — conversion of a string literal/constant to []byte/[]rune inside a loop (bytes-in-loop)
  As a developer
  I want a conversion of a string constant to []byte/[]rune to be computed once before the loop
  So that the same allocation and copy is not performed on every iteration

  # Uber rule: "avoid repeated string-to-byte conversions".
  # Analyzer gidbytesinloop, LoadMode TypesInfo. We match the conversion
  #   []byte(X) / []rune(X), where X is a string LITERAL or CONSTANT
  #   (value from pass.TypesInfo.Types[X].Value, type types.IsString),
  #   located inside a loop body (for/range), including nested blocks
  #   and bodies of closures declared in the loop (executed on every iteration).
  # We do NOT match:
  #   - conversion of a variable/parameter (not a constant) — the value changes;
  #   - conversion outside a loop — it is computed once anyway.
  # Generated code (ast.IsGenerated) is skipped.

  # === Class 1: positive (constant conversion in a loop) ===

  Scenario: positive — []byte of a literal in a for loop
    Given the body "for i := 0; i < 10; i++ { _ = []byte(\"hello\") }"
    When the analyzer checks the file
    Then the diagnostic "GID-182: converting to []byte inside a loop repeats the allocation. Fix: compute it once before the loop." is reported

  Scenario: positive — []byte of a constant in a range loop
    Given the const "constStr = \"const\"" and the body "for range items { _ = []byte(constStr) }"
    When the analyzer checks the file
    Then the diagnostic "GID-182: converting to []byte inside a loop repeats the allocation. Fix: compute it once before the loop." is reported

  Scenario: positive — []rune of a literal in a nested block of a loop
    Given the body "for i := 0; i < 10; i++ { if i > 5 { { _ = []rune(\"world\") } } }"
    When the analyzer checks the file
    Then the diagnostic "GID-182: converting to []rune inside a loop repeats the allocation. Fix: compute it once before the loop." is reported

  # === Class 2: negative (clean code) ===

  Scenario: negative — []byte of a literal outside a loop
    Given the expression "_ = []byte(\"hello\")" outside any loop
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — []byte of a variable in a loop
    Given the parameter "s string" and the body "for i := 0; i < 10; i++ { _ = []byte(s) }"
    When the analyzer checks the file
    Then no diagnostic is reported
    # Conversion of a variable (not a constant) — the value may change, it cannot be hoisted.

  # === Class 3: boundary ===

  Scenario: boundary — a closure declared in the loop contains []byte of a literal
    Given the body "for i := 0; i < 10; i++ { fn := func() { _ = []byte(\"closure\") }; fn() }"
    When the analyzer checks the file
    Then the diagnostic "GID-182: converting to []byte inside a loop repeats the allocation. Fix: compute it once before the loop." is reported
    # The closure runs on every iteration — the conversion is repeated.

  Scenario: boundary — []byte of a closure parameter in a loop
    Given the body "for i := 0; i < 10; i++ { fn := func(s string) { _ = []byte(s) }; fn(\"x\") }"
    When the analyzer checks the file
    Then no diagnostic is reported
    # s is a parameter (not a constant), it is not matched.

  # === Class 4: non-applicability ===

  Scenario: non-applicability — a file without loops or conversions
    Given a package without a single loop or conversion to []byte/[]rune
    When the analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-182)
#  [x] Layer chosen: go/analysis (analyzer gidbytesinloop in analyzers/bytesinloop)
#  [x] Severity and message are defined ("GID-182: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml
