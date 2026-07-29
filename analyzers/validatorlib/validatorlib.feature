# language: en

Feature: GID-164 — incoming data is validated with validator.go (gidvalidator)
  As a developer
  I want every request and event validated by one library
  So that rules, messages and i18n come from one place instead of three
  validation dialects meeting in one service

  # One analyzer over the file's import declarations, no type info needed.
  # Two independent checks:
  #   1. everywhere — a third-party validation library is forbidden:
  #      go-playground/validator, ozzo-validation, govalidator (matched by
  #      path prefix, so any major version counts);
  #   2. in a validate package — the package path ending with the "validate"
  #      segment (pathseg.EndsWith), i.e. server/*/handler/validate and
  #      kafka/consumer/validate — github.com/raoptimus/validator.go/v2 must be
  #      imported. One diagnostic per package is enough.
  # Exceptions: //nolint:gidvalidator pointwise; settings.exclude centrally —
  # a full import path or a trailing segment suffix.
  # Generated code (ast.IsGenerated) is skipped.

  # --- Class 1: positive ---

  Scenario: positive — a third-party validator imported by a service
    Given "validator \"github.com/go-playground/validator/v10\"" imported in /domain/service
    When the gidvalidator analyzer checks the file
    Then the diagnostic "GID-164: third-party validation library \"github.com/go-playground/validator/v10\" is forbidden. Fix: use github.com/raoptimus/validator.go/v2" is reported on the import

  Scenario: positive — a validate package without the library
    Given the package "svc/server/http/handler/validate" importing no validator
    When the gidvalidator analyzer checks the package
    Then the diagnostic "GID-164: validate package \"svc/server/http/handler/validate\" must use github.com/raoptimus/validator.go/v2. Fix: import it (exceptions: nolint or settings.exclude)" is reported on the package clause

  Scenario: positive — the same in a grpc validate package
    Given the package "svc/server/grpc/handler/validate" importing no validator
    When the gidvalidator analyzer checks the package
    Then the diagnostic "GID-164: validate package \"svc/server/grpc/handler/validate\" must use …" is reported

  # --- Class 2: negative ---

  Scenario: negative — a validate package on the library
    Given the package "svc/server/http/handler/validate" importing "github.com/raoptimus/validator.go/v2"
    When the gidvalidator analyzer checks the package
    Then no diagnostic is reported

  Scenario: negative — a package that validates nothing
    Given "svc/pkg/util" importing neither a validator nor the library
    When the gidvalidator analyzer checks the package
    Then no diagnostic is reported
    # The requirement to import the library applies to validate packages.

  # --- Class 3: boundary ---

  Scenario: boundary — a validate package listed in settings.exclude
    Given settings "exclude: [excluded/kafka/consumer/validate]" and that package without the library
    When the gidvalidator analyzer checks the package
    Then no diagnostic is reported

  Scenario: boundary — exclude by a trailing segment suffix
    Given settings "exclude: [server/http/handler/validate]" and the package "excluded/server/http/handler/validate"
    When the gidvalidator analyzer checks the package
    Then no diagnostic is reported
    # The exclusion matches a full path or a suffix of path segments.

  Scenario: boundary — an older major of the library
    Given a validate package importing "github.com/raoptimus/validator.go" without /v2
    When the gidvalidator analyzer checks the package
    Then the diagnostic "GID-164: validate package … must use github.com/raoptimus/validator.go/v2. …" is reported
    # The required version is part of the rule: v2 is the path that counts.

  Scenario: boundary — a subpackage of the library
    Given a validate package importing "github.com/raoptimus/validator.go/v2/rule"
    When the gidvalidator analyzer checks the package
    Then no diagnostic is reported
    # Any package of the library satisfies the import check.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — a package whose name merely contains "validate"
    Given the package "svc/pkg/validatehelpers"
    When the gidvalidator analyzer checks the package
    Then no "must use" diagnostic is reported
    # The check is on the last path segment, not on a substring.

  Scenario: non-applicability — generated code
    Given a file carrying the "Code generated … DO NOT EDIT." header
    When the gidvalidator analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-164)
#  [x] Layer chosen: go/analysis (package validatorlib)
#  [x] Message is defined ("GID-164: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
