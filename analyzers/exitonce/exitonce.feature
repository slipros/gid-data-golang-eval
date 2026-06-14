# language: en

Feature: GID-181 — exit once / exit in main
  As a developer
  I want the process to terminate in exactly one place — in func main() of the main package
  So that errors are returned up the call stack instead of killing the program at an arbitrary point

  # Detection via TypesInfo (package paths "os", "log", "github.com/sirupsen/logrus").
  # The rule covers:
  #   os.Exit;
  #   log.Fatal / log.Fatalf / log.Fatalln (std log);
  #   logrus.Fatal* / logrus.Exit, including methods of *logrus.Entry and *logrus.Logger (Fatal*).
  # Analyzer gidexitonce, LoadMode = TypesInfo. Generated code (ast.IsGenerated) is skipped.
  #
  # Checks:
  #   1. an exit call outside func main (in another function of the main package or in any non-main package).
  #   2. more than one exit call inside func main (the second and subsequent ones in source order).

  # === Class 1: positive (violations) ===

  Scenario: positive — os.Exit in a helper of the main package (outside func main)
    Given the package "main" with the function "fail" calling "os.Exit(1)"
    When the analyzer checks the file
    Then the diagnostic "GID-181: os.Exit is forbidden outside func main. Fix: return an error up the call stack" is reported on the os.Exit call

  Scenario: positive — log.Fatal in a non-main package
    Given a library package with a function calling "log.Fatal"
    When the analyzer checks the file
    Then the diagnostic "GID-181: log.Fatal is forbidden outside func main. Fix: return an error up the call stack" is reported

  Scenario: positive — logrus.Fatalf in a non-main package
    Given a library package with a call to "logrus.Fatalf"
    When the analyzer checks the file
    Then the diagnostic "GID-181: logrus.Fatalf is forbidden outside func main. Fix: return an error up the call stack" is reported

  Scenario: positive — the *logrus.Logger.Fatal method in a non-main package
    Given a library package with a call to "l.Fatal" on a "*logrus.Logger"
    When the analyzer checks the file
    Then the diagnostic "GID-181: logrus.Fatal is forbidden outside func main. Fix: return an error up the call stack" is reported

  Scenario: positive — two os.Exit calls in func main (duplicate)
    Given the package "main" with a func main containing two "os.Exit" calls
    When the analyzer checks the file
    Then the diagnostic "GID-181: duplicate os.Exit in main. Fix: exit the program in a single place" is reported on the second call

  # === Class 2: negative (clean code) ===

  Scenario: negative — exactly one os.Exit at the end of func main
    Given the package "main" with a func main containing a single "os.Exit"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — a function returns an error instead of terminating the process
    Given the function "func run() error" returning the error up the call stack
    When the analyzer checks the file
    Then no diagnostic is reported

  # === Class 3: boundary cases ===

  Scenario: boundary — defer + a single os.Exit in main
    Given a func main with "defer cleanup()" and a single "os.Exit(0)"
    When the analyzer checks the file
    Then no diagnostic is reported
    # defer does not affect the exit-call counter; there is one call — ok.

  Scenario: boundary — os.Exit inside a closure in main
    Given a func main with a closure calling "os.Exit", and that is the only such call
    When the analyzer checks the file
    Then no diagnostic is reported
    # The closure is lexically inside the main body → the call counts as "in main", not "outside main".

  # === Class 4: non-applicability ===

  Scenario: non-applicability — a library package without exit calls
    Given a library package using only Info/Error/Print (non-fatal)
    When the analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-181)
#  [x] Layer chosen: go/analysis (analyzer gidexitonce in analyzers/exitonce)
#  [x] Messages are defined ("GID-181: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml
