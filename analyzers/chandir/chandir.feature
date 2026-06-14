# language: en

Feature: GID-189 — direction of channel parameters (chan-direction)
  As a developer
  I want channel parameters in signatures to declare a direction (<-chan/chan<-)
  So that the intent (read from or write to the channel) is explicit and type-protected

  # Google rule: "channel direction".
  # Analyzer gidchandir, LoadMode Syntax (AST is enough):
  #   a parameter of a function/method/function literal whose type is a literal
  #   *ast.ChanType with Dir == SEND|RECV (bidirectional chan T) — matched.
  # Decision on function literals: MATCHED (their parameters are a signature too).
  # Only Params are checked, not Results.
  # NOT matched: <-chan/chan<-, return values, struct fields, local variables,
  #   a named channel type in a parameter (that is an *ast.Ident), []chan T (*ast.ArrayType).
  # Generated code (ast.IsGenerated) is skipped.
  # Targeted suppression — the standard //nolint:gidchandir.

  # === Class 1: positive (bidirectional channel parameter) ===

  Scenario: positive — a function with a chan T parameter
    Given the declaration "func consume(ch chan int)"
    When the analyzer checks the file
    Then the diagnostic "GID-189: channel parameter ch is bidirectional. Fix: declare a direction, <-chan to receive or chan<- to send." is reported

  Scenario: positive — a method with a chan T parameter
    Given the method "func (w worker) run(ch chan string)"
    When the analyzer checks the file
    Then the diagnostic "GID-189: channel parameter ch is bidirectional. Fix: declare a direction, <-chan to receive or chan<- to send." is reported

  Scenario: positive — a function literal with a chan T parameter
    Given the literal "func(ch chan int) { <-ch }"
    When the analyzer checks the file
    Then the diagnostic "GID-189: channel parameter ch is bidirectional. Fix: declare a direction, <-chan to receive or chan<- to send." is reported

  Scenario: positive — several names in one parameter group
    Given the declaration "func multi(a, b chan int)"
    When the analyzer checks the file
    Then a single diagnostic per group is reported: "GID-189: channel parameter a, b is bidirectional. Fix: declare a direction, <-chan to receive or chan<- to send."

  # === Class 2: negative (direction is declared) ===

  Scenario: negative — receive-only channel
    Given the declaration "func recvOnly(ch <-chan int)"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — send-only channel
    Given the declaration "func sendOnly(ch chan<- int)"
    When the analyzer checks the file
    Then no diagnostic is reported

  # === Class 3: boundary (not matched) ===

  Scenario: boundary — bidirectional channel in a return value
    Given the declaration "func produce() chan int"
    When the analyzer checks the file
    Then no diagnostic is reported
    # The channel owner sometimes needs a bidirectional one — review decides.

  Scenario: boundary — a struct field of type chan T
    Given the type "type holder struct { ch chan int }"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — a local channel variable
    Given the expression "var ch chan int"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — a named channel type in a parameter
    Given "type Pipe chan int" and the declaration "func namedParam(p Pipe)"
    When the analyzer checks the file
    Then no diagnostic is reported
    # A named type is a deliberate decision; in the parameter's AST it is an *ast.Ident.

  Scenario: boundary — a slice of channels in a parameter
    Given the declaration "func sliceParam(chs []chan int)"
    When the analyzer checks the file
    Then no diagnostic is reported
    # []chan T is an *ast.ArrayType, not a direct channel parameter.

  # === Class 4: non-applicability ===

  Scenario: non-applicability — a file without channels in signatures
    Given a package without a single channel parameter
    When the analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-189)
#  [x] Layer chosen: go/analysis (analyzer gidchandir in analyzers/chandir)
#  [x] Severity and message are defined ("GID-189: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml
