# language: en

Feature: GID-111 — input data by pointer, output by value (gidinout)
  As a developer
  I want layer methods to take their data as *T and return it as T
  So that a call site tells input from output by the signature alone, without
  wondering whether a returned pointer may be nil

  # One analyzer over exported method declarations, LoadModeTypesInfo (the
  # parameter and result types are resolved to their declaring package).
  # Scope: /dal/repository, /domain/service, /domain/usecase (pathseg.HasLayer,
  # anchored to the module root) plus handler leaf packages, matched by the
  # trailing "handler" segment — a handler is a leaf, not a module-root layer.
  # What counts as layer data there: a struct declared under /domain/model or
  # /dal/entity.
  # The /client/** tree is scoped separately: a client has no model or entity
  # tree of its own, so its own same-module named structs (its request and
  # response types, wherever they are declared) are its data. Same-module is
  # decided by pathseg.ModuleRoot — not by the first path segment, since every
  # github.com/<org>/<repo> package shares "github.com".
  # Two triggers: such a struct passed by value as a parameter, or returned as
  # a pointer.
  # Exceptions: //nolint:gidinout pointwise; settings.exclude centrally, in the
  # "Method" or "Type.Method" form.
  # Generated code (ast.IsGenerated) is skipped.

  # --- Class 1: positive ---

  Scenario: positive — a model struct passed by value
    Given "func (s *Snapshot) Create(ctx context.Context, in model.CreateSnapshot) error" in /domain/service
    When the gidinout analyzer checks the file
    Then the diagnostic "GID-111: input data must be passed by pointer. Fix: use *model.CreateSnapshot" is reported on the parameter

  Scenario: positive — a model struct returned by pointer
    Given "func (s *Snapshot) Snapshot(ctx context.Context, id string) (*model.Snapshot, error)"
    When the gidinout analyzer checks the file
    Then the diagnostic "GID-111: output data must be returned by value. Fix: use model.Snapshot" is reported on the result

  Scenario: positive — the client's own request type by value
    Given "func (c *Client) Create(ctx context.Context, in CreateSnapshotRequest) error" in /client/snapshot
    When the gidinout analyzer checks the file
    Then the diagnostic "GID-111: input data must be passed by pointer. Fix: use *snapshot.CreateSnapshotRequest" is reported

  Scenario: positive — the client's own response type by pointer
    Given "func (c *Client) Get(ctx context.Context, id string) (*Snapshot, error)" in /client/snapshot
    When the gidinout analyzer checks the file
    Then the diagnostic "GID-111: output data must be returned by value. Fix: use snapshot.Snapshot" is reported

  # --- Class 2: negative ---

  Scenario: negative — the canonical shape
    Given "func (s *Snapshot) Create(ctx context.Context, in *model.CreateSnapshot) error" and "func (s *Snapshot) Snapshot(ctx context.Context, id string) (model.Snapshot, error)"
    When the gidinout analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — scalar parameters and results
    Given "func (s *Snapshot) Count(ctx context.Context, status string) (int, error)"
    When the gidinout analyzer checks the file
    Then no diagnostic is reported
    # Only model/entity (or, in a client, its own) structs are the data the
    # rule shapes.

  # --- Class 3: boundary ---

  Scenario: boundary — the method is listed in settings.exclude
    Given settings "exclude: [Snapshot.Create]" and that method taking the model by value
    When the gidinout analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — a foreign-module struct in a client
    Given "func (c *Client) Send(ctx context.Context, in pb.CreateRequest) error" in /client/snapshot, where pb comes from another module
    When the gidinout analyzer checks the file
    Then no diagnostic is reported
    # Only the client's own types are its data; a generated stub is not shaped
    # by our rules.

  Scenario: boundary — an unexported method
    Given "func (s *Snapshot) toModel(in entity.Snapshot) model.Snapshot"
    When the gidinout analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — a handler leaf package
    Given the same by-value model parameter in /server/grpc/handler
    When the gidinout analyzer checks the file
    Then the diagnostic "GID-111: input data must be passed by pointer. …" is reported
    # Handlers are matched by their trailing segment, not as a module-root layer.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — a package outside every scope
    Given the same signatures in /pkg/util
    When the gidinout analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — a receiverless function
    Given "func Normalize(in model.Snapshot) model.Snapshot"
    When the gidinout analyzer checks the file
    Then no diagnostic is reported
    # The rule shapes the methods of a layer's entities.

  Scenario: non-applicability — generated code
    Given a file carrying the "Code generated … DO NOT EDIT." header
    When the gidinout analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-111)
#  [x] Layer chosen: go/analysis (package inout)
#  [x] Message is defined ("GID-111: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest (layer and client scopes)
#  [x] Rule enabled in .golangci.yml
