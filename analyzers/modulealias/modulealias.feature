# language: en

Feature: GID-240 — shared internal entities need a common-prefixed alias inside pkg/<module>
  As a service architect
  I want imports of shared internal/** entities inside an application module
  (pkg/<module>, module.md) to carry a common-prefixed alias
  So that a reader can tell the module's own types apart from the entities
  it borrows from internal/

  Scenario: no alias — violation
    Given the package "repo/pkg/billing/domain/usecase" imports "repo/internal/domain/service" without an alias
    When the analyzer checks the file
    Then a "GID-240" diagnostic is reported on the import "repo/internal/domain/service"

  Scenario: alias without the common prefix — violation
    Given the package "repo/pkg/billing/domain/usecase" imports "repo/internal/domain/model" as "svc"
    When the analyzer checks the file
    Then a "GID-240" diagnostic is reported on the import "repo/internal/domain/model"

  Scenario: a dot-import — violation
    Given the package "repo/pkg/billing/server/handler" imports "repo/internal/domain/model" as "."
    When the analyzer checks the file
    Then a "GID-240" diagnostic is reported on the import "repo/internal/domain/model"

  Scenario: a common-prefixed alias — ok
    Given the package "repo/pkg/billing/domain/service" imports "repo/internal/domain/service" as "commonservice"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a blank import — ok, not an entity reference
    Given the package "repo/pkg/billing/server/handler" imports "repo/internal/domain/service" as "_"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: outside pkg/<module>, internal-to-internal import — out of scope, ok
    Given the package "repo/internal/domain/usecase" imports "repo/internal/domain/service" without an alias
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: settings.prefix replaces the default "common" prefix
    Given settings.prefix: "shared" is set in .golangci.yml
    And the package "custom/pkg/billing/domain/usecase" imports "custom/internal/domain/service" as "commonservice"
    When the analyzer checks the file
    Then a "GID-240" diagnostic is reported on the import "custom/internal/domain/service"

  Scenario: settings.prefix — an alias with the custom prefix is ok
    Given settings.prefix: "shared" is set in .golangci.yml
    And the package "custom/pkg/billing/domain/service" imports "custom/internal/domain/service" as "sharedservice"
    When the analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the RULES.md registry
#  [x] Layer chosen: go/analysis (import-path segments and alias name are needed)
#  [x] Severity and message are defined
#  [x] Case classes covered: positive, negative, boundary, non-applicability (dot-import/blank-import)
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml
