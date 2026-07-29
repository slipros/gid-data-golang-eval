# language: en

Feature: GID-102 — the word Batch is not used in method names (gidnobatch)
  As a developer
  I want a method working on many entities to be named like the single-entity
  one, in the plural
  So that CreateJob and CreateJobs read as one pair instead of CreateJob and
  BatchCreate sitting in different mental places

  # One analyzer over the file's top-level function declarations, no type info
  # needed.
  # Trigger: a method (a function with a receiver) whose name contains the
  # substring "Batch" anywhere — prefix, middle or suffix.
  # Receiverless functions are out of scope: the rule is about the method
  # naming section of the styleguide.
  # Generated code (ast.IsGenerated) is skipped.

  # Scope: only a module laid out as a service — internal/modlayout walks up to
  # the package's go.mod and looks for a layer directory (domain, dal, server,
  # app, usecase, repository) at the module root or under internal/. A flat
  # library module is skipped.

  # --- Class 1: positive ---

  Scenario: positive — Batch in the middle of the name
    Given "func (j *Job) CreateBatchJobs(in []CreateJob) error"
    When the gidnobatch analyzer checks the file
    Then the diagnostic "GID-102: method \"CreateBatchJobs\" contains the word Batch. Fix: use a plural instead (CreateJob -> CreateJobs)" is reported on the method name

  Scenario: positive — Batch as a prefix
    Given "func (j *Job) BatchCreate(in []CreateJob) error"
    When the gidnobatch analyzer checks the file
    Then the diagnostic "GID-102: method \"BatchCreate\" contains the word Batch. …" is reported

  Scenario: positive — Batch inside a name that is not about creation
    Given "func (j *Job) UpdateBatchStatus(status string) error"
    When the gidnobatch analyzer checks the file
    Then the diagnostic "GID-102: method \"UpdateBatchStatus\" contains the word Batch. …" is reported

  # --- Class 2: negative ---

  Scenario: negative — the plural form
    Given "func (j *Job) CreateJobs(in []CreateJob) error"
    When the gidnobatch analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — a single-entity method
    Given "func (j *Job) CreateJob(in CreateJob) error"
    When the gidnobatch analyzer checks the file
    Then no diagnostic is reported

  # --- Class 3: boundary ---

  Scenario: boundary — a receiverless function with Batch in its name
    Given "func BatchSize() int"
    When the gidnobatch analyzer checks the file
    Then no diagnostic is reported
    # The styleguide section the rule ports is about method naming.

  Scenario: boundary — a struct field or a parameter named batch
    Given "type Job struct { batchSize int }" and "func (j *Job) Split(batch []Item) error"
    When the gidnobatch analyzer checks the file
    Then no diagnostic is reported
    # Only the method name is read.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — the lowercase word "batch" in a method name
    Given "func (j *Job) batchless() error"
    When the gidnobatch analyzer checks the file
    Then no diagnostic is reported
    # The match is case-sensitive: it targets the Batch word of a Go name.

  Scenario: non-applicability — generated code
    Given a file carrying the "Code generated … DO NOT EDIT." header
    When the gidnobatch analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — a library module wrapping a driver
    Given "func (q *Query) Batch(ctx context.Context) error" in libs/postgres
    When the gidnobatch analyzer checks the file
    Then no diagnostic is reported
    # There Batch is the driver's own term (pgx.Batch, SendBatch), not the
    # CreateBatchJobs smell the rule targets.

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-102)
#  [x] Layer chosen: go/analysis (package nobatch)
#  [x] Message is defined ("GID-102: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
