# language: en

Feature: GID-196 — call chains formatted one call per line
  As a developer
  I want call chains of min-calls links or more to be formatted
  one call per line, including the first one
  So that chains read vertically and edit diffs stay focused

  # Semantics (moved from "Partially checkable", requirement of 2026-06-07):
  # - a link is a call via a selector whose receiver is another call
  #   (including via intermediate fields: a.B().c.D());
  # - the threshold is configured by settings.min-calls, default 2 (same as GID-156);
  # - "including the first" — the first call must not be on the line of the base expression;
  # - type conversions (model.Status(v)) do not count as links;
  # - logrus chains are the domain of GID-156 (gidlogchain), skipped;
  # - *_test.go and generated code are not checked;
  # - exclusions: //nolint:gidchainperline.

  Scenario: positive — a chain of 2 calls on one line
    Given the expression "strings.NewReplacer(a, b).Replace(s)" on a single line
    When the analyzer checks the file
    Then a "GID-196" diagnostic is reported on the first call

  Scenario: positive — the first call on the base line
    Given the chain is broken across lines but the first call stayed on the base line
    When the analyzer checks the file
    Then a "GID-196" diagnostic is reported on the first call

  Scenario: positive — two links on one line inside a multi-line chain
    Given "job().name()" share one line inside a multi-line chain
    When the analyzer checks the file
    Then a "GID-196" diagnostic is reported on the second of them

  Scenario: positive — a chain via an intermediate field
    Given the expression "s.r.job().name()" on a single line
    When the analyzer checks the file
    Then a "GID-196" diagnostic is reported

  Scenario: negative — each call on its own line
    Given the chain is formatted one call per line, including the first
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — a single inline call
    Given the expression "strings.ToUpper(x)"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — a nested call is not a chain
    Given the expression "f(g(x).h())" — a chain nested in an argument
    When the analyzer checks the file
    Then a diagnostic is reported only on the inner chain "g(x).h()"

  Scenario: boundary — a conversion is not a link
    Given the expression "sub.Code(v).Upper()"
    When the analyzer checks the file
    Then no diagnostic is reported
    # the conversion is the base of the chain, only one call remains

  Scenario: boundary — a call on a function result
    Given the expression "factory().name()"
    When the analyzer checks the file
    Then no diagnostic is reported
    # factory() is a function call, the base of the chain; only one link

  Scenario: boundary — threshold min-calls: 3
    Given the linter setting "min-calls: 3"
    And an expression of 2 links on a single line
    When the analyzer checks the file
    Then no diagnostic is reported
    And a chain of 3 links on a single line is caught

  Scenario: non-applicability — a logrus chain
    Given the expression "l.WithField(a, 1).Info(x)"
    When the analyzer checks the file
    Then no GID-196 diagnostic is reported
    # multi-line formatting of logrus is the domain of GID-156

  Scenario: non-applicability — *_test.go
    Given an inline chain in a "*_test.go" file
    When the analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (GID-196, chain-call-per-line)
#  [x] Layer chosen: go/analysis (types needed: conversions, logrus methods)
#  [x] Severity and message are defined ("GID-196: a chain of N calls ...")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
