# language: en

Feature: GID-160 — a service calls gRPC through a repository (gidgrpcinservice)
  As a developer
  I want the domain layer to reach a remote system through a repository
  So that the business logic depends on an interface it owns, and the
  transport stays replaceable and mockable

  # One analyzer over the file's import declarations, LoadModeTypesInfo (the
  # imports of the imports are read from the type-checked package).
  # Scope: /domain/service and /domain/usecase, matched through
  # internal/pathseg.HasLayer — anchored to the module root.
  # Two triggers:
  #   1. a direct import of google.golang.org/grpc;
  #   2. an import of a package that itself imports google.golang.org/grpc —
  #      this catches generated pb stubs and hand-written gRPC clients without
  #      hardcoding their paths.
  # Exceptions: //nolint:gidgrpcinservice pointwise, settings.exclude (a list
  # of import paths) centrally.
  # Generated code (ast.IsGenerated) is skipped.

  # --- Class 1: positive ---

  Scenario: positive — grpc imported directly by a service
    Given "google.golang.org/grpc" imported in /domain/service
    When the gidgrpcinservice analyzer checks the file
    Then the diagnostic "GID-160: direct import of google.golang.org/grpc in the domain layer is forbidden. Fix: call gRPC through a repository (exceptions: nolint or settings.exclude)" is reported on the import

  Scenario: positive — a pb stub imported by a service
    Given "svc/pkg/api/orderpb", a package importing google.golang.org/grpc, imported in /domain/service
    When the gidgrpcinservice analyzer checks the file
    Then the diagnostic "GID-160: importing the gRPC package \"svc/pkg/api/orderpb\" in the domain layer is forbidden. Fix: call gRPC through a repository (exceptions: nolint or settings.exclude)" is reported

  Scenario: positive — the usecase layer is in scope too
    Given "google.golang.org/grpc" imported in /domain/usecase
    When the gidgrpcinservice analyzer checks the file
    Then the diagnostic "GID-160: direct import of google.golang.org/grpc in the domain layer is forbidden. …" is reported

  # --- Class 2: negative ---

  Scenario: negative — the repository owns the gRPC call
    Given "google.golang.org/grpc" and the pb stub imported in /dal/repository
    When the gidgrpcinservice analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — the service imports no transport
    Given a /domain/service package importing only context and the model
    When the gidgrpcinservice analyzer checks the file
    Then no diagnostic is reported

  # --- Class 3: boundary ---

  Scenario: boundary — the import path is listed in settings.exclude
    Given settings "exclude: [excluded/pkg/api/orderpb]" and that import in /domain/service
    When the gidgrpcinservice analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — a domain/service nested under another layer
    Given "google.golang.org/grpc" imported in svc/server/grpc/domain/service
    When the gidgrpcinservice analyzer checks the file
    Then no diagnostic is reported
    # The layer is anchored to the module root: a package below /server/grpc is
    # not the domain layer.

  Scenario: boundary — /domain/model
    Given the pb stub imported in /domain/model
    When the gidgrpcinservice analyzer checks the file
    Then no diagnostic is reported
    # The scope is the two layers named in the rule; the model is guarded by
    # its own rules.

  Scenario: boundary — a transitive dependency two levels down
    Given a service importing a package whose own dependency imports grpc
    When the gidgrpcinservice analyzer checks the file
    Then no diagnostic is reported
    # Only the direct imports of the imported package are inspected — one level,
    # deliberately: deeper matching would flag any library that happens to speak
    # gRPC somewhere inside.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — a package outside the domain layer
    Given "google.golang.org/grpc" imported in /server/grpc/handler
    When the gidgrpcinservice analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — generated code
    Given a file carrying the "Code generated … DO NOT EDIT." header
    When the gidgrpcinservice analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-160)
#  [x] Layer chosen: go/analysis (package grpcinservice)
#  [x] Message is defined ("GID-160: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
