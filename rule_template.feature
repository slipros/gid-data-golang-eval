# language: en
# TEMPLATE for a single rule. Copy and fill in for the actual rule.
# The example below is a placeholder RULE-002 (context.Context as the first argument).

Feature: RULE-002 — context.Context is always the first argument
  As a developer
  I want context.Context to be the first parameter of a function
  So that a unified style and the Go convention are followed

  Scenario: ctx as the first argument — ok
    Given a function with the signature "func Do(ctx context.Context, id int)"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: ctx is not the first argument — violation
    Given a function with the signature "func Do(id int, ctx context.Context)"
    When the analyzer checks the file
    Then a "RULE-002" diagnostic is reported on the "ctx" parameter

  Scenario: ctx is absent — the rule does not apply
    Given a function with the signature "func Do(id int)"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: method with a receiver — ctx after the receiver counts as first
    Given a method with the signature "func (s *Svc) Do(ctx context.Context, id int)"
    When the analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [ ] ID and description are recorded in the PRD registry (section 5)
#  [ ] Layer chosen: ruleguard (simple) or go/analysis (complex)
#  [ ] Severity and message are defined
#  [ ] Case classes covered: positive, negative, boundary, non-applicability
#  [ ] testdata with // want for analysistest (if go/analysis)
#  [ ] Rule enabled in .golangci.yml
