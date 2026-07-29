# language: en

Feature: GID-112 — Create*/Update* methods return only error (gidcreateupdate)
  As a developer
  I want a state-changing method to report success and nothing else
  So that the write path stays one query and callers fetch the result
  explicitly when they actually need it

  # One analyzer over exported method declarations, LoadModeTypesInfo (the
  # results are read from the checked signature).
  # Scope: /dal/repository and /domain/service, matched through
  # internal/pathseg.HasLayer — anchored to the module root.
  # Trigger: the method name starts with the word Create or Update — the bare
  # verb, or the verb followed by an uppercase letter or a digit
  # (CreateJob yes, CreatedAt no) — and the signature returns anything besides
  # error.
  # Exceptions: //nolint:gidcreateupdate pointwise; settings.exclude centrally,
  # in the "Method" or "Type.Method" form (internal/exclude).
  # Generated code (ast.IsGenerated) is skipped.

  # --- Class 1: positive ---

  Scenario: positive — an entity returned alongside the error
    Given "func (j *Job) CreateJob(ctx context.Context, name string) (Snapshot, error)" in /dal/repository
    When the gidcreateupdate analyzer checks the file
    Then the diagnostic "GID-112: method \"CreateJob\" creates/updates state and must return only error. Fix: fetch the entity with a separate query (exceptions: nolint or settings.exclude)" is reported

  Scenario: positive — a single non-error result
    Given "func (j *Job) UpdateJobStatus(ctx context.Context, status string) Snapshot"
    When the gidcreateupdate analyzer checks the file
    Then the diagnostic "GID-112: method \"UpdateJobStatus\" creates/updates state and must return only error. …" is reported

  Scenario: positive — the same in a service
    Given "func (j *Job) CreateSession(ctx context.Context) (Session, error)" in /domain/service
    When the gidcreateupdate analyzer checks the file
    Then the diagnostic "GID-112: method \"CreateSession\" creates/updates state and must return only error. …" is reported

  # --- Class 2: negative ---

  Scenario: negative — only error is returned
    Given "func (j *Job) CreateJob(ctx context.Context, in *entity.CreateJob) error"
    When the gidcreateupdate analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — a read method returning data
    Given "func (j *Job) Job(ctx context.Context, id string) (entity.Job, error)"
    When the gidcreateupdate analyzer checks the file
    Then no diagnostic is reported
    # The rule is about the write path.

  # --- Class 3: boundary ---

  Scenario: boundary — a name where Create is only a prefix of a word
    Given "func (j *Job) CreatedAt(ctx context.Context) (time.Time, error)"
    When the gidcreateupdate analyzer checks the file
    Then no diagnostic is reported
    # The verb must end a word: an uppercase letter or a digit has to follow.

  Scenario: boundary — the bare verb as a name
    Given "func (j *Job) Create(ctx context.Context, in *entity.Job) (entity.Job, error)"
    When the gidcreateupdate analyzer checks the file
    Then the diagnostic "GID-112: method \"Create\" …" is reported

  Scenario: boundary — the method is listed in settings.exclude
    Given settings "exclude: [Job.CreateJob]" and that method returning an entity
    When the gidcreateupdate analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — a repository nested under another layer
    Given "func (j *Job) CreateJob(ctx context.Context) (Snapshot, error)" in svc/server/grpc/dal/repository
    When the gidcreateupdate analyzer checks the file
    Then no diagnostic is reported
    # The layer is anchored to the module root.

  Scenario: boundary — a method returning nothing at all
    Given "func (j *Job) UpdateCache(ctx context.Context)"
    When the gidcreateupdate analyzer checks the file
    Then no diagnostic is reported
    # No results means nothing to trim; the missing error is another rule's
    # business.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — a package outside the two layers
    Given "func (b *Builder) CreateQuery(table string) (string, error)" in /pkg/builder
    When the gidcreateupdate analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — an unexported method
    Given "func (j *Job) createJob(ctx context.Context) (Snapshot, error)"
    When the gidcreateupdate analyzer checks the file
    Then no diagnostic is reported
    # The rule shapes the layer's API.

  Scenario: non-applicability — generated code
    Given a file carrying the "Code generated … DO NOT EDIT." header
    When the gidcreateupdate analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-112)
#  [x] Layer chosen: go/analysis (package createupdate)
#  [x] Message is defined ("GID-112: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
