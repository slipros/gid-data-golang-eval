# language: en

Feature: GID-169 — layer errors live in a dedicated file
  As a developer
  I want package-level error variables in /domain/model and /dal/entity
  to be declared only in a dedicated file (error.go/errors.go)
  So that layer errors are collected in one place and are easy to find

  # Refines GID-144 (domain-errors-in-model) and GID-145 (dal-errors-in-entity):
  # those define the home package for layer errors, while GID-169 defines the
  # specific file inside that home. Scope: only the root packages of the layer
  # (the path ends with domain/model or dal/entity); subpackages (model/filter
  # etc.) are not affected.

  Scenario: positive — an error var outside error.go in /domain/model
    Given the package path ends with the segments "domain/model"
    And "var ErrSnapshotConflict = errors.New(...)" is declared in the file "snapshot.go"
    When the analyzer checks the file
    Then a "GID-169" diagnostic is reported on the variable "ErrSnapshotConflict"

  Scenario: positive — an error var outside errors.go in /dal/entity
    Given the package path ends with the segments "dal/entity"
    And "var ErrRowLocked = errors.New(...)" is declared in the file "row.go"
    When the analyzer checks the file
    Then a "GID-169" diagnostic is reported on the variable "ErrRowLocked"

  Scenario: negative — an error var in an allowed file
    Given the package path ends with the segments "domain/model"
    And "var ErrSnapshotNotFound = errors.New(...)" is declared in the file "errors.go"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — an error var in errors.go in /dal/entity
    Given the package path ends with the segments "dal/entity"
    And "ErrRowNotFound" and "ErrDuplicateKey" are declared in the file "errors.go"
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

  Scenario: non-applicability — the /domain/service package
    Given the package path ends with the segments "domain/service"
    And "var ErrServiceLocal = errors.New(...)" is declared in the file "service.go"
    When the analyzer checks the file
    Then no GID-169 diagnostic is reported
    # (declaring an error here is the domain of GID-144, not GID-169)

  Scenario: non-applicability — the /domain/model/filter subpackage
    Given the package path ends with the segments "domain/model/filter"
    And "var ErrBadFilter = errors.New(...)" is declared in the file "filter.go"
    When the analyzer checks the file
    Then no GID-169 diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (GID-169, errors-in-error-file)
#  [x] Layer chosen: go/analysis (a type is needed — implementsError)
#  [x] Severity and message are defined ("GID-169: error %q is declared in %s — ...")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml (outside the scope of the task)
