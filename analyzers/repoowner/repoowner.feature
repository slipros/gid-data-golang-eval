# language: en
Feature: GID-280 — a repository has one owning service
  Scenario: positive — one repository in two services
    Given constructors of two different service types receive a concrete repository type
    When the analyzer checks their wiring package
    Then GID-280 names the repository and both services

  Scenario: positive — different instances or consumer interfaces
    Given two services receive separately constructed instances of one repository type
    And a repository is passed through a single-assignment interface variable or alias
    When the analyzer checks the wiring
    Then the same ownership conflict is reported

  Scenario: negative — repeated construction of one owning service
    Given several instances of one service use the same repository type
    When the analyzer checks the wiring
    Then no diagnostic is reported

  Scenario: boundary — getters, aliases, struct literals and tuple-returning constructors
    Given a service receives a repository through a concrete getter, type alias or keyed literal
    When the analyzer checks the wiring
    Then ownership is checked by concrete type including its package

  Scenario: non-applicability — unknown dynamic implementation
    Given an interface parameter or a reassigned interface variable without a definite concrete origin
    When the analyzer checks the wiring
    Then no speculative diagnostic is reported

  Scenario: non-applicability — tests, generated code and Options
    Given wiring in a test or generated file or an Options literal
    When the analyzer runs
    Then it does not establish repository ownership

  Scenario: boundary — package scope and targeted exceptions
    Given owners are wired in separate packages or settings.exclude names the fully qualified repository
    When the analyzer runs
    Then no cross-package ownership claim is made and excluded repositories are silent
