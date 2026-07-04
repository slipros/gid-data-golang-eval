# language: en
# Specification of rule GID-193 (no-pkg-stutter).
# Linter: gidpkgstutter (go/analysis, LoadModeSyntax — no types needed).

Feature: GID-193 — ban on the "package.SymbolWithPackageName" stutter
  As a backend-go service developer
  I want an exported symbol not to repeat the name of its package
  So that from outside the code reads as widget.Options, not widget.WidgetOptions

  # Top-level exported symbols are checked: types, functions, vars, consts.
  # The package name is compared with the first CamelCase word of the symbol case-insensitively.
  # A match counts only at a word boundary: after a prefix of length
  # len(pkgName) an uppercase letter (the next word) must begin.
  # Exceptions: New* constructors, methods (with a receiver), unexported
  # symbols, package main. Generated files are skipped.

  # --- Positive cases (the violation is caught) ---

  Scenario: positive — the type WidgetOptions in the widget package
    Given the package "widget" with the type "WidgetOptions"
    When the analyzer checks the package
    Then the diagnostic "GID-193: WidgetOptions repeats the package name widget. Fix: from outside it is widget.Options; drop the prefix" is reported

  Scenario: positive — the function WidgetCount
    Given the package "widget" with the function "WidgetCount"
    When the analyzer checks the package
    Then the diagnostic "GID-193: WidgetCount repeats the package name widget. Fix: from outside it is widget.Count; drop the prefix" is reported

  Scenario: positive — the var WidgetDefault
    Given the package "widget" with the variable "WidgetDefault"
    When the analyzer checks the package
    Then the diagnostic "GID-193: WidgetDefault repeats the package name widget. Fix: from outside it is widget.Default; drop the prefix" is reported

  Scenario: positive — the const WidgetMax
    Given the package "widget" with the constant "WidgetMax"
    When the analyzer checks the package
    Then the diagnostic "GID-193: WidgetMax repeats the package name widget. Fix: from outside it is widget.Max; drop the prefix" is reported

  # --- Negative cases (clean code passes) ---

  Scenario: negative — the type Options without a prefix
    Given the package "widget" with the type "Options"
    When the analyzer checks the package
    Then no diagnostic is reported

  Scenario: negative — the function Count without a prefix
    Given the package "widget" with the function "Count"
    When the analyzer checks the package
    Then no diagnostic is reported

  # --- Boundary cases ---

  Scenario: boundary — the constructor NewWidget is exempt in favor of GID-104
    Given the package "widget" with the function "NewWidget"
    When the analyzer checks the package
    Then no diagnostic is reported

  Scenario: boundary — the type Logger in the log package (word boundary)
    Given the package "log" with the type "Logger"
    When the analyzer checks the package
    Then no diagnostic is reported
    # (log is only a prefix of the word Logger, not a separate CamelCase word.)

  Scenario: boundary — the method (w *Widget) WidgetID() is not matched
    Given the package "widget" with the method "WidgetID" on the receiver Widget
    When the analyzer checks the package
    Then no diagnostic is reported

  Scenario: boundary — the unexported widgetCache is not matched
    Given the package "widget" with the variable "widgetCache"
    When the analyzer checks the package
    Then no diagnostic is reported

  Scenario: boundary — the symbol Widget exactly equals the package name (no next word)
    Given the package "widget" with the type "Widget"
    When the analyzer checks the package
    Then no diagnostic is reported

  # --- Non-applicability (the rule does not apply) ---

  Scenario: non-applicability — package main
    Given the package "main" with the type "MainOptions"
    When the analyzer checks the package
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-193)
#  [x] Layer chosen: go/analysis, LoadModeSyntax (no types needed — AST and the package name suffice)
#  [x] Severity and message are defined ("GID-193: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml

  # --- Suffix stutter (requirement 2026-07-04) ---

  Scenario: type ends with the package name — violation
    Given package "repository" declares "type SnapshotRepository struct{}"
    When the analyzer checks the file
    Then a "GID-193" diagnostic suggests naming the symbol after the entity ("repository.Snapshot")

  Scenario: dependency interface with another role's suffix — ok
    Given package "service" declares "type SnapshotRepository interface{...}"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: package name is only the tail of a word — ok
    Given package "story" declares "type History struct{}"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: exact match of the package name — ok
    Given package "repository" declares "type Repository struct{}"
    When the analyzer checks the file
    Then no diagnostic is reported (reads like time.Time)
