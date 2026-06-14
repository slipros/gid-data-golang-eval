# language: en

Feature: GID-173 — dependency interfaces are named with an entity prefix
  As a developer
  I want a dependency interface to be named with an entity prefix (HelloRepository, HelloConnection)
  So that the interface name shows which entity it is a dependency of, not just a bare role

  # Scope: packages in the layers /domain/service, /domain/usecase, /dal/repository,
  # /server/**, /event/**. The default dictionary of bare roles:
  # Repository, Service, Client, Connection, Producer, Consumer, Validator,
  # Storage, Cache. Configurable via settings.names. Generated code
  # (ast.IsGenerated) is skipped.

  Scenario: the bare role Repository in /domain/service — violation (positive)
    Given "type Repository interface { ... }" is declared in the "/domain/service" package
    When the analyzer checks the file
    Then a "GID-173" diagnostic is reported on the type name "Repository"

  Scenario: the bare role Connection in /dal/repository — violation (positive)
    Given "type Connection interface { ... }" is declared in the "/dal/repository" package
    When the analyzer checks the file
    Then a "GID-173" diagnostic is reported on the type name "Connection"

  Scenario: an interface with an entity prefix — ok (negative)
    Given "type HelloRepository interface { ... }" is declared in the "/domain/service" package
    And "type SnapshotConnection interface { ... }" is declared
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a struct type with a role name — the rule does not apply (boundary)
    Given "type Repository struct { ... }" is declared in the "/dal/repository" package
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: the name contains a role as a suffix but is not exactly equal to it — ok (boundary)
    Given "type RepositoryFactory interface { ... }" is declared in the "/domain/service" package
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a bare role out of scope — the rule does not apply (non-applicability)
    Given "type Repository interface { ... }" is declared in the "/domain/model" package
    And "type Repository interface { ... }" is declared in the "/internal/foo" package
    When the analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-173)
#  [x] Layer chosen: go/analysis (the package layer by path + type AST are needed)
#  [x] Severity and message are defined
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml
