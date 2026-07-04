# language: en

Feature: GID-224…229 — the layer-isolation matrix
  As a service architect
  I want each layer to import only what it is allowed to
  So that dependencies flow in one direction and wiring lives in the composition root

  # GID-224: transport sees only domain/model (and validate)

  Scenario: server imports domain/service — violation
    Given the package "svc/server/http/handler" imports "svc/domain/service"
    When the analyzer checks the file
    Then a "GID-224" diagnostic is reported on the import "svc/domain/service"

  Scenario: server imports dal/repository — violation
    Given the package "svc/server/http/handler" imports "svc/dal/repository"
    When the analyzer checks the file
    Then a "GID-224" diagnostic is reported on the import "svc/dal/repository"

  Scenario: schedule imports domain/service — violation
    Given the package "svc/schedule/sync" imports "svc/domain/service"
    When the analyzer checks the file
    Then a "GID-224" diagnostic is reported on the import "svc/domain/service"

  Scenario: validate imports dal/entity — violation
    Given the package "svc/validate" imports "svc/dal/entity"
    When the analyzer checks the file
    Then a "GID-224" diagnostic is reported on the import "svc/dal/entity"

  Scenario: event consumer imports dal/entity and domain/service — violations
    Given the package "svc/event/consumer" imports "svc/dal/entity" and "svc/domain/service"
    When the analyzer checks the file
    Then a "GID-224" diagnostic is reported on both imports

  Scenario: transport imports domain/model and validate — ok
    Given the package "svc/server/http/handler" imports "svc/domain/model" and "svc/validate"
    When the analyzer checks the file
    Then no diagnostic is reported

  # GID-225: the composition root and transport are leaves

  Scenario: domain imports app — violation
    Given the package "svc/domain/notifier" imports "svc/app"
    When the analyzer checks the file
    Then a "GID-225" diagnostic is reported on the import "svc/app"

  Scenario: domain imports server — violation
    Given the package "svc/domain/notifier" imports "svc/server/middleware"
    When the analyzer checks the file
    Then a "GID-225" diagnostic is reported on the import "svc/server/middleware"

  Scenario: app imports all layers — ok (composition root)
    Given the package "svc/app" imports repository, service, producer, client and metric
    When the analyzer checks the file
    Then no diagnostic is reported

  # GID-226: metric is self-contained

  Scenario: metric imports domain/model — violation
    Given the package "svc/metric" imports "svc/domain/model"
    When the analyzer checks the file
    Then a "GID-226" diagnostic is reported on the import "svc/domain/model"

  Scenario: domain/service imports metric — violation
    Given the package "svc/domain/service" imports "svc/metric"
    When the analyzer checks the file
    Then a "GID-226" diagnostic is reported on the import "svc/metric"

  Scenario: domain imports a package with the segment "metrics" — boundary, ok
    Given the package "svc/domain/boundary" imports "svc/metrics/registry"
    When the analyzer checks the file
    Then no diagnostic is reported

  # GID-227: model is a pure vocabulary

  Scenario: model imports transport — violation
    Given the package "svc/domain/model" imports "svc/server/middleware"
    When the analyzer checks the file
    Then a "GID-227" diagnostic is reported on the import "svc/server/middleware"

  Scenario: usecase imports a model subpackage — ok (the model layer)
    Given the package "svc/domain/usecase" imports "svc/domain/model/filter"
    When the analyzer checks the file
    Then no diagnostic is reported

  # GID-228: usecase does not call a client directly

  Scenario: domain/usecase imports client — violation
    Given the package "svc/domain/usecase" imports "svc/client/billing"
    When the analyzer checks the file
    Then a "GID-228" diagnostic is reported on the import "svc/client/billing"

  Scenario: domain/service imports client — ok (the service converts model <-> client models)
    Given the package "svc/domain/service" imports "svc/client/billing"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: dal/repository imports client — ok (client models are converted to entity in convert)
    Given the package "svc/dal/repository" imports "svc/client/billing"
    When the analyzer checks the file
    Then no diagnostic is reported

  # GID-229: the client is isolated

  Scenario: client imports domain/model — violation
    Given the package "svc/client/billing" imports "svc/domain/model"
    When the analyzer checks the file
    Then a "GID-229" diagnostic is reported on the import "svc/domain/model"

  Scenario: client imports a third-party package — the rule does not apply
    Given the package "svc/client/billing" imports "strconv"
    When the analyzer checks the file
    Then no diagnostic is reported

  # Module boundary: the pkg/<module> application-module layout (module.md)

  Scenario: server imports domain/service inside pkg/<module> — violation
    Given the package "repo/pkg/billing/server/handler" imports "repo/pkg/billing/domain/service"
    When the analyzer checks the file
    Then a "GID-224" diagnostic is reported on the import "repo/pkg/billing/domain/service"

  Scenario: dal/repository imports domain/model inside pkg/<module> — violation
    Given the package "repo/pkg/billing/dal/repository" imports "repo/pkg/billing/domain/model"
    When the analyzer checks the file
    Then a "GID-132" diagnostic is reported on the import "repo/pkg/billing/domain/model"

  Scenario: domain/service imports dal/repository inside pkg/<module> — violation
    Given the package "repo/pkg/billing/domain/service" imports "repo/pkg/billing/dal/repository"
    When the analyzer checks the file
    Then a "GID-132" diagnostic is reported on the import "repo/pkg/billing/dal/repository"

  Scenario: server imports domain/model inside pkg/<module> — ok
    Given the package "repo/pkg/billing/server/handler" imports "repo/pkg/billing/domain/model"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: pkg/<module> imports a dal entity from repo/internal/** — a different module, ok
    Given the package "repo/pkg/billing/server/handler" imports "repo/internal/dal/entity"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: pkg/<module> imports domain/model from repo/internal/** — a different module, ok
    Given the package "repo/pkg/billing/domain/service" imports "repo/internal/domain/model"
    When the analyzer checks the file
    Then no diagnostic is reported

  # Third-party libraries and settings

  Scenario: a layer imports a third-party library with the segment "client" — ok
    Given the import belongs to another module (the module prefix does not match)
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a rule is disabled via settings.disable — no diagnostic
    Given settings.disable: [GID-224] is set in .golangci.yml
    And the package "custom/server/handler" imports "custom/domain/service"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a custom rule via settings.rules — a diagnostic with its ID
    Given the rule SVC-1 is set in .golangci.yml: scope "domain/service", banned [legacy]
    And the package "custom/domain/service" imports "custom/legacy/store"
    When the analyzer checks the file
    Then an "SVC-1" diagnostic is reported on the import "custom/legacy/store"

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the RULES.md registry
#  [x] Layer chosen: go/analysis (import-path segments are needed)
#  [x] Severity and message are defined
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml

  # --- GID-241: repository imports are allow-listed (requirement 2026-07-04) ---

  Scenario: an unknown new folder imports the repository — violation
    Given package "svc/cron" (not present in the deny matrix) imports "svc/dal/repository"
    When the analyzer checks the file
    Then a "GID-241" diagnostic tells to declare a consumer-side interface and wire the repository in app

  Scenario: the composition root imports the repository — ok
    Given package "svc/app" imports "svc/dal/repository"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: the repository layer imports its own subpackage — ok
    Given package "svc/dal/repository" imports "svc/dal/repository/build"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a deny-matrix scope imports the repository — the matrix wins, one diagnostic
    Given package "svc/domain/service" imports "svc/dal/repository"
    When the analyzer checks the file
    Then only a "GID-132" diagnostic is reported (no GID-241 duplicate)
