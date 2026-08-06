# language: en

Feature: GID-146 — errors go through github.com/pkg/errors alone (gidonlypkgerrors)
  As a developer
  I want every error to be created and inspected by one library
  So that each error carries a stack trace and the whole codebase handles
  errors the same way

  # Two detectors in one analyzer, LoadModeTypesInfo.
  # Detector 1 — the import: a file importing the std "errors" package, under
  # any alias, is reported once, on the ImportSpec. pkg/errors v0.9.0+
  # re-exports Is/As/Unwrap (go113.go — one-line delegates to the std
  # functions), so the second package buys nothing.
  # Detector 2 — the call: std errors.New, errors.Join, fmt.Errorf; the callee
  # is resolved through typeutil.Callee, so a local function named Errorf is
  # not mistaken for fmt.Errorf.
  # The two do not double up: inside a file whose import was reported, the std
  # constructor calls stay silent — dropping the import is the single fix.
  # Generated code (ast.IsGenerated) is skipped.
  # Changed 2026-08-06: std errors.Is/As/Unwrap used to be allowed, on the
  # false premise that pkg/errors had no replacement for them.

  # --- Class 1: positive ---

  Scenario: positive — the std package pulled in for a single errors.Is
    Given "stderrors \"errors\"" imported next to "github.com/pkg/errors" and the call "stderrors.Is(err, ErrNoResult)"
    When the gidonlypkgerrors analyzer checks the file
    Then the diagnostic "GID-146: the std errors package is forbidden, github.com/pkg/errors re-exports Is/As/Unwrap. Fix: import \"github.com/pkg/errors\" alone and call errors.Is(err, ErrNoResult)" is reported on the import
    # The incident shape (2026-08-06, resource-registry internal/dal/entity/error.go).

  Scenario: positive — a sentinel through the std errors.New
    Given "stderrors \"errors\"" imported and "var ErrStd = stderrors.New(\"std\")"
    When the gidonlypkgerrors analyzer checks the file
    Then the diagnostic is reported on the import, and the call carries none
    # One defect, one fix: with the import gone the call resolves to pkg/errors.

  Scenario: positive — formatting through fmt.Errorf
    Given "return fmt.Errorf(\"job %s failed\", id)"
    When the gidonlypkgerrors analyzer checks the file
    Then the diagnostic "GID-146: fmt.Errorf is forbidden. Fix: use only github.com/pkg/errors for errors" is reported on the call
    # Reported on the call, not on the import: the fmt import itself is legitimate.

  # --- Class 2: negative ---

  Scenario: negative — pkg/errors constructors
    Given "errors.New(\"job not found\")" and "errors.Errorf(\"job %s failed\", id)" from github.com/pkg/errors
    When the gidonlypkgerrors analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — chain inspection through pkg/errors
    Given "errors.Is(err, errNoResult)", "errors.As(err, target)" and "errors.Unwrap(err)" from github.com/pkg/errors
    When the gidonlypkgerrors analyzer checks the file
    Then no diagnostic is reported

  # --- Class 3: boundary ---

  Scenario: boundary — the std package alone, inspection only
    Given "import \"errors\"" with no pkg/errors in the file and the call "errors.Is(err, target)"
    When the gidonlypkgerrors analyzer checks the file
    Then the diagnostic is reported on the import
    # The rule is not about two imports being redundant: one errors package for
    # the codebase, and it is pkg/errors.

  Scenario: boundary — a local function named Errorf
    Given a package-level "func Errorf(format string, args ...any) error" and a call to it
    When the gidonlypkgerrors analyzer checks the file
    Then no diagnostic is reported

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — the std package in a _test.go file
    Given a "_test.go" file with "stderrors \"errors\"" and "want := stderrors.Is(other.sentinel, p.sentinel)"
    When the gidonlypkgerrors analyzer checks the file
    Then no diagnostic is reported on the import
    # A test of an error predicate takes its expected value from the std
    # function on purpose — checking pkg/errors with pkg/errors proves nothing.

  Scenario: non-applicability of the relaxation — creating errors in a _test.go file
    Given a "_test.go" file with "stderrors.New(\"test sentinel\")" and "stderrors.Join(a, b)"
    When the gidonlypkgerrors analyzer checks the file
    Then both calls are reported
    # The import relaxation covers inspection; creating errors through the std
    # package is judged in a test the same as in production code.

  Scenario: non-applicability — fmt used for formatting, not for errors
    Given "fmt.Sprintf(\"job %s\", id)"
    When the gidonlypkgerrors analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — generated code
    Given a file carrying the "Code generated … DO NOT EDIT." header
    When the gidonlypkgerrors analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-146)
#  [x] Layer chosen: go/analysis (package onlypkgerrors)
#  [x] Message is defined ("GID-146: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
