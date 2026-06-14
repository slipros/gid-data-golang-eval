# language: en

Feature: GID-184 — an error message describes the operation, not the fact of failure (failedto)
  As a developer
  I want the message in errors.Wrap/Wrapf/WithMessage/WithMessagef/Errorf/New
  to describe the operation being performed ("select user"), not the fact of failure ("failed to select user")
  So that unwinding the error chain reads as a sequence of operations

  # One analyzer failedto → linter gidfailedto, LoadModeTypesInfo.
  # pkg/errors is recognized by the import path github.com/pkg/errors via TypesInfo (a stub in testdata).
  # Only a string-literal message is checked; a variable / concatenation with a variable is not matched.
  # Forbidden prefixes (case-insensitive, at a word boundary):
  #   failed to, failed, unable to, error, couldn't, could not, can't, cannot
  # The list is the default, configurable via Settings{Prefixes []string `json:"prefixes"`} (fully replaces the default).
  # Generated code (ast.IsGenerated) is skipped.

  # --- Class 1: positive (the violation is caught) ---

  Scenario: positive — errors.Wrap with "failed to"
    Given "errors.Wrap(err, \"failed to select\")"
    When the gidfailedto analyzer checks the file
    Then the diagnostic "GID-184: error message starts with \"failed to\". Fix: describe the operation, e.g. \"failed to select user\" → \"select user\"" is reported

  Scenario: positive — errors.New("Failed: x") in a var (case-insensitive)
    Given "var ErrSelect = errors.New(\"Failed: x\")"
    When the gidfailedto analyzer checks the file
    Then the diagnostic "GID-184: error message starts with \"failed\"" is reported

  Scenario: positive — errors.WithMessage with "unable to"
    Given "errors.WithMessage(err, \"unable to parse\")"
    When the gidfailedto analyzer checks the file
    Then the diagnostic "GID-184: error message starts with \"unable to\"" is reported

  Scenario: positive — errors.Errorf with "error"
    Given "errors.Errorf(\"error while loading %d\", id)"
    When the gidfailedto analyzer checks the file
    Then the diagnostic "GID-184: error message starts with \"error\"" is reported

  Scenario: positive — errors.Wrapf with "cannot" and errors.WithMessagef with "could not"
    Given "errors.Wrapf(err, \"cannot save %d\", id)" and "errors.WithMessagef(err, \"could not commit %d\", id)"
    When the gidfailedto analyzer checks the file
    Then diagnostics are reported on both calls

  # --- Class 2: negative (clean code passes) ---

  Scenario: negative — the message describes the operation
    Given "errors.Wrap(err, \"select user\")" and "errors.New(\"parse config\")"
    When the gidfailedto analyzer checks the file
    Then no diagnostic is reported

  # --- Class 3: boundary (looks like a violation but is allowed) ---

  Scenario: boundary — "failure mode" (the word failure is not in the list)
    Given "errors.Wrap(err, \"failure mode handling\")"
    When the gidfailedto analyzer checks the file
    Then no diagnostic is reported
    # "failure" does not equal the prefix "failed"/"failed to"; the word boundary protects against substrings.

  Scenario: boundary — fmt.Sprintf is not matched (another package)
    Given "errors.Wrap(err, fmt.Sprintf(\"%s\", \"x\"))"
    When the gidfailedto analyzer checks the file
    Then no diagnostic is reported
    # The message argument is not a string literal but a fmt.Sprintf call.

  Scenario: boundary — std errors.New is not matched (the domain of GID-146)
    Given "stderrors.New(\"failed to do thing\")" from the standard "errors" package
    When the gidfailedto analyzer checks the file
    Then no diagnostic is reported
    # The linter matches only github.com/pkg/errors; std errors is the domain of GID-146.

  Scenario: boundary — a non-literal message (variable / concatenation)
    Given "errors.Wrap(err, msg)" and "errors.Wrap(err, \"failed to \"+name)"
    When the gidfailedto analyzer checks the file
    Then no diagnostic is reported
    # A variable and a concatenation with a variable have no constant value — not checked.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — a file without github.com/pkg/errors
    Given a package with a local function "Wrap(err, \"failed to select\")" and no import of pkg/errors
    When the gidfailedto analyzer checks the file
    Then no diagnostic is reported
    # A same-named local Wrap function is not a pkg/errors call (checked via TypesInfo).

  # --- Class 5: configuration ---

  Scenario: configuration — settings.prefixes fully replaces the default
    Given settings.prefixes = ["oops"], "errors.Wrap(err, \"oops broken\")" and "errors.Wrap(err, \"failed to select\")"
    When the gidfailedto analyzer checks the file
    Then a diagnostic is reported only on "oops broken"; "failed to select" is not caught

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-184)
#  [x] Layer chosen: go/analysis (package failedto: gidfailedto), LoadModeTypesInfo
#  [x] Message is defined ("GID-184: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability, configuration
#  [x] testdata with // want for analysistest + a github.com/pkg/errors stub
#  [ ] Rule enabled in .golangci.yml
