# language: en

Feature: GID-197 — an interface contains only used methods
  As a developer
  I want dependency interfaces in service/usecase/repository/server/event
  not to contain methods the consumer does not use
  So that the dependency contract reflects the real needs of the code,
  not the capabilities of the implementation

  # Semantics (moved from "Partially checkable", requirement of 2026-06-07):
  # - relies on GID-134: the interface is declared at the consumer, so all
  #   usages of its methods are visible in the package;
  # - usage = a call or a method value outside *_test.go;
  #   a "tests-only" method is a violation (the contract describes production);
  # - embedded interfaces are not checked (explicit methods only);
  #   via embedding a reference resolves to the same method object — it counts;
  # - FP-safe escape: an interface value in an untrackable context
  #   (assignment/argument/return under a different type, type assertion,
  #   generic constraint, unknown context) — the interface is skipped;
  # - exclusions: //nolint:gidifacemin or settings.exclude
  #   ("Interface" as a whole | "Interface.Method").

  Scenario: positive — an unused method of a dependency interface
    Given the package path ends with the segments "domain/service"
    And the interface "SnapshotRepository" contains the method "DeleteSnapshot"
    And there is neither a call nor a method value of "DeleteSnapshot" in the package
    When the analyzer checks the package
    Then a "GID-197" diagnostic is reported on the method "DeleteSnapshot"

  Scenario: positive — a method used only from *_test.go
    Given the method "Ping" of the interface "SnapshotProbe" is called only in a test
    When the analyzer checks the package
    Then a "GID-197" diagnostic is reported on the method "Ping"

  Scenario: negative — all methods are called
    Given the interface methods are called via a field of the consumer struct
    When the analyzer checks the package
    Then no diagnostic is reported

  Scenario: negative — a method value
    Given the method "Warm" is taken as a method value "warm := c.cache.Warm"
    When the analyzer checks the package
    Then no diagnostic is reported

  Scenario: boundary — the interface value escapes under a different type
    Given an "AuditSink" value is assigned to a variable of type "any"
    And the method "Flush" is never called
    When the analyzer checks the package
    Then no diagnostic is reported
    # method consumption cannot be tracked — the interface is skipped entirely

  Scenario: boundary — an embedded interface of the same package
    Given "snapshotReadWriter" embeds "snapshotReader"
    And "ReadSnapshot" is called via a "snapshotReadWriter" value
    When the analyzer checks the package
    Then no diagnostic is reported
    # the method object is shared — the usage counts for both

  Scenario: boundary — an embedded standard-library interface
    Given the interface embeds "io.Closer"
    When the analyzer checks the package
    Then "Close" is not checked — explicit methods only

  Scenario: boundary — a generated file
    Given the interface is declared in "zz_generated.go"
    When the analyzer checks the package
    Then no diagnostic is reported

  Scenario: non-applicability — the model layer
    Given the package path ends with the segments "domain/model"
    And the interface contains unused methods
    When the analyzer checks the package
    Then no GID-197 diagnostic is reported
    # model interfaces may describe a contract for external consumers

  Scenario: non-applicability — settings.exclude
    Given the linter setting "exclude: [LegacyGateway, AlertSink.Flush]"
    When the analyzer checks the package
    Then "LegacyGateway" is skipped entirely, "AlertSink.Flush" — pointwise
    And other violations of the package are still caught

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (GID-197, interface-minimal)
#  [x] Layer chosen: go/analysis (method objects, escape analysis of values)
#  [x] Severity and message are defined ("GID-197: method %q of interface %q ...")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
