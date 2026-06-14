# language: en

Feature: GID-136 — errors.New only in a package-level var (errnew)
  As a developer
  I want errors.New (github.com/pkg/errors) to be called only in a
  package-level var declaration
  So that static errors are declared up front (var ErrX) rather than
  constructed at runtime

  # One analyzer errnew → linter giderrnew, LoadModeTypesInfo.
  # pkg/errors is recognized by the import path github.com/pkg/errors via
  # TypesInfo (a stub in testdata).
  # The rule covers only errors.New from github.com/pkg/errors:
  #   - errors.Errorf — NOT covered (dynamic context is legitimate, governed by GID-144/145);
  #   - std errors.New — NOT covered (already forbidden by GID-146);
  #   - New from any other package — NOT covered.
  # Runtime = the body of a function, method, or func literal (even when the literal
  # is assigned to a package-level var). Generated code (ast.IsGenerated) is skipped.

  Scenario: positive — errors.New in a function body
    Given the package imports "github.com/pkg/errors"
    And "errors.New(...)" is called in the body of the function "loadSomething"
    When the analyzer checks the file
    Then a "GID-136" diagnostic is reported on the call

  Scenario: positive — errors.New in a method body
    Given the package imports "github.com/pkg/errors"
    And "errors.New(...)" is called in the body of the method "Repo.Find"
    When the analyzer checks the file
    Then a "GID-136" diagnostic is reported on the call

  Scenario: positive — errors.New in a func literal inside a package-level var
    Given the package imports "github.com/pkg/errors"
    And the package-level "var makeErr = func() error { return errors.New(...) }"
    When the analyzer checks the file
    Then a "GID-136" diagnostic is reported on the call in the literal body

  Scenario: negative — errors.New in a single package-level var
    Given the package imports "github.com/pkg/errors"
    And "var ErrNotFound = errors.New(\"not found\")" is declared
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — errors.New in a var block with several errors
    Given the package imports "github.com/pkg/errors"
    And "ErrConflict" and "ErrLocked" are declared via errors.New in a var block
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — errors.Errorf in a function body
    Given the package imports "github.com/pkg/errors"
    And "errors.Errorf(...)" is called in the body of the function "formatSomething"
    When the analyzer checks the file
    Then no diagnostic is reported
    # (dynamic context is legitimate — the domain of GID-144/145, not GID-136)

  Scenario: boundary — standard errors.New in a function body
    Given the package imports "errors" (stdlib)
    And "errors.New(...)" is called in the body of the function "stdNew"
    When the analyzer checks the file
    Then no GID-136 diagnostic is reported
    # (std errors.New is already forbidden by GID-146)

  Scenario: boundary — a local New function from another package
    Given the package imports "svc/othernew" with its own "New"
    And "othernew.New(...)" is called in the body of the function "otherNew"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — a generated file is skipped
    Given the file "zz_generated.go" is marked "// Code generated ... DO NOT EDIT."
    And "errors.New(...)" is called in a function body
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — a package without github.com/pkg/errors
    Given the package does not import "github.com/pkg/errors"
    And "errors.New(...)" from stdlib is called in a function body
    When the analyzer checks the file
    Then no GID-136 diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (GID-136, errors-new-static)
#  [x] Layer chosen: go/analysis (types needed — TypesInfo for pkg/errors)
#  [x] Severity and message are defined ("GID-136: errors.New at runtime — ...")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml (outside the scope of the task)
