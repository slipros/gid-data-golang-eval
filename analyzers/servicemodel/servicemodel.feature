# language: en

Feature: GID-151 — the service API works with model, not entity (gidservicemodel)
  As a developer
  I want exported service methods to take and return model types
  So that callers of the domain never learn the storage shape, and swapping a
  repository does not ripple through the whole service API

  # One analyzer over exported method declarations, LoadModeTypesInfo (the
  # signature is read from the checked types, not from the AST).
  # Scope: the root of /domain/service, matched through
  # internal/pathseg.EndsWith.
  # Trigger: a parameter or a result referencing a type from /dal/entity. The
  # search is recursive through pointers, slices, arrays, maps (key and value),
  # channels, aliases and named types' underlying types, with a seen-set to
  # survive recursive types.
  # The entity layer is matched with pathseg.HasLayer — anchored to the module
  # root, so a dal/entity nested under another layer is not it.
  # The body is not inspected: converting to an entity inside the method is
  # exactly what the rule asks for.
  # Generated code (ast.IsGenerated) is skipped.

  # --- Class 1: positive ---

  Scenario: positive — an entity as a parameter
    Given "func (s *Snapshot) CreateSnapshot(ctx context.Context, in *entity.CreateSnapshot) error"
    When the gidservicemodel analyzer checks the file
    Then the diagnostic "GID-151: method \"CreateSnapshot\" uses the entity type entity.CreateSnapshot (parameter). Fix: the service API takes and returns model, convert to entity internally" is reported

  Scenario: positive — an entity as a result
    Given "func (s *Snapshot) SnapshotRaw(ctx context.Context, id string) (entity.Snapshot, error)"
    When the gidservicemodel analyzer checks the file
    Then the diagnostic "GID-151: method \"SnapshotRaw\" uses the entity type entity.Snapshot (result). …" is reported

  Scenario: positive — an entity behind a named slice type
    Given "func (s *Snapshot) SnapshotsRaw(ctx context.Context) (entity.Snapshots, error)"
    When the gidservicemodel analyzer checks the file
    Then the diagnostic "GID-151: method \"SnapshotsRaw\" uses the entity type entity.Snapshots (result). …" is reported

  Scenario: positive — an entity as a map value
    Given "func (s *Snapshot) SnapshotsByID(ctx context.Context) (map[string]*entity.Snapshot, error)"
    When the gidservicemodel analyzer checks the file
    Then the diagnostic "GID-151: method \"SnapshotsByID\" uses the entity type entity.Snapshot (result). …" is reported
    # The search walks into the map, and through the pointer behind it.

  # --- Class 2: negative ---

  Scenario: negative — a model-only signature
    Given "func (s *Snapshot) Snapshot(ctx context.Context, id string) (model.Snapshot, error)"
    When the gidservicemodel analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — the entity is used inside the body
    Given a method converting a model to "entity.Snapshot" and passing it to the repository
    When the gidservicemodel analyzer checks the file
    Then no diagnostic is reported
    # Conversion inside is the intended design.

  # --- Class 3: boundary ---

  Scenario: boundary — an unexported method
    Given "func (s *Snapshot) toEntity(in model.Snapshot) entity.Snapshot"
    When the gidservicemodel analyzer checks the file
    Then no diagnostic is reported
    # The rule guards the service's API; an unexported converter is internals.

  Scenario: boundary — a receiverless function in the service package
    Given "func convertSnapshot(in entity.Snapshot) model.Snapshot"
    When the gidservicemodel analyzer checks the file
    Then no diagnostic is reported
    # Only methods form the service API.

  Scenario: boundary — a dal/entity nested under another layer
    Given a signature using a type from svc/server/grpc/dal/entity
    When the gidservicemodel analyzer checks the file
    Then no diagnostic is reported
    # The entity layer is anchored to the module root.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — a package outside /domain/service
    Given the same signature in /domain/usecase or /dal/repository
    When the gidservicemodel analyzer checks the file
    Then no diagnostic is reported
    # A repository is where the entity belongs.

  Scenario: non-applicability — generated code
    Given a file carrying the "Code generated … DO NOT EDIT." header
    When the gidservicemodel analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-151)
#  [x] Layer chosen: go/analysis (package servicemodel)
#  [x] Message is defined ("GID-151: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
