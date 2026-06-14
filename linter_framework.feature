# language: en
Feature: Custom linter framework (custom-gcl)
  As a developer
  I want the linter to deterministically block rule violations
  So that compliance with internal code rules is guaranteed

  # Implementation note: in Go these scenarios map naturally onto
  # golang.org/x/tools/go/analysis/analysistest + testdata with // want comments.
  # Gherkin here is a specification layer, not a way to run tests.

  Background:
    Given the "custom-gcl" binary is built with all analyzers
    And the repository contains the reference ".golangci.yml"

  Scenario: Clean code passes the check
    Given a Go file violates no rule
    When I run "custom-gcl run"
    Then the exit code equals 0
    And no diagnostic message is printed

  Scenario: A rule violation blocks
    Given a Go file violates rule "RULE-001"
    When I run "custom-gcl run"
    Then the exit code is not 0
    And a diagnostic for "RULE-001" is printed with the file and line

  Scenario: Multiple violations in a single run
    Given a Go file violates rules "RULE-001" and "RULE-002"
    When I run "custom-gcl run"
    Then the exit code is not 0
    And separate diagnostics are printed for "RULE-001" and for "RULE-002"

  Scenario: A rule can be disabled in the config
    Given linter "RULE-003" is disabled in ".golangci.yml"
    And a Go file violates rule "RULE-003"
    When I run "custom-gcl run"
    Then no diagnostic for "RULE-003" is printed

  Scenario: ruleguard and custom analyzers work in a single run
    Given a Go file violates a ruleguard rule and a go/analysis rule
    When I run "custom-gcl run"
    Then diagnostics of both kinds are printed
    And the exit code is not 0

  Scenario: CI gate blocks merge
    Given a PR contains a file violating a covered rule
    When CI runs "custom-gcl run" as a required check
    Then the check fails
    And the merge is blocked
