# language: en

Feature: GID-168 — ban on db tags on struct fields in /domain/**
  As a developer
  I want struct fields in /domain/** to have no database-mapping tags
  So that the model stays a pure business object and the DB column mapping lives in entity (DAL)

  # Scope: pathseg.Contains(pkgPath, "domain") — the model layer and all its subpackages.
  # The mapping tag defaults to ["db"], configurable via Settings.Tags.
  # Other tags (json etc.) are not touched. Generated code is skipped.

  Scenario: positive — a field with a db tag in domain
    Given a package in "/domain/model" with the struct "Snapshot" having the field "ID string `db:\"id\"`"
    When the analyzer checks the file
    Then the diagnostic "GID-168: field Snapshot.ID has a \"db\" tag in the domain layer. Fix: keep db mapping in /dal/entity" is reported on the field "ID"

  Scenario: positive — a private field with a db tag is also flagged
    Given a package in "/domain/model" with the struct "cursor" having the private field "offset int `db:\"offset\"`"
    When the analyzer checks the file
    Then a "GID-168" diagnostic is reported on the field "offset"

  Scenario: negative — fields without mapping tags
    Given a package in "/domain/model" with the struct "Job" whose fields have only json/validate tags or no tags at all
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — an embedded field with a db tag is flagged
    Given a package in "/domain/model" with a struct embedding "Snapshot `db:\"snapshot\"`"
    When the analyzer checks the file
    Then a "GID-168" diagnostic is reported on the embedded field "Snapshot"

  Scenario: boundary — a ch tag is not flagged with default settings
    Given Settings.Tags is not set (the default ["db"] applies)
    And a package in "/domain/model" with the struct "Metric" having the field "ID string `ch:\"id\"`"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — the same struct in /dal/entity
    Given a package in "/dal/entity" with the struct "Snapshot" having fields with db tags
    When the analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-168)
#  [x] Layer chosen: go/analysis (analyzer gidmodeltags in analyzers/dbtags)
#  [x] Severity and message are defined ("GID-168: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml
