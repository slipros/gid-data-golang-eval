# language: en

Feature: GID-213 — the validator shape (validator-shape)
  As a developer
  I want every validator to be a struct with the method
  Validate(ctx context.Context, req *T) error
  So that validators are uniform and match the operation by name

  # Scope: only packages of the validate layer (the "validate" segment in the import path).
  # Exported struct types are checked, except names with the Options suffix.
  # Sufficient: the first parameter is context.Context, the only result is error.

  Scenario: positive — an exported struct without a Validate method
    Given "type CreateJob struct{}" without a Validate method is declared in a validate package
    When the analyzer checks the file
    Then a "GID-213" diagnostic is reported on the type "CreateJob"

  Scenario: positive — Validate without ctx
    Given a validator with the method "func (v *UpdateJob) Validate(req *Req) error"
    When the analyzer checks the file
    Then a "GID-213" diagnostic is reported on the type "UpdateJob"

  Scenario: positive — Validate returns (bool, error)
    Given a validator with the method "func (v *DeleteJob) Validate(ctx context.Context, req *Req) (bool, error)"
    When the analyzer checks the file
    Then a "GID-213" diagnostic is reported on the type "DeleteJob"

  Scenario: negative — a correct validator
    Given a validator with the method "func (v *ListJobs) Validate(ctx context.Context, req *Req) error"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — a type with the Options suffix is not a validator
    Given "type ListJobsOptions struct{}" without a Validate method is declared in a validate package
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — an unexported struct is not flagged
    Given "type helper struct{}" without a Validate method is declared in a validate package
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — Validate with a value receiver
    Given a validator with the method "func (v GetJob) Validate(ctx context.Context, req *Req) error"
    When the analyzer checks the file
    Then no diagnostic is reported

  # The decision fixed in the rule: ctx as the first parameter and a single
  # error result are sufficient. The number of parameters after ctx is not limited.
  Scenario: boundary — Validate with two parameters after ctx
    Given a validator with the method "func (v PatchJob) Validate(ctx context.Context, req *Req, opt Opt) error"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — a struct without Validate in an ordinary package
    Given "type Worker struct{}" without a Validate method is declared in the "/domain/service" package
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — a type in settings.exclude
    Given the type "HealthCheck" is listed in settings.exclude
    And "type HealthCheck struct{}" without a Validate method is declared in a validate package
    When the analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md — outside this task's change)
#  [x] Layer chosen: go/analysis (types needed — context.Context, error)
#  [x] Severity and message are defined
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml (outside this task's change)
