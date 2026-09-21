# language: en

Feature: GID-278 — transport validation stays at ingress (gidvalidationplace)
  As a developer
  I want request and command shape validation owned by transport validate packages
  So that domain models and services receive validated typed input instead of depending on transport contracts

  # The rule intentionally enforces a structural boundary rather than guessing
  # whether an arbitrary branch is transport policy or a business invariant.
  # A transport-validation declaration is an error-returning method named
  # Validate on a /domain/model type whose name ends in Request or Command, or
  # on any type under /domain/model/request. Domain service and usecase calls
  # are intentionally broader: they must not call an error-returning Validate
  # method on any same-module domain model, regardless of its type name. This
  # catches misplaced input validation without relying on naming conventions.
  #
  # Scope uses pathseg.HasLayer, covering internal/domain/... and
  # pkg/<module>/domain/... while staying anchored to the module root. Calls on
  # another module's model are not judged: this repository cannot move that
  # dependency's validation implementation. Generated and _test.go files are
  # skipped. A targeted //nolint:gidvalidationplace remains the escape hatch.

  # --- Class 1: positive ---

  Scenario: positive — a request validates itself in domain model
    Given ApproveWorkOrderRequest in /domain/model with "Validate() error"
    When the gidvalidationplace analyzer checks the package
    Then a GID-278 diagnostic is reported on Validate
    And the fix moves validation to the ingress validate package

  Scenario: positive — a command validates itself with context
    Given CompleteWorkOrderCommand in /domain/model with "Validate(context.Context) error"
    When the gidvalidationplace analyzer checks the package
    Then a GID-278 diagnostic is reported on Validate

  Scenario: positive — domain service calls model validation
    Given /domain/service calls ApproveWorkOrderRequest.Validate
    When the gidvalidationplace analyzer checks the package
    Then a GID-278 diagnostic says transport requests and commands must be validated at ingress

  Scenario: positive — domain usecase calls model validation
    Given /domain/usecase calls ApproveWorkOrderRequest.Validate
    When the gidvalidationplace analyzer checks the package
    Then the same ingress diagnostic is reported

  Scenario: positive — domain service validation does not depend on the type name
    Given Order in /domain/model has an error-returning Validate method
    And /domain/service calls Order.Validate
    When the gidvalidationplace analyzer checks the service
    Then a GID-278 diagnostic requires validation at ingress
    But the model declaration itself is not reported as transport validation

  Scenario: positive — request package marks transport ownership
    Given Update in /domain/model/request has an error-returning Validate method
    When the gidvalidationplace analyzer checks the model and service packages
    Then the declaration and the domain-service call are both reported

  Scenario: positive — pkg module layout
    Given RefundRequest and its service live under pkg/billing/domain
    When the gidvalidationplace analyzer checks the packages
    Then the declaration and service call are both reported

  # --- Class 2: negative ---

  Scenario: negative — business entity invariant declaration
    Given Order in /domain/model has "Validate() error"
    When the gidvalidationplace analyzer checks only the model declaration
    Then no diagnostic is reported because Order is not structurally a transport model

  Scenario: negative — a differently named business method
    Given ListRequest in /domain/model has "ValidateState() error"
    When the gidvalidationplace analyzer checks the package
    Then no diagnostic is reported

  Scenario: negative — Validate does not implement the error contract
    Given ReadyCommand in /domain/model has "Validate() bool"
    When the gidvalidationplace analyzer checks the package
    Then no diagnostic is reported

  Scenario: negative — canonical ingress validator
    Given a validator in /server/grpc/handler/validate has "Validate(context.Context, *ApproveWorkOrderRequest) error"
    When the gidvalidationplace analyzer checks the package
    Then no diagnostic is reported

  # --- Class 3: boundary ---

  Scenario: boundary — pointer and value receivers
    Given one request uses a pointer receiver and one command uses a value receiver
    When the gidvalidationplace analyzer checks the package
    Then both are reported

  Scenario: boundary — method arguments do not bypass placement
    Given CompleteWorkOrderCommand.Validate accepts context.Context
    When the gidvalidationplace analyzer checks its declaration and call
    Then both are reported because ownership and the error result define the violation

  Scenario: boundary — error may be one of several results
    Given ParsedRequest has "Validate() (bool, error)"
    When the gidvalidationplace analyzer checks its declaration and a service call
    Then both are reported because any error result establishes the validation contract

  Scenario: boundary — another module owns its model
    Given /domain/service calls Validate on ExternalRequest from another module
    When the gidvalidationplace analyzer checks the service
    Then no diagnostic is reported

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — transport layer call
    Given /server/grpc/handler calls a domain model's Validate method
    When the gidvalidationplace analyzer checks the handler
    Then no call diagnostic is reported because only domain service and usecase calls are judged

  Scenario: non-applicability — a _test.go file
    Given a test-only request type in /domain/model has Validate
    When the gidvalidationplace analyzer checks the package
    Then no diagnostic is reported

  Scenario: non-applicability — generated code
    Given a generated request type in /domain/model has Validate
    When the gidvalidationplace analyzer checks the package
    Then no diagnostic is reported
