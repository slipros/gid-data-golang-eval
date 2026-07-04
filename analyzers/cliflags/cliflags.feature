# language: en

Feature: GID-238/GID-239 — urfave/cli/v3 flag literal hygiene (cli-flag-naming, cli-flag-required)
  As a developer
  I want flag literals of urfave/cli/v3 (StringFlag, IntFlag, BoolFlag, ...,
  any project *Flag type from the cli package) to have consistently cased
  names and env vars, and to always carry either Required or a default Value
  So that flags wired into the app cannot be silently forgotten as optional
  and env var lookups do not silently miss because of a casing mismatch

  # Applicability: a composite literal (bare or "&cli.XxxFlag{...}") whose
  # type is a named/aliased type — including the generic FlagBase an
  # alias like StringFlag resolves to — from a package path ending in
  # "urfave/cli/v3" or "urfave/cli", with a type name ending in "Flag".
  # Only keyed fields are inspected.

  # --- GID-238 (cli-flag-naming) ---

  Scenario: positive — Name in camelCase
    Given a flag literal 'cli.StringFlag{Name: "myFlag"}'
    When the analyzer checks the file
    Then a "GID-238" diagnostic is reported on the Name value

  Scenario: positive — Name in snake_case
    Given a flag literal 'cli.StringFlag{Name: "my_flag"}'
    When the analyzer checks the file
    Then a "GID-238" diagnostic is reported on the Name value

  Scenario: positive — EnvVars entry not UPPER_SNAKE_CASE (v2-style field)
    Given a flag literal 'cli.StringFlag{Name: "db-url", EnvVars: []string{"db-url"}}'
    When the analyzer checks the file
    Then a "GID-238" diagnostic is reported on the "db-url" env var element

  Scenario: positive — EnvVars entry in camelCase (v2-style field)
    Given a flag literal 'cli.StringFlag{Name: "db-url", EnvVars: []string{"dbUrl"}}'
    When the analyzer checks the file
    Then a "GID-238" diagnostic is reported on the "dbUrl" env var element

  Scenario: positive — cli.EnvVars(...) argument not UPPER_SNAKE_CASE (v3-style Sources)
    Given a flag literal 'cli.StringFlag{Name: "db-url", Sources: cli.EnvVars("db-url")}'
    When the analyzer checks the file
    Then a "GID-238" diagnostic is reported on the "db-url" argument

  Scenario: negative — Name is kebab-case and env vars are UPPER_SNAKE_CASE
    Given a flag literal 'cli.StringFlag{Name: "db-url", EnvVars: []string{"DB_URL"}}'
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — Sources: cli.EnvVars(...) with a correctly-cased argument
    Given a flag literal 'cli.StringFlag{Name: "db-url", Sources: cli.EnvVars("DB_URL")}'
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — single-word Name and env var need no separator
    Given a flag literal 'cli.IntFlag{Name: "port", EnvVars: []string{"PORT"}}'
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — struct with a Name field that is not a cli flag type
    Given a struct 'type Config struct{ Name string }' and a literal 'Config{Name: "my_config"}'
    When the analyzer checks the file
    Then no diagnostic is reported, because Config is not declared in the cli package

  # --- GID-239 (cli-flag-required) ---

  Scenario: positive — neither Required nor Value is set
    Given a flag literal 'cli.StringFlag{Name: "db-host"}'
    When the analyzer checks the file
    Then a "GID-239" diagnostic is reported naming the flag "db-host"

  Scenario: negative — Required: true is set
    Given a flag literal 'cli.StringFlag{Name: "db-host", Required: true}'
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — a default Value is set
    Given a flag literal 'cli.StringFlag{Name: "db-host", Value: "localhost"}'
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — an explicit zero Value still counts as a deliberate default
    Given a flag literal 'cli.IntFlag{Name: "retry-count", Value: 0}'
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — Name is not a string literal, the flag name in the message falls back to a placeholder
    Given a flag literal 'cli.StringFlag{Name: name}' where name is a function parameter
    When the analyzer checks the file
    Then a "GID-239" diagnostic is reported naming the flag "<flag>"

  Scenario: non-applicability — flag name listed in settings.exclude
    Given "legacy-mode" is listed in settings.exclude
    And a flag literal 'cli.BoolFlag{Name: "legacy-mode"}' with neither Required nor Value
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — struct with a Name field that is not a cli flag type
    Given a struct 'type Config struct{ Name string }' and a literal 'Config{Name: "my_config"}'
    When the analyzer checks the file
    Then no GID-239 diagnostic is reported, because Config is not declared in the cli package

# --- Checklist for adding a new rule ---
#  [x] ID and description recorded (wired by the parent task into RULES.md)
#  [x] Layer chosen: go/analysis (types needed — flag type identity, literal field values)
#  [x] Severity and message defined
#  [x] Case classes covered: positive, negative, boundary, non-applicability (both rules)
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml (wired by the parent task)
