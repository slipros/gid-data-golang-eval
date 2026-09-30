# language: en

Feature: GID-144 / GID-145 — an error TYPE lives in the layer's error home too
  As a developer
  I want a named type implementing error to be declared in /domain/model or /dal/entity
  So that the layer's errors have one home whether they are sentinels or typed values

  # Semantics (owner decision of 2026-09-30; incident: a classifiedError in a repository that
  # carried an entity sentinel and the client's cause in one Unwrap() []error chain):
  # - scope as before: GID-144 — the /domain tree outside /domain/model, GID-145 — the /dal tree
  #   outside /dal/entity (matched by path segments);
  # - trigger: a package-level named type that is not an interface and not an alias, whose pointer
  #   has the method `Error() string` (promoted from an embedded error included), generics too;
  # - the layer constructs the type, it does not declare it; the fix moves the declaration home;
  # - not judged: an interface (a contract `interface{ error; NotFound() bool }`), an alias,
  #   a function-local type, Error with another signature, _test.go files (GID-250), generated code;
  # - exclusions: //nolint:giddalerrors / //nolint:giddomainerrors.

  Scenario: positive — the incident shape in a repository
    Given the package path ends with the segments "dal/repository"
    And "type classifiedError struct" has "func (c *classifiedError) Error() string"
    When the analyzer checks the package
    Then a "GID-145" diagnostic on the type suggests declaring it in /dal/entity

  Scenario: positive — a value receiver, a non-struct type, a promoted Error, a generic type
    Given "type QueryError string", "type wrapped struct{ error }" and "type genericError[T any] struct" are declared in /dal/repository
    When the analyzer checks the package
    Then a "GID-145" diagnostic is reported on each

  Scenario: positive — service and usecase
    Given an error type is declared in "domain/service" or "domain/usecase"
    When the analyzer checks the package
    Then a "GID-144" diagnostic suggests declaring it in /domain/model

  Scenario: negative — the home layer and constructing a home type
    Given the error type is declared in "dal/entity" or "domain/model"
    And the repository or service only builds it (&entity.ClassifiedError{...})
    When the analyzer checks the package
    Then no diagnostic is reported

  Scenario: boundary — not an error value type
    Given "Error(code int) string", "Error() int", an interface embedding error, an alias of an entity error type and a function-local type are declared
    When the analyzer checks the package
    Then no diagnostic is reported

  Scenario: non-applicability — outside the layers, a nested package named like a layer, test files
    Given an error type is declared in "pkg/util", in "server/interceptor/dal", or in a _test.go file of domain/service
    When the analyzer checks the package
    Then no diagnostic is reported
