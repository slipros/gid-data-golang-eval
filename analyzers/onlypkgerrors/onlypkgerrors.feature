# language: en

Feature: GID-146 — errors are created only through github.com/pkg/errors (gidonlypkgerrors)
  As a developer
  I want every error to be created by one library
  So that each error carries a stack trace and the whole codebase wraps and
  inspects errors the same way

  # One analyzer over call expressions, LoadModeTypesInfo: the callee is
  # resolved through typeutil.Callee, so a local function named Errorf is not
  # mistaken for fmt.Errorf.
  # Trigger: a call to a std error constructor — errors.New, errors.Join,
  # fmt.Errorf.
  # Inspecting the chain is not creating: std errors.Is, errors.As and
  # errors.Unwrap stay allowed — pkg/errors has no replacement for them.
  # Generated code (ast.IsGenerated) is skipped.

  # --- Class 1: positive ---

  Scenario: positive — a sentinel through std errors.New
    Given "var ErrStd = stderrors.New(\"std\")"
    When the gidonlypkgerrors analyzer checks the file
    Then the diagnostic "GID-146: errors.New is forbidden. Fix: use only github.com/pkg/errors for errors" is reported on the call

  Scenario: positive — formatting through fmt.Errorf
    Given "return fmt.Errorf(\"job %s failed\", id)"
    When the gidonlypkgerrors analyzer checks the file
    Then the diagnostic "GID-146: fmt.Errorf is forbidden. Fix: use only github.com/pkg/errors for errors" is reported

  Scenario: positive — errors.Join
    Given "return stderrors.Join(a, b)"
    When the gidonlypkgerrors analyzer checks the file
    Then the diagnostic "GID-146: errors.Join is forbidden. Fix: use only github.com/pkg/errors for errors" is reported

  # --- Class 2: negative ---

  Scenario: negative — pkg/errors constructors
    Given "errors.New(\"job not found\")" and "errors.Errorf(\"job %s failed\", id)" from github.com/pkg/errors
    When the gidonlypkgerrors analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — wrapping through pkg/errors
    Given "errors.Wrap(err, \"exec context\")"
    When the gidonlypkgerrors analyzer checks the file
    Then no diagnostic is reported

  # --- Class 3: boundary ---

  Scenario: boundary — std errors.Is and errors.As
    Given "stderrors.Is(err, model.ErrNotFound)" and "stderrors.As(err, &target)"
    When the gidonlypkgerrors analyzer checks the file
    Then no diagnostic is reported
    # Inspecting the chain is not creating an error, and pkg/errors offers no
    # substitute for these two.

  Scenario: boundary — std errors imported under an alias
    Given "stderrors \"errors\"" and the call "stderrors.New(\"std\")"
    When the gidonlypkgerrors analyzer checks the file
    Then the diagnostic "GID-146: errors.New is forbidden. …" is reported
    # The callee is resolved by its package path, not by the local alias — the
    # diagnostic still names the real package.

  Scenario: boundary — a local function named Errorf
    Given a package-level "func Errorf(format string, args ...any) error" and a call to it
    When the gidonlypkgerrors analyzer checks the file
    Then no diagnostic is reported

  # --- Class 4: non-applicability ---

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
