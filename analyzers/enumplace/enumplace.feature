# language: en
# Eval spec of rule GID-211 (enum-location), linter gidenumplace.

Feature: GID-211 — DAL-layer enums live in /dal/entity/enum
  As a DAL-layer developer
  I want string enums to be declared only in /dal/entity/enum
  So that each enum lives in a separate file named after the entity (entity.md)

  Scenario: a string enum in /dal/entity (the root) — violation
    Given "type Status string" with consts of this type is declared in the /dal/entity package
    When the analyzer checks the file
    Then a "GID-211" diagnostic is reported on the type name "Status"

  Scenario: a string enum in /dal/repository — violation
    Given "type Mode string" with consts of this type is declared in the /dal/repository package
    When the analyzer checks the file
    Then a "GID-211" diagnostic is reported on the type name "Mode"

  Scenario: a string enum in /dal/entity/enum — ok
    Given "type Status string" with consts of this type is declared in the /dal/entity/enum package
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a string type without consts in DAL — not an enum
    Given "type RawJSON string" without consts of this type is declared in the /dal/entity package
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: an alias to string with consts — the domain of GID-123, not GID-211
    Given "type Code = string" with consts is declared in the /dal/entity package
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a named int type with consts — not a string enum
    Given "type Priority int" with consts of this type is declared in the /dal/entity package
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a string enum in /domain/model — the rule does not apply
    Given "type Status string" with consts of this type is declared in the /domain/model package
    When the analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md)
#  [x] Layer chosen: go/analysis (type information needed)
#  [x] Severity and message are defined
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
