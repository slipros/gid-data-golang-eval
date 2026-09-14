# language: en

Feature: GID-275 — an import of the module's own package is aliased only when the file needs it
  As a developer
  I want an import of my module's package to keep the package's own name
  unless something in the file already takes that name
  So that a reader does not decode domainserviceconvert where convert is the only convert in sight

  Scenario: an alias nothing calls for — violation
    Given the file imports "example.com/svc/internal/domain/service/convert" as "serviceconvert"
    And nothing else in the file is called "convert"
    When the analyzer checks the file
    Then a "GID-275" diagnostic is reported on the import, suggesting to drop the alias and refer to it as "convert"

  Scenario: an alias spelling the name the path already suggests — violation
    Given the file imports "example.com/svc/internal/domain/model" as "model"
    And the file imports "example.com/svc/internal/client/billing/v2" (package billing) as "billing"
    When the analyzer checks the file
    Then a "GID-275" diagnostic is reported on each import

  Scenario: an alias hiding a package name the path does not suggest — violation, fix names the real package
    Given the file imports "example.com/svc/internal/event/v1" (package eventv1) as "ev"
    When the analyzer checks the file
    Then a "GID-275" diagnostic suggests importing it as eventv1 "example.com/svc/internal/event/v1"

  Scenario: the real package name goimports writes — ok
    Given the file imports "example.com/svc/internal/event/v1" (package eventv1) as "eventv1"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: another import takes the package name — ok
    Given the file imports "example.com/svc/internal/dal/repository/convert"
    And the file imports "example.com/svc/internal/domain/service/convert" as "serviceconvert"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: two aliased packages of the same name justify each other — ok
    Given the file imports "…/dal/repository/convert" as "repoconvert" and "…/domain/service/convert" as "serviceconvert"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a package-level declaration takes the package name — ok
    Given the package declares "type model struct{}"
    And the file imports "example.com/svc/internal/domain/model" as "entitymodel"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a package named like a predeclared identifier — ok
    Given the file imports a package named "error" as "modelerror" and uses the built-in error type
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a local name shadows the package where the file uses it — ok
    Given "func New(model int) int { return model + domainmodel.X }"
    When the analyzer checks the file
    Then no diagnostic is reported on the "domainmodel" import

  Scenario: boundary — a shadow in a function that does not use the import justifies nothing
    Given "func Other() int { return usecasemodel.X }" and "func Third(model int) int { return model }"
    When the analyzer checks the file
    Then a "GID-275" diagnostic is reported on the "usecasemodel" import

  Scenario: boundary — pkg/<module> importing the shared internal/** keeps the GID-240 alias
    Given the package "example.com/svc/pkg/billing/domain/usecase" imports "example.com/svc/internal/domain/service" as "commonservice"
    And the same file imports its own "example.com/svc/pkg/billing/domain/service" as "billingservice"
    When the analyzer checks the file
    Then no diagnostic is reported on "commonservice"
    And a "GID-275" diagnostic is reported on "billingservice" — the GID-240 import holds only its alias

  Scenario: boundary — a common prefix on the module's own package marks nothing
    Given the package "example.com/svc/pkg/billing/domain/usecase" imports "example.com/svc/pkg/billing/domain/model" as "commonmodel"
    When the analyzer checks the file
    Then a "GID-275" diagnostic is reported on the import

  Scenario: boundary — a common prefix outside pkg/<module> marks nothing
    Given the package "example.com/svc/internal/app" imports "example.com/svc/internal/domain/model" as "commonmodel"
    When the analyzer checks the file
    Then a "GID-275" diagnostic is reported on the import

  Scenario: a nested module borrows its parent module's package under the common prefix — ok
    Given the package "example.com/svc/pkg/integration/push/firebase/domain/service" imports "example.com/svc/pkg/integration/domain/model" as "commonmodel"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: settings.prefix replaces the common marker
    Given settings.prefix: "shared" is set in .golangci.yml
    When a nested module imports its parent's model as "commonmodel"
    Then a "GID-275" diagnostic is reported
    But "sharedmodel" is not reported

  Scenario: a _test.go file is judged
    Given "cabinet_test.go" imports "example.com/svc/internal/domain/model" as "domainmodel" with no collision
    When the analyzer checks the file
    Then a "GID-275" diagnostic is reported

  Scenario: non-applicability — another module's import keeps its own convention
    Given the file imports "example.com/lib/errs" as "gderrs"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — blank, dot imports and generated code
    Given a blank import, a dot import, and a generated file with a redundant alias
    When the analyzer checks the files
    Then no diagnostic is reported

  Scenario: non-applicability — no module information
    Given the packages are loaded in GOPATH mode
    When the analyzer checks a redundant alias
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the RULES.md registry
#  [x] Layer chosen: go/analysis (the real package name, scopes and the module path are needed)
#  [x] Severity and message are defined
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml, gid-golangci.yml, gid-golangci-rules.yml
