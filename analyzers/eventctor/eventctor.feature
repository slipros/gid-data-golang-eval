# language: en
# GID-216 — event-ctor-deps (linter gideventctor). Source: event.md.

Feature: GID-216 — dependencies of event-layer constructors
  As an event-layer developer
  I want consumer constructors to take a logrus logger,
  and producer constructors not to
  So that the consumer builds an Entry with broker/consumer fields, and the producer
  propagates errors to the calling code

  Scenario: a consumer constructor without a logger — violation (positive)
    Given a package with the segments event and consumer
    And the constructor "func NewOrderConsumer(svc Service) *OrderConsumer"
    When the analyzer checks the file
    Then a "GID-216" diagnostic is reported saying that a consumer takes *logrus.Logger

  Scenario: a producer constructor with *logrus.Logger — violation (positive)
    Given a package with the segments event and producer
    And the constructor "func NewOrderProducer(log *logrus.Logger) *OrderProducer"
    When the analyzer checks the file
    Then a "GID-216" diagnostic is reported saying that a producer does not take a logger

  Scenario: a consumer constructor with *logrus.Logger — ok (negative)
    Given a package with the segments event and consumer
    And the constructor "func NewPaymentConsumer(log *logrus.Logger) *PaymentConsumer"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a consumer constructor with *logrus.Entry — ok (negative)
    Given a package with the segments event and consumer
    And the constructor "func NewRefundConsumer(log *logrus.Entry) *RefundConsumer"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a producer constructor without a logger — ok (negative)
    Given a package with the segments event and producer
    And the constructor "func NewPaymentProducer(svc Service) *PaymentProducer"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a schema function returns a foreign package type — not a constructor (boundary)
    Given a package with the segments event and consumer
    And the function "func NewOrderCreatedSchema() *registry.Schema" returns a type of a foreign package
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: an unexported helper — not a constructor (boundary)
    Given a package with the segments event and consumer
    And the function "func newHelper() *helper"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a validator in event/kafka/consumer/validate — not a consumer (boundary)
    Given a package with the segments event, consumer and validate
    And the constructor "func NewOrderValidator() *OrderValidator"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a constructor outside the event layer — the rule does not apply (non-applicability)
    Given a package in the domain/service layer
    And the constructor "func NewService() *Service"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a constructor listed in settings.exclude — skipped (non-applicability)
    Given a package with the segments event and consumer
    And the constructor "func NewLegacyConsumer() *LegacyConsumer" is listed in settings.exclude
    When the analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md)
#  [x] Layer chosen: go/analysis (complex — types needed)
#  [x] Severity and message are defined
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
