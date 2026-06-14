# Eval spec for rule GID-231 (fsm-map-unexported), linter gidfsmmap.
# Source: go-styleguide model.md, section "Enum and State Machine (FSM)":
# the transition map is an UNEXPORTED package-level variable
# var <entity>StatusTransitions; consumers use the CanTransitionTo method.

Feature: GID-231 — FSM transition map in /domain/model must be unexported
  As a domain model developer
  I want FSM transition maps to be unexported package-level variables
  So that state transitions are only reachable through CanTransitionTo (model.md)

  # --- positive: violation is caught ---

  Scenario: exported map[E][]E over a local string enum — violation
    Given package /domain/model declares "var SnapshotStatusTransitions = map[SnapshotStatus][]SnapshotStatus{...}"
    And "SnapshotStatus" is a string enum (named string type with const values) of the same package
    When the analyzer checks the file
    Then a "GID-231" diagnostic is reported on the var name "SnapshotStatusTransitions"

  Scenario: exported map[E]map[E]struct{} (set form) — violation
    Given package /domain/model declares "var StatusTransitionSet = map[SnapshotStatus]map[SnapshotStatus]struct{}{...}"
    When the analyzer checks the file
    Then a "GID-231" diagnostic is reported on the var name "StatusTransitionSet"

  Scenario: exported map[E]map[E]bool (flag form) — violation
    Given package /domain/model declares "var StatusTransitionFlags = map[SnapshotStatus]map[SnapshotStatus]bool{...}"
    When the analyzer checks the file
    Then a "GID-231" diagnostic is reported on the var name "StatusTransitionFlags"

  Scenario: exported var of a named map type with transition-map underlying — violation
    Given package /domain/model declares "type transitionMap map[SnapshotStatus][]SnapshotStatus" and "var NamedTransitions = transitionMap{...}"
    When the analyzer checks the file
    Then a "GID-231" diagnostic is reported on the var name "NamedTransitions"

  # --- negative: clean code passes ---

  Scenario: unexported transition map — ok
    Given package /domain/model declares "var snapshotStatusTransitions = map[SnapshotStatus][]SnapshotStatus{...}"
    When the analyzer checks the file
    Then no diagnostic is reported

  # --- boundary ---

  Scenario: exported map with non-enum value — not a transition map
    Given package /domain/model declares "var StatusLabels = map[SnapshotStatus][]string{...}"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: exported map[E][]F over two different enums — not a transition map
    Given package /domain/model declares "var MixedTransitions = map[SnapshotStatus][]OtherStatus{...}" where both are local enums
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: exported map over an int-based enum — not a string enum, rule does not apply
    Given package /domain/model declares "type Priority int" with const values and "var PriorityTransitions = map[Priority][]Priority{...}"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: exported map over a string type without const values — not an enum
    Given package /domain/model declares "type RawTag string" without consts and "var RawTransitions = map[RawTag][]RawTag{...}"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: exported map[E]map[E]int — inner value is neither struct{} nor bool
    Given package /domain/model declares "var StatusMatrix = map[SnapshotStatus]map[SnapshotStatus]int{...}"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: generated file — skipped
    Given a generated file in /domain/model declares "var GeneratedTransitions = map[SnapshotStatus][]SnapshotStatus{...}"
    When the analyzer checks the file
    Then no diagnostic is reported

  # --- non-applicability ---

  Scenario: exported transition map outside /domain/model — rule does not apply
    Given package /domain/service declares "var JobStatusTransitions = map[JobStatus][]JobStatus{...}" over a local enum
    When the analyzer checks the file
    Then no diagnostic is reported

# --- Out of scope (deliberate) ---
# Requiring the transition map and the enum to live in the SAME FILE is not
# enforced: a model package may legitimately split a large model across files
# (model.go + status.go), and the styleguide only fixes the package layout,
# not file colocation. Enforcing it would produce false positives without a
# deterministic source in model.md.
# Enums imported from another package are not considered: a /domain/model
# transition map over a foreign enum is already a smell covered by layer
# rules, and enum detection here is intentionally package-local.

# --- Checklist for adding a new rule ---
#  [x] ID and description registered in RULES.md
#  [x] Layer chosen: go/analysis (type info needed to detect enums and map shapes)
#  [x] Severity and message defined
#  [x] Cases covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
