# language: en

Feature: GID-190 — error in function results (error-last)
  As a developer
  I want error to be the last result and to be returned as an interface
  So that the Google convention is followed and the typed-nil trap is avoided

  # Source: Google best practices (errors).
  # Analyzer giderrlast, LoadMode TypesInfo. Two checks of the results
  # of functions and methods (pass.TypesInfo):
  #   1. the results contain the error type followed by other
  #      results → error must be the last one;
  #   2. a result is a CONCRETE type implementing error (a named type
  #      or a pointer to it: *MyError / MyError), not the error interface.
  #      A concrete type in an interface position causes a typed-nil trap.
  #      Checked via types.Implements on the result type and on a pointer
  #      to it; the error interface is taken from types.Universe.
  # NOT matched by check 2:
  #   - the error interface itself;
  #   - interfaces extending error (a custom error interface is deliberate);
  #   - error constructor functions in error.go / errors.go files
  #     (a concrete type is legitimate there — those are constructors).
  # Generated code (ast.IsGenerated) is skipped.
  # Targeted suppression — the standard //nolint:giderrlast.

  # === Class 1: positive (violations) ===

  Scenario: positive — error is not last (error, int)
    Given the function "func f() (error, int)"
    When the analyzer checks the file
    Then the diagnostic "GID-190: error must be the last return value. Fix: move it to the end" is reported

  Scenario: positive — the concrete error type *MyError
    Given the function "func g() *MyError" where *MyError implements error
    When the analyzer checks the file
    Then the diagnostic "GID-190: return the error interface, not *errlast.MyError. Fix: a concrete type in the error position causes a typed-nil trap" is reported

  Scenario: positive — a method with (err error, ok bool)
    Given the method "func (t T) Do() (err error, ok bool)"
    When the analyzer checks the file
    Then the diagnostic "GID-190: error must be the last return value. Fix: move it to the end" is reported

  # === Class 2: negative (clean code) ===

  Scenario: negative — (int, error)
    Given the function "func ok1() (int, error)"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — (T, error) where T is an ordinary struct
    Given the function "func ok2() (T, error)" where T is a struct without an Error method
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — error as the only result
    Given the function "func e() error"
    When the analyzer checks the file
    Then no diagnostic is reported

  # === Class 3: boundary ===

  Scenario: boundary — a custom error interface is not matched
    Given the declaration "type ErrIface interface { error; Code() int }" and the function "func h() ErrIface"
    When the analyzer checks the file
    Then no diagnostic is reported
    # ErrIface is an interface extending error; a deliberate decision of the author.

  Scenario: boundary — the constructor NewMyError() *MyError in errors.go
    Given the function "func NewMyError() *MyError" in the file errors.go
    When the analyzer checks the file
    Then no diagnostic is reported
    # Error constructors in error.go/errors.go legitimately return a concrete type.

  Scenario: boundary — the only result is (error)
    Given the function "func single() error"
    When the analyzer checks the file
    Then no diagnostic is reported
    # error in the only (last) position is the norm.

  # === Class 4: non-applicability ===

  Scenario: non-applicability — a function without error in the results
    Given the function "func plain() (int, string)"
    When the analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-190)
#  [x] Layer chosen: go/analysis (analyzer giderrlast in analyzers/errlast)
#  [x] Severity and message are defined ("GID-190: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml
