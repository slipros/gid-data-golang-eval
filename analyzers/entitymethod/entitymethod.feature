# language: en

Feature: GID-114 — repo/service methods are named after the entity (entitymethod)
  As a developer
  I want exported struct methods in /dal/repository and /domain/service
  to be named after the entity: no List prefix, no ByID suffix, with the entity name in the method name
  So that the layer API reads uniformly (Jobs instead of ListJobs, Job(ctx, id) instead of JobByID)

  # One analyzer entitymethod → linter gidentitymethod, LoadModeTypesInfo.
  # Scope — root packages of the layer by path segments (pathseg.EndsWith):
  #   /dal/repository and /domain/service. Subpackages convert/build are out of scope.
  # Only EXPORTED struct methods are checked (fn.Recv != nil).
  # New* constructors are functions without a receiver and do not fall under this.
  # Generated code (ast.IsGenerated) is skipped.
  #
  # Three checks (in priority order, the first one that fires is reported):
  #   1. the List prefix at a CamelCase boundary is forbidden → "drop the List prefix …";
  #   2. the exact ByID suffix is forbidden → "drop the ByID suffix …"
  #      (ByStageID and other By<Field>ID are allowed — selection refinement);
  #   3. the method name must contain the receiver type name as a CamelCase substring.
  #      Applies only to a meaningful entity name (len > 2);
  #      one-letter/auxiliary receivers (T, S, ID) are not checked.
  #
  # FP zone: verb methods without the entity name (Close, Ping, Flush) are rarely
  # legitimate in a repository, but they happen — they fall under check 3 and are
  # suppressed via //nolint:gidentitymethod or settings.exclude ("Method" | "Type.Method").

  # --- Class 1: positive (the violation is caught) ---

  Scenario: positive — List prefix
    Given the method "func (j *Job) ListJobs(ctx context.Context) ([]Snapshot, error)" in /dal/repository
    When the gidentitymethod analyzer checks the file
    Then the diagnostic "GID-114: drop the List prefix. Fix: use the plural Jobs instead of ListJobs" is reported

  Scenario: positive — ByID suffix
    Given the method "func (j *Job) JobByID(ctx context.Context, id string) (Snapshot, error)" in /dal/repository
    When the gidentitymethod analyzer checks the file
    Then the diagnostic "GID-114: drop the ByID suffix. Fix: use Job(ctx, id) instead of JobByID" is reported

  Scenario: positive — the method name does not contain the entity
    Given the method "func (j *Job) Fetch(ctx context.Context) (Snapshot, error)" in /dal/repository
    When the gidentitymethod analyzer checks the file
    Then the diagnostic "GID-114: method name \"Fetch\" must contain the entity name \"Job\"" is reported

  Scenario: positive — a verb method without the entity (FP zone)
    Given the method "func (j *Job) Close() error" in /dal/repository
    When the gidentitymethod analyzer checks the file
    Then the diagnostic "GID-114: method name \"Close\" must contain the entity name \"Job\"" is reported
    # This is exactly the FP zone: a legitimate Close/Ping/Flush is caught — suppressed via exclude/nolint.

  # --- Class 2: negative (clean code passes) ---

  Scenario: negative — names derived from the entity
    Given the methods "Job", "Jobs", "CreateJob", "DeleteJob" on the type Job in /dal/repository
    When the gidentitymethod analyzer checks the file
    Then no diagnostic is reported

  # --- Class 3: boundary (looks like a violation but is allowed) ---

  Scenario: boundary — the ByStageID suffix is allowed
    Given the method "func (j *Job) JobsByStageID(ctx context.Context, stageID string) ([]Snapshot, error)" in /dal/repository
    When the gidentitymethod analyzer checks the file
    Then no diagnostic is reported
    # ByStageID is a selection refinement (By<Field>ID), not the exact ByID suffix; the name contains Job.

  Scenario: boundary — Listen does not count as the List prefix
    Given the method "func (j *Job) ListenJobEvents(ctx context.Context) error" in /dal/repository
    When the gidentitymethod analyzer checks the file
    Then no diagnostic is reported
    # CamelCase boundary: "List" is followed by a lowercase "e" — it is not the word List.

  Scenario: boundary — an unexported method is not matched
    Given the method "func (j *Job) listJobsInternal(ctx context.Context) ([]Snapshot, error)" in /dal/repository
    When the gidentitymethod analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — a one-letter entity: check 3 does not apply
    Given the type "S" and the method "func (x *S) Touch(ctx context.Context) error" in /domain/service
    When the gidentitymethod analyzer checks the file
    Then no diagnostic is reported
    # len(recv) <= 2 — the entity name is auxiliary, the "must contain the entity" requirement is not checked.
    # The List prefix / ByID suffix are still caught (they do not depend on the length).

  Scenario: boundary — settings.exclude suppresses a verb method
    Given settings.exclude = ["Job.Close", "Ping"] and the methods Close, Ping, Flush on the type Job
    When the gidentitymethod analyzer checks the file
    Then a diagnostic is reported only on Flush; Close and Ping are suppressed

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — the convert subpackage is out of scope
    Given the method "ListSnapshots" on the type Mapper in /dal/repository/convert
    When the gidentitymethod analyzer checks the file
    Then no diagnostic is reported
    # Scope is the layer root (EndsWith), the convert/build subpackages are not affected.

  Scenario: non-applicability — /domain/usecase is out of scope
    Given the methods "ListJobs" and "Fetch" on the type Job in /domain/usecase
    When the gidentitymethod analyzer checks the file
    Then no diagnostic is reported
    # The rule's scope is only repository and service; usecase is not affected.

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-114)
#  [x] Layer chosen: go/analysis (package entitymethod: gidentitymethod), LoadModeTypesInfo
#  [x] Message is defined ("GID-114: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml
