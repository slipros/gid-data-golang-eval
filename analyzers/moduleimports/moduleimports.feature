# language: en

Feature: GID-259 — an application module owns its data layer (gidmoduleimports)
  As a developer
  I want a module under pkg/<module> to talk to its own dal and to the core
  only through the core's business layer
  So that a module cannot write another layer's tables behind its back

  # Scope: a package sitting inside a JUDGED layer of an application module —
  # pkg/<module>/<layer>/**, settings.layers, default ["domain", "dal"],
  # recognised through pathseg.LayerSegments. LoadModeTypesInfo; imports are
  # read off the file, so the alias does not matter.
  # Trigger: an import of <service>/internal/<layer>/... whose layer is not in
  # settings.allow (default ["domain"]).
  # Generated code (ast.IsGenerated) is skipped; _test.go files are NOT.

  # --- Class 1: positive ---

  Scenario: positive — the module's service reaching into the core dal
    Given a package "svc/pkg/integration/push/firebase/domain/service" importing "svc/internal/dal/entity"
    When the gidmoduleimports analyzer checks the file
    Then the diagnostic "GID-259: module package \"svc/pkg/integration/push/firebase/domain/service\" must not import the core layer \"svc/internal/dal/entity\" — a module owns its dal, and only the core /domain/** is shared. Fix: declare the repository interface over the module's own entity (<module>/dal/entity), and take core data through a core service injected in module.go" is reported on the import
    # The incident shape (2026-08-06, resource-registry): the module's service
    # declared CoreIntegrationRepository over the CORE entity and wrote
    # public.integration through the core data layer.

  Scenario: positive — the core repository imported by a module layer
    Given the same package importing "svc/internal/dal/repository"
    When the gidmoduleimports analyzer checks the file
    Then the diagnostic is reported
    # A concrete repository is wiring material; a layer package consumes an
    # interface (GID-134/241).

  Scenario: positive — a core client is not a shared layer either
    Given the same package importing "svc/internal/client/vendor"
    When the gidmoduleimports analyzer checks the file
    Then the diagnostic is reported

  Scenario: positive — the module's repository reaching for the core sentinels
    Given "svc/pkg/integration/push/firebase/dal/repository" importing "svc/internal/dal/entity" for commonentity.ErrNoResult
    When the gidmoduleimports analyzer checks the file
    Then the diagnostic is reported
    # The module's dal is judged for the same reason its service is: the core's
    # failures reach a module as DOMAIN errors of the core service/usecase, and
    # a module usecase handles them from there (owner decision 2026-08-06).

  # --- Class 2: negative ---

  Scenario: negative — the shared core /domain/**
    Given imports of "svc/internal/domain/model" and "svc/internal/domain/service"
    When the gidmoduleimports analyzer checks the file
    Then no diagnostic is reported
    # The core model is the shared vocabulary; the core service is how a module
    # takes core data.

  Scenario: negative — the module's own layers
    Given an import of "svc/pkg/integration/push/firebase/dal/entity"
    When the gidmoduleimports analyzer checks the file
    Then no diagnostic is reported

  # --- Class 3: boundary ---

  Scenario: boundary — a module nested deeper than one segment
    Given the module "pkg/integration/push/firebase" (category and vendor directories above it)
    When the gidmoduleimports analyzer checks the file
    Then its packages are still recognised as module layers and judged
    # Before the pathseg fix of 2026-08-06 the layer path of such a package came
    # out as ["push","firebase","domain","service"], so every layer rule was
    # silent inside the whole module tree.

  Scenario: boundary — settings.allow widens the shared core
    Given settings.allow ["domain", "client"] and imports of the core client and the core dal
    When the gidmoduleimports analyzer checks the file
    Then only the core dal import is reported

  Scenario: boundary — settings.layers adds the module transport to the scope
    Given settings.layers ["domain", "dal", "server"] and "custom/pkg/billing/server/http" importing the core dal
    When the gidmoduleimports analyzer checks the file
    Then the diagnostic is reported

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — the module root (module.go)
    Given the package "svc/pkg/integration/push/firebase" importing "svc/internal/dal/repository"
    When the gidmoduleimports analyzer checks the file
    Then no diagnostic is reported
    # The module root is its composition root — the same carve-out GID-241
    # makes for /app/**. A grouping directory above the module is skipped for
    # the same reason: it holds no layer.

  Scenario: non-applicability — a transport layer of the module
    Given "svc/pkg/integration/push/firebase/server/grpc" importing "svc/internal/server/router"
    When the gidmoduleimports analyzer checks the file
    Then no diagnostic is reported
    # Out of the default settings.layers: what a module takes from the core
    # there is shared infrastructure — the validator's i18n registrar, the http
    # router wiring (lk-api) — not core data.

  Scenario: non-applicability — a package of the core itself
    Given the package "svc/internal/domain/usecase" importing "svc/internal/dal/repository"
    When the gidmoduleimports analyzer checks the file
    Then no diagnostic is reported
    # The core reaching into its own layers is GID-132/241's business.

  Scenario: non-applicability — settings.exclude
    Given settings.exclude ["custom/internal/dal/repository"]
    When the gidmoduleimports analyzer checks the file
    Then that import is silent, and "custom/internal/dal/entity" is still reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-259)
#  [x] Layer chosen: go/analysis (package moduleimports)
#  [x] Message is defined ("GID-259: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
