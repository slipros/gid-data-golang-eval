# language: en

Feature: GID-169 — package errors live in a dedicated file
  As a developer
  I want package-level error variables and concrete error types in every production package
  to be declared only in a dedicated file (error.go)
  So that package errors are collected in one place and are easy to find

  # GID-144 and GID-145 define error ownership by layer. GID-169 independently
  # defines the declaration file in every production package, including client
  # packages and subpackages such as domain/model/filter.

  Scenario: positive — an error var outside error.go in /domain/model
    Given the package path ends with the segments "domain/model"
    And "var ErrSnapshotConflict = errors.New(...)" is declared in the file "snapshot.go"
    When the analyzer checks the file
    Then a "GID-169" diagnostic is reported on the variable "ErrSnapshotConflict"

  Scenario: positive — an error var outside error.go in /dal/entity
    Given the package path ends with the segments "dal/entity"
    And "var ErrRowLocked = errors.New(...)" is declared in the file "row.go"
    When the analyzer checks the file
    Then a "GID-169" diagnostic is reported on the variable "ErrRowLocked"

  Scenario: positive — errors.go is no longer allowed by default in /domain/model
    Given the package path ends with the segments "domain/model"
    And "var ErrSnapshotNotFound = errors.New(...)" is declared in the file "errors.go"
    When the analyzer checks the file with the default settings.files
    Then a "GID-169" diagnostic is reported on the variable "ErrSnapshotNotFound"

  Scenario: positive — errors.go is no longer allowed by default in /dal/entity
    Given the package path ends with the segments "dal/entity"
    And "ErrRowNotFound" and "ErrDuplicateKey" are declared in the file "errors.go"
    When the analyzer checks the file with the default settings.files
    Then a "GID-169" diagnostic is reported on each variable

  Scenario: negative — an error var in error.go (the default allowed file) in /domain/model
    Given the package path ends with the segments "domain/model"
    And "var ErrSnapshotArchived = errors.New(...)" is declared in the file "error.go"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — an error var in error.go (the default allowed file) in /dal/entity
    Given the package path ends with the segments "dal/entity"
    And "ErrRowExpired" and "ErrRowArchived" are declared in the file "error.go"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — settings.files is configurable to allow other names
    Given settings.files is customized to include "errors.go"
    And the package path ends with the segments "domain/model" or "dal/entity"
    And an error var is declared in the file "errors.go"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — a type implements error via a pointer
    Given the package path ends with the segments "domain/model"
    And the type "ValidationError" has an "Error()" method on the pointer
    And "var errValidation = &ValidationError{}" is declared in the file "snapshot.go"
    When the analyzer checks the file
    Then a "GID-169" diagnostic is reported on the variable "errValidation"

  Scenario: boundary — a value of a type with Error on the pointer does not implement error
    Given the package path ends with the segments "domain/model"
    And the type "ValidationError" has an "Error()" method on the pointer
    And "var errValidationValue = ValidationError{}" is declared in the file "snapshot.go"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — a non-error package-level variable
    Given the package path ends with the segments "domain/model"
    And "var DefaultLimit = 100" is declared in the file "snapshot.go"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — *_test.go is skipped
    Given the package path ends with the segments "domain/model"
    And "var errTestOnly = errors.New(...)" is declared in the file "snapshot_extra_test.go"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — a generated file is skipped
    Given the package path ends with the segments "domain/model"
    And the file "zz_generated.go" is marked "// Code generated ... DO NOT EDIT."
    And "var ErrGenerated = errors.New(...)" is declared in it
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: positive — the /domain/service package
    Given the package path ends with the segments "domain/service"
    And "var ErrServiceLocal = errors.New(...)" is declared in the file "service.go"
    When the analyzer checks the file
    Then a GID-169 diagnostic is reported
    # GID-144 separately checks ownership.

  Scenario: positive — the /domain/model/filter subpackage
    Given the package path ends with the segments "domain/model/filter"
    And "var ErrBadFilter = errors.New(...)" is declared in the file "filter.go"
    When the analyzer checks the file
    Then a GID-169 diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (GID-169, errors-in-error-file)
#  [x] Layer chosen: go/analysis (a type is needed — implementsError)
#  [x] Severity and message are defined ("GID-169: error %q is declared in %s — ...")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in all three configs

  Scenario: Error types and all their methods share the error file
    Given a named concrete type implements error on its value or pointer
    And its declaration or any receiver method is outside error.go
    When giderrfile analyzes any production package
    Then it reports GID-169 at each misplaced declaration
    But aliases, interfaces, non-error types, tests and generated files remain outside this check
