# language: en

Feature: GID-172 — client does not import the dal layer
  As a service architect
  I want /client/** packages not to import /dal/**
  So that the client has its own types and knows nothing about entity/repository

  Scenario: client imports dal/entity — violation
    Given the package "svc/client/snapshot" imports "svc/dal/entity"
    When the analyzer checks the file
    Then a "GID-172" diagnostic is reported on the import "svc/dal/entity"

  Scenario: client imports dal/repository — violation
    Given the package "svc/client/snapshot" imports "svc/dal/repository"
    When the analyzer checks the file
    Then a "GID-172" diagnostic is reported on the import "svc/dal/repository"

  Scenario: client imports domain/model — ok
    Given the package "svc/client/billing" imports "svc/domain/model"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: client imports a third-party package — the rule does not apply
    Given the package "svc/client/billing" imports "strconv"
    When the analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the PRD registry (section 5)
#  [x] Layer chosen: go/analysis (import-path segments are needed)
#  [x] Severity and message are defined
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml
