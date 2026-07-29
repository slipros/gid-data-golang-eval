# language: en

Feature: GID-105/GID-135 — converters are named <Dst><Type>From<Src> and live in convert/ (gidconvnaming)
  As a developer
  I want every converter named after what it produces and from what, in one
  known package
  So that a conversion is found by name and the layers do not grow their own
  private dialect of ToXxx/ConvertXxx helpers

  # One analyzer carrying two rules over exported top-level functions, no type
  # info needed: the package path and the function name decide.
  # The converter name pattern is ^[A-Z][A-Za-z0-9]*From[A-Z][A-Za-z0-9]*$ —
  # words on both sides of From (EntityCreateSnapshotFromModel).
  # GID-105 applies inside a convert package (pathseg.EndsWith "convert"):
  # every exported function must match the pattern.
  # GID-135 applies outside one: a function matching the pattern must move into
  # the convert/ subpackage of its layer. Scope: the dal, domain, server and
  # event layers (pathseg.HasLayer, anchored to the module root).
  # Exceptions: <Name>FromContext helpers are ctx accessors (GID-166), not
  # converters; go-test entry points (Test/Benchmark/Example/Fuzz) carry names
  # the framework mandates and are never judged.
  # Methods and unexported functions are out of scope. Generated code
  # (ast.IsGenerated) is skipped.

  # --- Class 1: positive ---

  Scenario: positive — a Convert-prefixed name in a convert package
    Given "func ConvertSnapshot(in *model) entity" in /domain/service/convert
    When the gidconvnaming analyzer checks the file
    Then the diagnostic "GID-105: converter \"ConvertSnapshot\" must be named <Dst><Type>From<Src>. Fix: rename it, e.g. EntityCreateSnapshotFromModel" is reported

  Scenario: positive — a To-prefixed name in a convert package
    Given "func ToEntity(in *model) entity" in /domain/service/convert
    When the gidconvnaming analyzer checks the file
    Then the diagnostic "GID-105: converter \"ToEntity\" must be named <Dst><Type>From<Src>. …" is reported

  Scenario: positive — a converter declared outside convert/
    Given "func ModelSnapshotFromRow(in *Row) Snapshot" in /domain/service
    When the gidconvnaming analyzer checks the file
    Then the diagnostic "GID-135: converter \"ModelSnapshotFromRow\" must live in a convert/ subpackage of its layer. Fix: move it into convert/" is reported

  # --- Class 2: negative ---

  Scenario: negative — a correctly named converter in convert/
    Given "func EntityCreateSnapshotFromModel(in *model.CreateSnapshot) entity.CreateSnapshot" in /domain/service/convert
    When the gidconvnaming analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — an ordinary function outside a convert package
    Given "func ParseDuration(s string) (time.Duration, error)" in /pkg/parse
    When the gidconvnaming analyzer checks the file
    Then no diagnostic is reported
    # It does not match the converter pattern, and /pkg is not a layer in scope.

  # --- Class 3: boundary ---

  Scenario: boundary — a ctx accessor
    Given "func UserIDFromContext(ctx context.Context) (string, bool)" in /server/http/handler
    When the gidconvnaming analyzer checks the file
    Then no diagnostic is reported
    # <Name>FromContext is GID-166's shape, not a converter.

  Scenario: boundary — a test entry point in a convert package
    Given "func TestEntityFromModel(t *testing.T)" in /domain/service/convert
    When the gidconvnaming analyzer checks the file
    Then no diagnostic is reported
    # The name is fixed by `go test`; tests live in the same package (GID-250),
    # so they land in the convert package too.

  Scenario: boundary — a converter-shaped name in a package outside every layer
    Given "func ModelSnapshotFromRow(in *Row) Snapshot" in /pkg/parse
    When the gidconvnaming analyzer checks the file
    Then no diagnostic is reported
    # GID-135 moves converters into their layer's convert/; a package outside
    # the dal/domain/server/event layers has no such subpackage.

  Scenario: boundary — an unexported converter in a convert package
    Given "func entitySnapshotFromModel(in model.Snapshot) entity.Snapshot" in /domain/service/convert
    When the gidconvnaming analyzer checks the file
    Then no diagnostic is reported
    # The naming rule states the package's public vocabulary.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — a method with a converter-shaped name
    Given "func (s *Snapshot) ModelFromEntity(in entity.Snapshot) model.Snapshot"
    When the gidconvnaming analyzer checks the file
    Then no diagnostic is reported
    # Only package-level functions are converters here.

  Scenario: non-applicability — generated code
    Given a file carrying the "Code generated … DO NOT EDIT." header
    When the gidconvnaming analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-105, GID-135)
#  [x] Layer chosen: go/analysis (package convnaming)
#  [x] Messages are defined ("GID-105: …", "GID-135: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
