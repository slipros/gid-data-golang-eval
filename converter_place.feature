# language: en
Feature: GID-277 — substantial cross-representation conversion lives in convert
  As a service developer
  I want protobuf/model field mapping centralized in leaf convert packages
  So that gRPC and event adapters keep boundary ownership explicit

  Scenario: substantial protobuf-to-model mapping in a gRPC handler — violation
    Given a handler helper accepts a generated protobuf request
    And it returns a domain model by constructing a value with multiple fields
    When the analyzer checks the package
    Then a "GID-277" diagnostic is reported on the helper

  Scenario: substantial mapping in a leaf convert package — ok
    Given the same protobuf-to-model helper is in a leaf package named "convert"
    When the analyzer checks the package
    Then no diagnostic is reported

  Scenario: complete delegation or local one-field policy — boundary
    Given a handler returns the result of a convert function without local field mapping
    Or it maps a domain result into a one-field protobuf response wrapper
    When the analyzer checks the package
    Then no diagnostic is reported

  Scenario: a convert call does not hide local mapping — violation
    Given a handler calls a convert function but still constructs a multi-field result locally
    When the analyzer checks the package
    Then a "GID-277" diagnostic is reported on the handler helper

  Scenario: test fixture or unrelated package — the rule does not apply
    Given a cross-representation fixture is in a _test.go file or outside gRPC and event layers
    When the analyzer checks the package
    Then no diagnostic is reported
