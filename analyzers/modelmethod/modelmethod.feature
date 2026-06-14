# language: en

Feature: GID-195 — model behavior lives as a model method
  As a developer
  I want private service/usecase functions working with a single
  value of a model type to become public methods of that type in model
  So that model behavior belongs to the model itself rather than being smeared
  across business layers as helpers

  # Semantics (fixed by the requirement of 2026-06-07):
  # - scope: the roots of /domain/service and /domain/usecase (EndsWith) — the
  #   convert/ and repository subpackages are not affected;
  # - trigger: a private function with strictly one parameter T or *T, where
  #   T is a named type of the model layer (struct, enum; not an interface);
  # - a private method NOT using its receiver is the same case;
  # - non-movable ones are not affected: a method accessing its receiver; a function
  #   referencing package-level symbols of its own package (including package types
  #   in the results);
  # - complements GID-133: that one pushes private functions into methods of their
  #   struct, GID-195 refines — if the function is entirely about a model value,
  #   its home is the model type itself;
  # - exclusions: //nolint:gidmodelmethod or settings.exclude
  #   ("Function" | "Type.Method").

  Scenario: positive — a private function over a model struct
    Given the package path ends with the segments "domain/service"
    And the function "snapshotTitle(s *model.Snapshot) string" using only "s" is declared
    When the analyzer checks the package
    Then a "GID-195" diagnostic is reported with a hint to make it a public method of model.Snapshot

  Scenario: positive — a model enum by value
    Given the function "isDone(st model.Status) bool" is declared
    When the analyzer checks the package
    Then a "GID-195" diagnostic about model.Status is reported

  Scenario: positive — a method not using its receiver
    Given the method "(s *SnapshotService) renderSnapshot(snap *model.Snapshot) string" is declared
    And the body does not access "s"
    When the analyzer checks the package
    Then a "GID-195" diagnostic about the method is reported

  Scenario: positive — an unnamed receiver
    Given the method "(*SnapshotService) pingSnapshot(s *model.Snapshot) bool" is declared
    When the analyzer checks the package
    Then a "GID-195" diagnostic about the method is reported

  Scenario: negative — the method uses its receiver
    Given the method "decorate" reads "s.prefix"
    When the analyzer checks the package
    Then no diagnostic is reported
    # the method legitimately belongs to the struct

  Scenario: negative — the function depends on its own package
    Given the function "tagSnapshot" uses a package-level constant of the package
    When the analyzer checks the package
    Then no diagnostic is reported
    # non-movable into model without moving the dependencies

  Scenario: negative — the result contains a type of its own package
    Given the function "wrapSnapshot(s *model.Snapshot) *SnapshotService"
    When the analyzer checks the package
    Then no diagnostic is reported
    # moving it into model would create a reverse dependency model → service

  Scenario: boundary — not a single value
    Given "equalSnapshots(a, b *model.Snapshot)", "joinSnapshots(...model.Snapshot)" and "firstName([]model.Snapshot)" are declared
    When the analyzer checks the package
    Then no diagnostic is reported

  Scenario: boundary — an interface of the model layer
    Given the function "validateAny(v model.Validator)" is declared
    When the analyzer checks the package
    Then no diagnostic is reported
    # a method cannot be added to an interface

  Scenario: boundary — a parameter of a type from its own package
    Given the function "optionsName(o *ServiceOptions)" is declared
    When the analyzer checks the package
    Then no diagnostic is reported

  Scenario: boundary — a generic function and an exported function
    Given "anyTitle[T any](v T)" and "TitleSnapshot(s *model.Snapshot)" are declared
    When the analyzer checks the package
    Then no diagnostic is reported

  Scenario: non-applicability — /dal/repository
    Given the package path ends with the segments "dal/repository"
    And a private function with a single model parameter is declared
    When the analyzer checks the package
    Then no GID-195 diagnostic is reported

  Scenario: non-applicability — the convert subpackage of service
    Given the package path ends with the segments "domain/service/convert"
    When the analyzer checks the package
    Then no GID-195 diagnostic is reported

  Scenario: non-applicability — settings.exclude
    Given the linter setting "exclude: [legacyTitle, Service.legacyRender]"
    When the analyzer checks the package
    Then no diagnostic is reported on "legacyTitle" and "Service.legacyRender"
    And other violations of the package are still caught

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (GID-195, belongs-to-model)
#  [x] Layer chosen: go/analysis (parameter types and TypesInfo.Uses are needed)
#  [x] Severity and message are defined ("GID-195: ... make it a public method of this type")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
