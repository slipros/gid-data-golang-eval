# language: en

Feature: GID-251 — the logger is not configuration (gidloggerplace)
  As a developer
  I want the logger to be a separate constructor parameter, never a field of an options struct
  So that the constructor signature states the dependency and nobody falls back to the global logger

  # One analyzer over type declarations, LoadModeTypesInfo (the field type is
  # resolved through internal/lgr, so both stacks count: logrus types including
  # its interfaces, and *slog.Logger).
  # Scope: struct types whose name ends with one of settings.suffixes
  # (default: Options, Config, Settings), in any layer.
  # Generated code (ast.IsGenerated) is skipped.
  # Rationale: a logger in Options erases the dependency from the signature —
  # GID-153 (logger after opts) has nothing to place — and invites the
  # "nil means slog.Default()" fallback, which is GID-214's territory.

  # --- Class 1: positive ---

  Scenario: positive — a slog logger in an options struct
    Given a struct "SinkOptions" with a field "Logger *slog.Logger"
    When the gidloggerplace analyzer checks the file
    Then the diagnostic "GID-251: options struct \"SinkOptions\" holds a logger — a logger is a dependency, not configuration. Fix: drop the field and take the logger as a separate constructor parameter, after opts (GID-153); the entity names itself on it in the constructor (GID-154)" is reported

  Scenario: positive — a logrus logger in a Config struct
    Given a struct "CacheConfig" with a field "Logger *logrus.Entry"
    When the gidloggerplace analyzer checks the file
    Then the diagnostic "GID-251: …" is reported

  # --- Class 2: negative ---

  Scenario: negative — options without a logger, the logger is a parameter
    Given "SenderOptions" holds only configuration and "NewSender(opts SenderOptions, logger *slog.Logger)" takes the logger separately
    When the gidloggerplace analyzer checks the file
    Then no diagnostic is reported

  # --- Class 3: boundary ---

  Scenario: boundary — an interface-typed logger field
    Given a struct "ReporterSettings" with a field "Logger logrus.FieldLogger"
    When the gidloggerplace analyzer checks the file
    Then the diagnostic "GID-251: …" is reported
    # The type is resolved through internal/lgr, so a logger interface counts too
    # (such a field is also GID-217's territory — the interface carries no context).

  Scenario: boundary — a struct without the options suffix keeps its logger
    Given a struct "Sink" with a field "logger *slog.Logger"
    When the gidloggerplace analyzer checks the file
    Then no diagnostic is reported
    # The entity holds its logger; that is exactly what GID-154 works with.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — an options struct of plain configuration
    Given a struct "PlainOptions" with fields "Timeout time.Duration" and "Debug bool"
    When the gidloggerplace analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — settings.suffixes replaces the default list
    Given settings.suffixes contains only "Params" and the file declares "SinkOptions" with a logger
    When the gidloggerplace analyzer checks the file
    Then no diagnostic is reported
    # Only "SinkParams" would be checked in that configuration.

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-251)
#  [x] Layer chosen: go/analysis (package loggerplace)
#  [x] Message is defined ("GID-251: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml
