# language: en

Feature: GID-235 — convert-purity: a converter is a pure function over vocabulary types
  As a service architect
  I want /**/convert packages not to import business layers or side-effect libraries
  So that a converter stays a pure function operating only on vocabulary types
    (model/entity/dto/client/pb)

  Scenario: convert imports the neighboring domain/service layer — violation
    Given the package "svc/domain/service/convert" imports "svc/domain/service"
    When the analyzer checks the file
    Then a "GID-235" diagnostic is reported on the import "svc/domain/service"

  Scenario: convert imports domain/usecase — violation
    Given the package "svc/event/consumer/convert" imports "svc/domain/usecase"
    When the analyzer checks the file
    Then a "GID-235" diagnostic is reported on the import "svc/domain/usecase"

  Scenario: convert imports the neighboring dal/repository layer — violation
    Given the package "svc/dal/repository/convert" imports "svc/dal/repository"
    When the analyzer checks the file
    Then a "GID-235" diagnostic is reported on the import "svc/dal/repository"

  Scenario: convert imports a banned third-party logging library — violation
    Given the package "svc/server/grpc/handler/convert" imports "github.com/sirupsen/logrus"
    When the analyzer checks the file
    Then a "GID-235" diagnostic is reported on the import "github.com/sirupsen/logrus"

  Scenario: convert imports domain/model, dal/entity, client/*, event/dto and stdlib — ok
    Given the package "svc/mapper/convert" imports "svc/domain/model", "svc/dal/entity", "svc/client/billing", "svc/event/dto" and "time"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: convert imports event/dto — the event/dto exception applies, ok
    Given the package "svc/event/consumer/convert" imports "svc/event/dto"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: package path ends with "xconvert", not the exact "convert" segment — the rule does not apply
    Given the package "svc/convertx/xconvert" imports "svc/domain/service"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: "convert" is a middle segment, the package path ends with "util" — the rule does not apply
    Given the package "svc/domain/service/convert/util" imports "svc/dal/repository"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: an ordinary (non-convert) package imports domain/service — the rule does not apply
    Given the package "svc/app/wiring" imports "svc/domain/service"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: settings.packages replaces the default third-party ban list
    Given the analyzer is configured with settings.packages = ["example.com/inhouse/somelib"]
    And the package "custom/adapter/convert" imports "example.com/inhouse/somelib" and "github.com/sirupsen/logrus"
    When the analyzer checks the file
    Then a "GID-235" diagnostic is reported on the import "example.com/inhouse/somelib"
    And no diagnostic is reported on the import "github.com/sirupsen/logrus"

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the PRD registry (section 5)
#  [x] Layer chosen: go/analysis (import-path segments are needed)
#  [x] Severity and message are defined
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml
