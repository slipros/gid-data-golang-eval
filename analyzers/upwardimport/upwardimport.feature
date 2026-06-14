# language: en

Feature: GID-131 — a child package does not import its parent
  As a service architect
  I want a child package not to import its parent package
  So that dependencies flow downward: shared code is pushed down, and the parent imports
  the children, not the other way around

  Scenario: a child package imports its parent — violation
    Given the package "app/parent/badchild" imports "app/parent"
    And "app/parent" is a strict segment-wise prefix of "app/parent/badchild"
    When the analyzer checks the file
    Then a "GID-131" diagnostic is reported on the import "app/parent"

  Scenario: the parent imports a child — the correct direction, ok
    Given the package "app/parent" imports "app/parent/child"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a child package imports a sibling — not the parent, ok
    Given the package "app/parent/badchild" imports "app/parent/other"
    And "app/parent/other" is not a prefix of "app/parent/badchild"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: "app/parentx" is not a child of "app/parent" — the prefix is segment-wise
    Given the package "app/parentx" imports "app/parent"
    And "app/parent" is a string prefix but not a segment-wise prefix of "app/parentx"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a leaf package without imports of its own module — the rule does not apply
    Given the package "app/parent/child" does not import its own module
    When the analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, the GID-131 row)
#  [x] Layer chosen: go/analysis (pass.Pkg.Path() and import-path segments are needed)
#  [x] Severity and message are defined
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml
