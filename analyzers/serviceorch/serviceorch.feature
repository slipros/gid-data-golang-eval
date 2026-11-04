# language: en
Feature: GID-260 — one repository per service
  Scenario: positive — multiple repositories
    Given a service holding two repository dependencies
    When the analyzer runs
    Then GID-260 asks to compose the entities in a usecase

  Scenario: negative — transaction on one entity
    Given a service holding one repository and model.InTransactionFunc
    When the analyzer runs
    Then no diagnostic is reported

  Scenario: boundary — transaction does not exempt multiple repositories
    Given a service holding a transaction and two repositories
    When the analyzer runs
    Then only the multiple-repository diagnostic is reported

  Scenario: boundary — custom suffix and field exclusion
    Given settings.suffixes includes Store and settings.exclude contains Type.Field
    When the analyzer runs
    Then only non-excluded repository fields count

  Scenario: non-applicability — options, usecase, generated and test code
    Given an Options struct or a usecase or generated or test file
    When the analyzer runs
    Then no diagnostic is reported
