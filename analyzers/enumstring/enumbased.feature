# language: en

Feature: GID-123 — an enum is a named type based on string, not a bare string/int
  As a developer
  I want enums in /domain/model and /dal/entity to be named string types
  So that enumerations are type-safe and uniform
  Source: styleguide.md#enum
  Scope: packages in /domain/model/** and /dal/entity/**

  # --- Positive class: the violation is caught ---

  Scenario: an alias to basic string — violation
    Given "type ConsentEventType = string" is declared in /domain/model
    When the analyzer checks the file
    Then a "GID-123" diagnostic is reported with the text "a named type, not an alias"

  Scenario: an int enum with two or more values — violation
    Given "type Status int" with the const values "StatusA, StatusB" is declared in /dal/entity
    When the analyzer checks the file
    Then a "GID-123" diagnostic is reported with the text "must be based on string, not int"

  Scenario: a group of untyped string constants — violation
    Given a const group of two untyped string constants is declared in /domain/model
    When the analyzer checks the file
    Then a single "GID-123" diagnostic is reported on the first constant of the group

  # --- Negative class: clean code passes ---

  Scenario: a named string type with consts — ok
    Given "type EventType string" with const values is declared in /domain/model
    When the analyzer checks the file
    Then no diagnostic is reported

  # --- Boundary class ---

  Scenario: a single const of a named int type — not an enum, not flagged
    Given "type Limit int" and a single "const DefaultLimit Limit = 100" are declared in /domain/model
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a single untyped string const — ok
    Given a single "const DefaultName = \"x\"" is declared in /dal/entity
    When the analyzer checks the file
    Then no diagnostic is reported

  # --- Non-applicability class: the rule does not apply out of scope ---

  Scenario: an alias to string outside /domain/model and /dal/entity — the rule does not apply
    Given "type ConsentEventType = string" is declared in /domain/service
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a generated file — skipped
    Given the file is marked "// Code generated ... DO NOT EDIT."
    When the analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-123)
#  [x] Layer chosen: go/analysis (types needed for underlying/alias)
#  [x] Severity and message are defined ("GID-123: ...")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml (outside the scope of this task)
