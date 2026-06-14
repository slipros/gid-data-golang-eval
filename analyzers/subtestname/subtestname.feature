# language: en

Feature: GID-191 — subtest names in t.Run/b.Run without spaces or slashes (subtest-naming)
  As a developer
  I want subtest names to be snake_case without spaces or slashes
  So that `go test -run 'Test/name'` finds the specific subtest

  # Google rule: "subtest names". go test -run matches subtests by name,
  # replacing spaces with '_' and using '/' as the level separator. A name with
  # a space or '/' breaks an exact -run.
  #
  # Analyzer gidsubtestname, LoadMode TypesInfo. We match a call of the Run method on
  #   a receiver of type *testing.T / *testing.B (the receiver type from the testing
  #   package, determined via pass.TypesInfo) where the first argument is a string
  #   LITERAL or CONSTANT (the value from pass.TypesInfo.Types[arg].Value,
  #   constant.Kind == String) containing a space (' ') or a slash ('/').
  # NOT matched:
  #   - a non-constant name (tt.name from table-driven) — the value is unknown
  #     statically; names in the table are a separate area of responsibility (review);
  #   - a Run method on any type outside the testing package (an own type with Run(string,func)).
  # Generated code (ast.IsGenerated) is skipped.
  # Only *_test.go files are checked (analysistest supports them in testdata).

  # === Class 1: positive (the violation is caught) ===

  Scenario: positive — t.Run with a space in the literal
    Given "t *testing.T" and the call "t.Run(\"with space\", func(t *testing.T) {})"
    When the analyzer checks the file
    Then the diagnostic "GID-191: subtest name \"with space\" contains a space. Fix: use snake_case, go test -run 'Test/name' will not match it" is reported

  Scenario: positive — t.Run with a slash in the literal
    Given "t *testing.T" and the call "t.Run(\"a/b\", func(t *testing.T) {})"
    When the analyzer checks the file
    Then the diagnostic "GID-191: subtest name \"a/b\" contains a slash '/'. Fix: use snake_case, go test -run 'Test/name' will not match it" is reported

  Scenario: positive — b.Run with a space in the literal
    Given "b *testing.B" and the call "b.Run(\"x y\", func(b *testing.B) {})"
    When the analyzer checks the file
    Then the diagnostic "GID-191: subtest name \"x y\" contains a space. Fix: use snake_case, go test -run 'Test/name' will not match it" is reported

  Scenario: positive — t.Run with a constant whose value contains a space
    Given the const "nameWithSpace = \"has space\"" and the call "t.Run(nameWithSpace, func(t *testing.T) {})"
    When the analyzer checks the file
    Then the diagnostic "GID-191: subtest name \"has space\" contains a space. Fix: use snake_case, go test -run 'Test/name' will not match it" is reported

  # === Class 2: negative (clean code passes) ===

  Scenario: negative — a snake_case name
    Given the call "t.Run(\"ok_name\", func(t *testing.T) {})"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — a CamelCase name without spaces or slashes
    Given the call "t.Run(\"CamelCase\", func(t *testing.T) {})"
    When the analyzer checks the file
    Then no diagnostic is reported

  # === Class 3: boundary ===

  Scenario: boundary — table-driven tt.name is not matched
    Given a table with a name field and the call "t.Run(tt.name, func(t *testing.T) {})"
    When the analyzer checks the file
    Then no diagnostic is reported
    # Limitation: a non-constant name (tt.name) is statically unknown — names in the
    # table are checked at review, not by this linter.

  Scenario: boundary — an own type with a Run(string, func) method is not matched
    Given the type "fakeT" with the method "Run(name string, fn func())" and the call "ft.Run(\"with space\", func() {})"
    When the analyzer checks the file
    Then no diagnostic is reported
    # The receiver is not *testing.T/*testing.B — the Run method of a foreign type is out of scope.

  # === Class 4: non-applicability ===

  Scenario: non-applicability — an ordinary .go file without testing
    Given a package without a testing import and without t.Run/b.Run calls
    When the analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-191)
#  [x] Layer chosen: go/analysis (analyzer gidsubtestname in analyzers/subtestname)
#  [x] Severity and message are defined ("GID-191: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest (_test.go files in the testdata package)
#  [ ] Rule enabled in .golangci.yml
