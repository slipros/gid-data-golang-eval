# language: en

Feature: GID-267 — a service reaches a client through a repository
  As a service architect
  I want /domain/service not to import /client/**
  So that the chain stays client -> repository -> service and the domain never
  speaks another system's models

  # Positive: the direct path service -> client

  Scenario: service imports the client layer — violation
    Given the module "svc" owns a data layer
    And the package "svc/domain/service" imports "svc/client/billing"
    When the analyzer checks the file
    Then a "GID-267" diagnostic is reported on the import "svc/client/billing"

  # Negative: the canonical chain

  Scenario: repository imports the client layer — ok
    Given the package "svc/dal/repository" imports "svc/client/billing"
    When the analyzer checks the file
    Then no "GID-267" diagnostic is reported

  Scenario: service imports dal/entity and domain/model — ok
    Given the package "svc/domain/service" imports "svc/dal/entity" and "svc/domain/model"
    When the analyzer checks the file
    Then no "GID-267" diagnostic is reported

  # Boundary: what looks like the client layer but is not

  Scenario: service imports a nested client segment below another layer — ok
    Given the package "svc/domain/service" imports "svc/connect/client/interceptor"
    When the analyzer checks the file
    Then no diagnostic is reported on that import

  Scenario: service imports a third-party package with a client segment — the rule does not apply
    Given the package "svc/domain/service" imports a package outside its own module
    When the analyzer checks the file
    Then no "GID-267" diagnostic is reported

  # Non-applicability: a BFF, and test files

  Scenario: a BFF service imports the client layer — the rule does not apply
    Given the module "bff" owns no data layer (no /dal and no /repository)
    And the package "bff/domain/service" imports "bff/client/billing"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a _test.go file of the service imports the client layer — not judged
    Given the file "svc/domain/service/service_test.go" imports "svc/client/billing"
    When the analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md)
#  [x] Layer chosen: go/analysis (import-path segments + the module layout are needed)
#  [x] Severity and message are defined
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml (rides in the existing linter gidlayerimports)
