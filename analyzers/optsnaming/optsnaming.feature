# language: en

Feature: GID-126 — the Options pattern: type name and defaults (optsnaming)
  As a developer
  I want the options type to be named with an entity prefix (JobOptions),
  and the defaults to live in a Default<X>Options variable
  So that the Options pattern is uniform; in the app layer bare Options is the norm (composition)

  # Analyzer gidoptsnaming, default LoadMode (TypesInfo).
  # The app layer is detected by the path segment "app" via internal/pathseg.
  # Generated code (ast.IsGenerated) is skipped.
  #
  # Checks:
  #   1. A struct type named EXACTLY Options outside the app layer.
  #      In the app layer bare Options is the norm (it aggregates GRPCOptions/KafkaOptions), FINDINGS.md §2.3.
  #      Non-struct types (an alias to Options, an interface) are not affected.
  #   2. A package-level var of type <X>Options (including a pointer) whose name does not start with Default.
  #      Local variables are not affected. A var in the app layer is also checked.
  #
  # The neighboring rule GID-152 (gidoptsstyle) checks pointer/embedding of opts —
  # not duplicated here.

  # === Class 1: positive (the violation is caught) ===

  Scenario: positive — a struct Options outside the app layer
    Given a package in "/domain/service" with the type "type Options struct { Retries int }"
    When the analyzer checks the file
    Then the diagnostic "GID-126: an options type must have an entity prefix. Fix: use JobOptions, not bare Options" is reported on the type "Options"

  Scenario: positive — a package-level var of type JobOptions without the Default prefix
    Given a package in "/domain/service" with "var Opts = JobOptions{}"
    When the analyzer checks the file
    Then the diagnostic "GID-126: option defaults must be a Default<X>Options variable. Fix: rename it" is reported on the variable "Opts"

  Scenario: positive — a package-level var with the explicit type JobOptions without Default
    Given a package in "/domain/service" with "var defaults JobOptions"
    When the analyzer checks the file
    Then the diagnostic "GID-126: option defaults must be a Default<X>Options variable. Fix: rename it" is reported on the variable "defaults"

  Scenario: positive — a default without the Default prefix is checked in the app layer too
    Given a package in "/internal/app/config" with "var kafkaOpts = KafkaOptions{}"
    When the analyzer checks the file
    Then the diagnostic "GID-126: option defaults must be a Default<X>Options variable. Fix: rename it" is reported on the variable "kafkaOpts"

  # === Class 2: negative (clean code passes) ===

  Scenario: negative — a type with the entity prefix JobOptions
    Given a package in "/domain/service" with the type "type JobOptions struct { Retries int }"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — defaults in a Default<X>Options variable
    Given a package in "/domain/service" with "var DefaultJobOptions = JobOptions{}"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — bare Options in the app layer (composition of GRPCOptions/KafkaOptions)
    Given a package in "/internal/app/config" with the type "type Options struct { GRPC GRPCOptions; Kafka KafkaOptions }"
    When the analyzer checks the file
    Then no diagnostic is reported

  # === Class 3: boundary (similar but not flagged) ===

  Scenario: boundary — a local variable opts is not matched
    Given a function with the local "opts := JobOptions{}"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — a pointer var with the Default prefix — ok
    Given a package in "/domain/service" with "var DefaultGRPCOptions *GRPCOptions"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — a function with an opts parameter — not this rule's domain
    Given the function "func New(ctx context.Context, opts *JobOptions) int"
    When the analyzer checks the file
    Then no diagnostic is reported
    # (opts parameters/fields are checked by GID-152, not GID-126.)

  Scenario: boundary — an alias to Options and an interface are not affected
    Given a package in "/domain/model" with "type Options = entOptions" and "type OptionsProvider interface { ... }"
    When the analyzer checks the file
    Then no diagnostic is reported
    # (Check 1 applies only to struct types named exactly Options.)

  # === Class 4: non-applicability ===

  Scenario: non-applicability — a package without Options types
    Given a package in "/domain/model" with the type "type Job struct { ID int }" and "var DefaultJob = Job{}"
    When the analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-126)
#  [x] Layer chosen: go/analysis (analyzer gidoptsnaming in analyzers/optsnaming)
#  [x] Severity and message are defined ("GID-126: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml
