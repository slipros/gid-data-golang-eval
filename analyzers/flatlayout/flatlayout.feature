# language: en

Feature: GID-138 — repositories and services live at the root of their layer (gidflatlayout)
  As a developer
  I want every repository and service to sit directly in /dal/repository and
  /domain/service
  So that an entity is found by its layer alone, without guessing which
  grouping subfolder someone chose for it

  # One analyzer over the package path, no type info needed.
  # Two layer roots are checked: dal/repository and domain/service, both
  # matched through internal/pathseg (LayerSegments/HasLayer) — anchored to the
  # module root, so a dal/repository nested under another layer is not it.
  # Trigger: the package sits one or more segments below the layer root and
  # that next segment is not an allowed one.
  # Allowed subpackages come from the styleguide: convert/ and build/ under
  # /dal/repository, convert/ under /domain/service.
  # The diagnostic is reported on the package clause of every file of the
  # offending package.

  # --- Class 1: positive ---

  Scenario: positive — a technology-grouping subpackage of a repository
    Given the package "svc/dal/repository/redis"
    When the gidflatlayout analyzer checks the package
    Then the diagnostic "GID-138: package \"svc/dal/repository/redis\". Fix: grouping subpackages in /dal/repository are forbidden, keep layer entities at its root" is reported on the package clause

  Scenario: positive — the same grouping under a service
    Given the package "svc/domain/service/redis"
    When the gidflatlayout analyzer checks the package
    Then the diagnostic "GID-138: package \"svc/domain/service/redis\". Fix: grouping subpackages in /domain/service are forbidden, keep layer entities at its root" is reported

  Scenario: positive — build/ under a service
    Given the package "svc/domain/service/build"
    When the gidflatlayout analyzer checks the package
    Then the diagnostic "GID-138: package \"svc/domain/service/build\". …" is reported
    # build/ is legitimate for a repository (SQL builders); a service has no
    # such subpackage — only convert/.

  # --- Class 2: negative ---

  Scenario: negative — the layer root itself
    Given the packages "svc/dal/repository" and "svc/domain/service"
    When the gidflatlayout analyzer checks the packages
    Then no diagnostic is reported

  Scenario: negative — the allowed subpackages of a repository
    Given the packages "svc/dal/repository/convert" and "svc/dal/repository/build"
    When the gidflatlayout analyzer checks the packages
    Then no diagnostic is reported

  # --- Class 3: boundary ---

  Scenario: boundary — a "repository" directory outside the dal layer
    Given the package "svc/pkg/repository"
    When the gidflatlayout analyzer checks the package
    Then no diagnostic is reported
    # The rule matches the layer root dal/repository as a segment sequence, not
    # the word "repository" anywhere in the path.

  Scenario: boundary — dal/repository nested under another layer
    Given the package "svc/server/grpc/dal/repository/redis"
    When the gidflatlayout analyzer checks the package
    Then no diagnostic is reported
    # The layer is anchored to the module root, so this is not the dal layer.

  Scenario: boundary — a second level below the layer root
    Given the package "svc/dal/repository/redis/internal"
    When the gidflatlayout analyzer checks the package
    Then the diagnostic "GID-138: …" is reported
    # The segment right below the root decides, however deep the package goes.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — a layer the rule says nothing about
    Given the package "svc/domain/usecase/redis"
    When the gidflatlayout analyzer checks the package
    Then no diagnostic is reported
    # Only the two layer roots listed in the rule are checked.

  Scenario: non-applicability — a package outside any layer
    Given the package "svc/internal/app"
    When the gidflatlayout analyzer checks the package
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-138)
#  [x] Layer chosen: go/analysis (package flatlayout)
#  [x] Message is defined ("GID-138: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
