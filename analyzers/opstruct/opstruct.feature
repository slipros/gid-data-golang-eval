# language: en
# Spec of rule GID-210 (op-struct-fields, linter gidopstruct).

Feature: GID-210 — operational Create structs contain a minimal set of fields
  As a developer
  I want Create structs not to drag in generated fields
  So that ID/CreatedAt/UpdatedAt are set at their own layer (service/convert/DB)

  # --- Positive class: violations ---

  Scenario: a model Create with ID/CreatedAt/UpdatedAt — violation
    Given the type "CreateJob" with the fields "ID", "CreatedAt", "UpdatedAt" in /domain/model
    When the analyzer checks the file
    Then a "GID-210" diagnostic is reported on each of the fields "ID", "CreatedAt", "UpdatedAt"

  Scenario: an entity Create with UpdatedAt — violation
    Given the type "CreateJob" with the field "UpdatedAt" in /dal/entity
    When the analyzer checks the file
    Then a "GID-210" diagnostic is reported on the field "UpdatedAt"

  # --- Negative class: clean code passes ---

  Scenario: a clean model Create struct — ok
    Given the type "CreateJob" with the fields "Title", "Status" in /domain/model
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: an entity Create with ID and CreatedAt but without UpdatedAt — ok
    Given the type "CreateJob" with the fields "ID", "CreatedAt", "Title" in /dal/entity
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: an ordinary non-operational struct with ID/CreatedAt — ok
    Given the type "Snapshot" with the fields "ID", "CreatedAt" in /domain/model
    When the analyzer checks the file
    Then no diagnostic is reported

  # --- Boundary class ---

  Scenario: the field CreatedBy is not confused with CreatedAt
    Given the type "CreateJob" with the field "CreatedBy" in /domain/model
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: the type CreatedSnapshot does not match ^Create[A-Z]
    Given the type "CreatedSnapshot" with the field "ID" in /domain/model
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: Update structs are not affected
    Given the type "UpdateJob" with the fields "ID", "UpdatedAt" in /domain/model
    When the analyzer checks the file
    Then no diagnostic is reported

  # --- Non-applicability ---

  Scenario: a Create struct outside model/entity — the rule does not apply
    Given the type "CreateJob" with the fields "ID", "CreatedAt", "UpdatedAt" in /client or /event/dto
    When the analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md)
#  [x] Layer chosen: go/analysis (complex — depends on the package layer and the field set)
#  [x] Severity and message are defined
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml
