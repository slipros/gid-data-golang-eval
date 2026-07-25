# language: en

Feature: GID-214 — the logger is created once in the composition root
  As a developer
  I want the global logger to be created or grabbed only in the composition root
  So that there is a single configured logger and the layers receive a ready one via the constructor
  Source: libs.md (do not create new logger instances, pass the existing one along)
  Scope: all packages EXCEPT package main and composition-root packages (the path contains internal/app)
  The rule is stack-agnostic: logrus.New()/StandardLogger() and slog.New()/Default()/SetDefault()
  are banned alike. The package is resolved by import path via types; _test.go and generated
  files are skipped.

  # --- Positive class: the violation is caught ---

  Scenario: logrus.New() in /domain/service — violation
    Given "logrus.New()" is called in /domain/service
    When the analyzer checks the file
    Then a "GID-214" diagnostic is reported with the text "only in the composition root"

  Scenario: logrus.StandardLogger() in /dal/repository — violation
    Given "logrus.StandardLogger()" is called in /dal/repository
    When the analyzer checks the file
    Then a "GID-214" diagnostic is reported with the text "only in the composition root"

  Scenario: slog.New() in /domain/service — violation
    Given "slog.New(slog.NewTextHandler(...))" is called in /domain/service
    When the analyzer checks the file
    Then a "GID-214" diagnostic is reported with the text "only in the composition root"

  Scenario: slog.Default()/slog.SetDefault() outside the composition root — violation
    Given "slog.Default()" or "slog.SetDefault(...)" is called in /domain/service
    When the analyzer checks the file
    Then a "GID-214" diagnostic is reported with the text "only in the composition root"
    # slog.Default() is the same smell as logrus.StandardLogger().

  # --- Negative class: clean code passes ---

  Scenario: logrus.New() in package main — ok
    Given "logrus.New()" is called in package main
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: logrus.New() in internal/app — ok
    Given "logrus.New()" is called in a package whose path contains internal/app
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a ready *slog.Logger arrives through the constructor — ok
    Given a constructor in /domain/service takes "logger *slog.Logger" and calls "logger.With(...)"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: logger.WithField in /domain/service — not an instance creation, ok
    Given "logger.WithField(...)" is called in /domain/service
    When the analyzer checks the file
    Then no diagnostic is reported

  # --- Boundary class ---

  Scenario: New() from another package named logrus — resolved by import path, ok
    Given a foreign package "logrus" with a different import path is imported and "logrus.New()" is called
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a generated file with logrus.New() — skipped
    Given the file is marked "// Code generated ... DO NOT EDIT."
    When the analyzer checks the file
    Then no diagnostic is reported

  # --- Non-applicability class ---

  Scenario: logrus.New() in _test.go — the rule does not apply
    Given "logrus.New()" is called in a "*_test.go" file in /domain/handler
    When the analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-214)
#  [x] Layer chosen: go/analysis (LoadModeTypesInfo — logrus/slog resolved by import path)
#  [x] Severity and message are defined ("GID-214: ...")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml (outside the scope of this task)
