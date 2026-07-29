# language: en

Feature: GID-161 — panic only in package main (gidnopanic)
  As a developer
  I want errors returned and handled explicitly outside the bootstrap
  So that a library or a service layer never takes the whole process down

  # One analyzer over call expressions, LoadModeTypesInfo (the callee is
  # resolved to the builtin through pass.TypesInfo.Uses).
  # Scope: the whole package is skipped when its name is "main" — that is the
  # bootstrap, where a failed wiring may legitimately panic.
  # Trigger: a call to the builtin panic. A local function or a variable named
  # panic shadowing the builtin is not it.
  # Generated code (ast.IsGenerated) is skipped.

  # --- Class 1: positive ---

  Scenario: positive — panic with a literal in a service package
    Given "panic(\"boom\")" inside a function of package service
    When the gidnopanic analyzer checks the file
    Then the diagnostic "GID-161: panic is allowed only in package main. Fix: return an error instead" is reported on the call

  Scenario: positive — panic with an error value
    Given "panic(err)" inside a function of package repository
    When the gidnopanic analyzer checks the file
    Then the diagnostic "GID-161: panic is allowed only in package main. Fix: return an error instead" is reported

  # --- Class 2: negative ---

  Scenario: negative — panic in package main
    Given "panic(err)" inside func main
    When the gidnopanic analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — the error is returned
    Given "return errors.Wrap(err, \"load config\")" instead of a panic
    When the gidnopanic analyzer checks the file
    Then no diagnostic is reported

  # --- Class 3: boundary ---

  Scenario: boundary — a local function named panic
    Given a package-level "func panic(v any)" shadowing the builtin, and a call to it
    When the gidnopanic analyzer checks the file
    Then no diagnostic is reported
    # The callee is resolved through the type info: it is not *types.Builtin.

  Scenario: boundary — recover without panic
    Given "defer func() { _ = recover() }()"
    When the gidnopanic analyzer checks the file
    Then no diagnostic is reported
    # The rule bans raising a panic, not surviving one.

  Scenario: boundary — panic inside a test helper of a non-main package
    Given "panic(\"unreachable\")" in a _test.go file of package service
    When the gidnopanic analyzer checks the file
    Then the diagnostic "GID-161: …" is reported
    # Tests are the same package (GID-250) and get the same rule; a deliberate
    # case is silenced with //nolint:gidnopanic.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — generated code
    Given a file carrying the "Code generated … DO NOT EDIT." header
    When the gidnopanic analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — the word panic in a string or a comment
    Given the statement "log.Info(\"do not panic\")"
    When the gidnopanic analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-161)
#  [x] Layer chosen: go/analysis (package nopanic)
#  [x] Message is defined ("GID-161: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
