# language: en

Feature: GID-192 — flags (flag registration only in main, snake_case names)
  As a developer
  I want only the binary (the main package) to declare flags, while libraries take parameters
  And flag names to be snake_case
  So that binary configuration is uniform and libraries are reusable

  # Detection of the flag package — via TypesInfo (package path "flag", stdlib).
  # The rule covers calls to flag package functions (flag.String/Int/Bool/.../
  # flag.Parse/flag.Var/flag.NewFlagSet) and methods of flag package types (*flag.FlagSet).
  # Analyzer gidflagmain, LoadMode = TypesInfo. Generated code
  # (ast.IsGenerated) is skipped. *_test.go files and packages with the _test
  # suffix are skipped — flag in tests can be legitimate.
  #
  # Checks:
  #   1. Any call to the flag package outside the main package → forbidden.
  #   2. In the main package: the first CONSTANT string name argument
  #      (flag.String/Int/Bool/Duration/Float64/...; for *Var and flag.Var — the second
  #      argument) must be snake_case: uppercase letters and hyphens are forbidden,
  #      digits and `_` are allowed. A camelCase VARIABLE name is not checked
  #      (that is revive/ST1003).

  # === Class 1: positive (violations) ===

  Scenario: positive — flag.String in a library package
    Given a library (non-main) package with a call to "flag.String"
    When the analyzer checks the file
    Then the diagnostic "GID-192: registering a flag outside package main is forbidden. Fix: declare flags in the binary, let libraries take parameters" is reported

  Scenario: positive — flag.Parse in a library
    Given a library package with a call to "flag.Parse()"
    When the analyzer checks the file
    Then the diagnostic "GID-192: registering a flag outside package main is forbidden. Fix: declare flags in the binary, let libraries take parameters" is reported

  Scenario: positive — a *flag.FlagSet method in a library
    Given a library package with a call to "fs.String" on a "*flag.FlagSet"
    When the analyzer checks the file
    Then the diagnostic "GID-192: registering a flag outside package main is forbidden. Fix: declare flags in the binary, let libraries take parameters" is reported

  Scenario: positive — a camelCase flag name in main
    Given the package "main" with a call to "flag.String(\"maxRetries\", ...)"
    When the analyzer checks the file
    Then the diagnostic "GID-192: flag name \"maxRetries\". Fix: use snake_case" is reported

  Scenario: positive — a hyphenated flag name in main
    Given the package "main" with a call to "flag.Int(\"max-retries\", ...)"
    When the analyzer checks the file
    Then the diagnostic "GID-192: flag name \"max-retries\". Fix: use snake_case" is reported

  # === Class 2: negative (clean code) ===

  Scenario: negative — a snake_case flag name in main
    Given the package "main" with a call to "flag.String(\"max_retries\", ...)"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — flag.Parse in main
    Given the package "main" with a call to "flag.Parse()"
    When the analyzer checks the file
    Then no diagnostic is reported

  # === Class 3: boundary cases ===

  Scenario: boundary — an own package named flag (not stdlib)
    Given a local package "flag" with a different import path
    And a call "flag.String(\"maxRetries\")" to that package
    When the analyzer checks the file
    Then no diagnostic is reported
    # The import path is not "flag" (stdlib) → TypesInfo does not attribute the call to the flag package.

  Scenario: boundary — the flag name is not a constant (in main)
    Given the package "main" with a call to "flag.String(dyn, ...)" where dyn is a variable
    When the analyzer checks the file
    Then no diagnostic is reported
    # The name is unknown statically → part 2 (snake_case) is not checked.

  Scenario: boundary — the flag name is not a constant but in a library
    Given a library package with a call to "flag.String(name, ...)" where name is a parameter
    When the analyzer checks the file
    Then the diagnostic "GID-192: registering a flag outside package main is forbidden. Fix: declare flags in the binary, let libraries take parameters" is reported
    # Part 1 (registration outside main) does not depend on the name being constant.

  Scenario: boundary — flag in a *_test.go file
    Given a "*_test.go" file in a non-main package with "flag.Bool(\"updateGolden\", ...)"
    When the analyzer checks the file
    Then no diagnostic is reported
    # Test files are skipped — flag in tests is legitimate.

  # === Class 4: non-applicability ===

  Scenario: non-applicability — a library package without flag
    Given a library package that does not import flag
    When the analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-192)
#  [x] Layer chosen: go/analysis (analyzer gidflagmain in analyzers/flagmain)
#  [x] Messages are defined ("GID-192: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml
