# language: en

Feature: GID-170 — domain and dal do not import the event layer
  As a service architect
  I want /domain/** and /dal/** packages not to import /event/**
  So that the event layer (kafka producer/consumer, DTOs) depends on domain/model
  and converts model <-> DTO, not the other way around

  Scenario: domain imports event/dto — violation
    Given the package "svc/domain/notifier" imports "svc/event/dto"
    When the analyzer checks the file
    Then a "GID-170" diagnostic is reported on the import "svc/event/dto"

  Scenario: dal imports event/dto — violation
    Given the package "svc/dal/outbox" imports "svc/event/dto"
    When the analyzer checks the file
    Then a "GID-170" diagnostic is reported on the import "svc/event/dto"

  Scenario: event imports domain/model — ok
    Given the package "svc/event/producer" imports "svc/domain/model"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: the segment "events" (plural) — the rule does not apply
    Given the package "svc/domain/boundary" imports "svc/events/dto"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: the segment "event-api" — not equal to the segment "event", the rule does not apply
    Given the package "svc/domain/boundary" imports "svc/event-api/contract"
    When the analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the PRD registry (section 5)
#  [x] Layer chosen: go/analysis (import-path segments are needed)
#  [x] Severity and message are defined
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml
