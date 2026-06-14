# language: en

Feature: GID-186 — the format string of printf functions is a literal or a const (fmtconst)
  As a developer
  I want the format string of printf-style functions to be a literal or a constant,
  not a variable
  So that go vet (the printf check) can statically verify the format verbs against the arguments

  # Analyzer gidfmtconst, LoadMode TypesInfo.
  # Target functions and the format-argument index are recognized by the typed
  # package path (pass.TypesInfo, typeutil.Callee):
  #   - fmt.Printf/Sprintf/Errorf → format arg 0; fmt.Fprintf → arg 1;
  #   - github.com/pkg/errors Errorf → 0, Wrapf → 1, WithMessagef → 1;
  #   - log.Printf/Fatalf → arg 0.
  # Constancy is checked via tv.Value != nil (a literal, a const identifier,
  # a concatenation of constants). pkg/errors — a stub in testdata.
  # Generated code (ast.IsGenerated) is skipped.

  # --- Class 1: positive (the violation is caught) ---

  Scenario: positive — a variable in fmt.Sprintf
    Given "func f(s string, x int) string { return fmt.Sprintf(s, x) }"
    When the gidfmtconst analyzer checks the file
    Then the diagnostic "GID-186: the format string is a variable. Fix: declare a const, otherwise vet cannot check the arguments" is reported on "s"

  Scenario: positive — a variable in the format position of fmt.Fprintf (arg 1)
    Given "fmt.Fprintf(w, s, x)" where s is a variable
    When the gidfmtconst analyzer checks the file
    Then the diagnostic "GID-186: the format string is a variable …" is reported on "s"

  Scenario: positive — a variable in errors.Wrapf (arg 1)
    Given "errors.Wrapf(err, s, x)" where s is a variable
    When the gidfmtconst analyzer checks the file
    Then the diagnostic "GID-186: the format string is a variable …" is reported on "s"

  Scenario: positive — fmt.Printf / fmt.Errorf / errors.Errorf / errors.WithMessagef / log.Printf / log.Fatalf
    Given a variable in the format position of each of the listed functions
    When the gidfmtconst analyzer checks the file
    Then the diagnostic "GID-186: the format string is a variable …" is reported on each

  # --- Class 2: negative (clean code passes) ---

  Scenario: negative — a string literal
    Given "fmt.Sprintf(\"value %d\", x)"
    When the gidfmtconst analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — a const identifier
    Given "const fmtStr = \"value %d\"" and "fmt.Sprintf(fmtStr, x)"
    When the gidfmtconst analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — a concatenation of constants
    Given "fmt.Sprintf(\"a\"+\"b %d\", x)"
    When the gidfmtconst analyzer checks the file
    Then no diagnostic is reported
    # The value of a concatenation of constants is also a constant (tv.Value != nil).

  # --- Class 3: boundary (looks like a violation but is allowed) ---

  Scenario: boundary — fmt.Sprint (not printf, no format position)
    Given "fmt.Sprint(s)" where s is a variable
    When the gidfmtconst analyzer checks the file
    Then no diagnostic is reported
    # Sprint formats operands with default formats, there is no format string.

  Scenario: boundary — a local function printf(format, ...)
    Given "func printf(format string, args ...any) {}" and the call "printf(s, x)"
    When the gidfmtconst analyzer checks the file
    Then no diagnostic is reported
    # The function is not from the target packages (typeutil.Callee → a different path) — not matched.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — a file without fmt/log/pkg-errors
    Given a package without printf functions (only string concatenation)
    When the gidfmtconst analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-186)
#  [x] Layer chosen: go/analysis (package fmtconst: gidfmtconst)
#  [x] Message is defined ("GID-186: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml
