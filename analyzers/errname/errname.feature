# language: en

Feature: GID-234 — model errors are bound to their entity
  As a developer
  I want every package-level error in /domain/model to carry an entity prefix
  (Err<Entity><Reason>) instead of a generic name (ErrNotFound)
  So that errors stay concrete and can be handled precisely on the upper layers

  # Source: model.md, layer errors section. Generic names (ErrNoResult,
  # ErrAlreadyExists) are the convention of /dal/entity, so the dal layer is
  # out of scope. Scope: /domain/model — root package and subpackages
  # (matched via pathseg). The banned list is configurable via settings.names,
  # point exceptions via //nolint:giderrname, centralized via settings.exclude.

  Scenario: positive — generic ErrNotFound in /domain/model
    Given a package whose path contains the segments "domain/model"
    And the file declares "var ErrNotFound = errors.New(...)" at package level
    When the analyzer checks the file
    Then a diagnostic "GID-234" is reported on the variable "ErrNotFound"

  Scenario: positive — generic name in a /domain/model subpackage
    Given a package whose path ends with "domain/model/promo"
    And the file declares "var ErrAlreadyExists = errors.New(...)" at package level
    When the analyzer checks the file
    Then a diagnostic "GID-234" is reported on the variable "ErrAlreadyExists"

  Scenario: positive — banned name with explicit error type
    Given a package whose path contains the segments "domain/model"
    And the file declares "var ErrConflict error" at package level
    When the analyzer checks the file
    Then a diagnostic "GID-234" is reported on the variable "ErrConflict"

  Scenario: negative — entity-bound error name
    Given a package whose path contains the segments "domain/model"
    And the file declares "var ErrSnapshotNotFound = errors.New(...)" at package level
    When the analyzer checks the file
    Then no diagnostic is reported
    # ErrSnapshotNotFound merely contains "NotFound" as a suffix — that is the goal

  Scenario: boundary — non-error variable with a banned name
    Given a package whose path contains the segments "domain/model"
    And the file declares "var ErrNotFoundMessage = \"not found\"" (type string)
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — local variable inside a function
    Given a package whose path contains the segments "domain/model"
    And a function body declares "errNotFound := errors.New(...)" locally
    When the analyzer checks the file
    Then no diagnostic is reported
    # only package-level declarations are in scope

  Scenario: boundary — *_test.go and generated files are skipped
    Given a package whose path contains the segments "domain/model"
    And "model_extra_test.go" or a "// Code generated ... DO NOT EDIT." file
    declares "var ErrNotFound = errors.New(...)"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — settings.names overrides the banned list
    Given settings.names is ["ErrOops"]
    And a /domain/model package declares "ErrOops" and "ErrNotFound"
    When the analyzer checks the file
    Then a diagnostic "GID-234" is reported only on "ErrOops"

  Scenario: boundary — settings.exclude exempts a variable
    Given settings.exclude contains "ErrNotFound"
    And a /domain/model package declares "var ErrNotFound = errors.New(...)"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — /dal/entity keeps generic names
    Given a package whose path contains the segments "dal/entity"
    And the file declares "var ErrNoResult = errors.New(...)" at package level
    When the analyzer checks the file
    Then no diagnostic "GID-234" is reported
    # generic names are the convention of the dal layer

  Scenario: non-applicability — /domain/service is out of scope
    Given a package whose path contains the segments "domain/service"
    And the file declares "var ErrNotFound = errors.New(...)" at package level
    When the analyzer checks the file
    Then no diagnostic "GID-234" is reported
    # declaring errors there is the zone of GID-144, not GID-234

# --- Checklist for adding a new rule ---
#  [x] ID and description registered in RULES.md (GID-234, model-error-entity-name)
#  [x] Layer chosen: go/analysis (types needed — implementsError)
#  [x] Severity and message set ("GID-234: generic error name %q in domain model. Fix: ...")
#  [x] Cases covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml (wired by the parent)
