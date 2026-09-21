# language: en
Feature: GID-277 — substantial protobuf conversion lives in the boundary-owned convert package
  As a service developer
  I want gRPC protobuf/model mapping in /server/grpc/service/handler/convert
  And event protobuf/model mapping in /event/kafka/producer/convert
  So that gRPC handlers and event adapters keep boundary ownership explicit

  Scenario: substantial protobuf-to-model mapping in a gRPC handler — violation
    Given a helper under /server/grpc/service/handler accepts a generated protobuf request
    And it returns a domain model by constructing a value with multiple fields
    When the analyzer checks the package
    Then a "GID-277" diagnostic is reported on the helper

  Scenario: substantial gRPC mapping in handler/convert — ok
    Given the same protobuf-to-model helper is in /server/grpc/service/handler/convert
    When the analyzer checks the package
    Then no diagnostic is reported

  Scenario: substantial gRPC mapping in another convert package — violation
    Given the same protobuf-to-model helper is in /server/grpc/service/convert
    Or it is in a convert descendant below /server/grpc/service/handler/convert
    When the analyzer checks the package
    Then a "GID-277" diagnostic directs it to /server/grpc/service/handler/convert

  Scenario: substantial event mapping in producer/convert — ok
    Given a model-to-protobuf helper is in /event/kafka/producer/convert
    When the analyzer checks the package
    Then no diagnostic is reported

  Scenario: substantial event mapping in another convert package — violation
    Given the same model-to-protobuf helper is in /event/kafka/convert
    Or it is in a convert descendant below /event/kafka/producer/convert
    When the analyzer checks the package
    Then a "GID-277" diagnostic directs it to /event/kafka/producer/convert

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
    Given a cross-representation fixture is in a _test.go file or outside /server/grpc/service and event packages
    When the analyzer checks the package
    Then no diagnostic is reported
