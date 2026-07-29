# language: en

Feature: GID-154 — a constructor with a logger names its entity on it (gidlogconstruct)
  As a developer
  I want every entity to stamp its layer and its name onto the logger it keeps
  So that a log line says who wrote it and can be filtered on a key that means
  the same thing in every service

  # One analyzer over package-level constructors, LoadModeTypesInfo (the logger
  # is recognised through internal/lgr, so both stacks count).
  # Two shapes trigger the requirement, either one is enough:
  #   1. the constructed entity has a logger field — a New<Entity> function
  #      matched against a struct of the same name;
  #   2. the constructor takes a logger parameter — whatever it stores it in
  #      (a differently spelled struct, a closure, another constructor), the
  #      logger must leave already carrying the entity name.
  # Naming happens through an enrichment call on the logger: WithField /
  # WithFields (logrus) or With / WithGroup (slog).
  # The pair is fixed:
  #   - the KEY is the layer the entity lives in — service, usecase,
  #     repository, client, event, job, schedule, and "handler" for /server/**.
  #     A free-form key ("component") is not filterable in the logs. Layers are
  #     matched with pathseg.HasLayer, anchored to the module root; in a package
  #     outside every known layer only the value is checked.
  #   - the VALUE is the entity name in lower snake_case ("device_access").
  # Only literal keys and values are judged; anything computed at runtime is
  # left alone.
  # The composition root — package main and the app layer — is exempt, as in
  # GID-104 and GID-214: it builds no entity of its own, it wires the service
  # and hands the logger to components that name themselves.
  # Generated code (ast.IsGenerated) is skipped.

  # --- Class 1: positive ---

  Scenario: positive — a constructor that never names the entity (slog)
    Given "func NewSnapshot(logger *slog.Logger) *Snapshot" storing the logger as is
    When the gidlogconstruct analyzer checks the file
    Then the diagnostic "GID-154: entity \"Snapshot\" has a logger. Fix: constructor \"NewSnapshot\" must name the entity on it — logger.WithField(<entity>, <name>) (logrus) or logger.With(\"<entity>\", <name>) (slog)" is reported

  Scenario: positive — the same on the logrus stack
    Given "func NewSnapshot(logger *logrus.Entry) *Snapshot" storing the logger as is
    When the gidlogconstruct analyzer checks the file
    Then the diagnostic "GID-154: entity \"Snapshot\" has a logger. …" is reported

  Scenario: positive — the constructor takes a logger and stores it elsewhere
    Given "func NewReporter(logger *slog.Logger) *reporter" where the struct is spelled differently
    When the gidlogconstruct analyzer checks the file
    Then the diagnostic "GID-154: entity \"Reporter\" has a logger. Fix: constructor \"NewReporter\" must name the entity on it — …" is reported
    # The parameter shape is enough: whatever it is stored in, the logger must
    # leave named.

  Scenario: positive — a key that is not the layer
    Given "func NewEmail(logger *slog.Logger) *Email" in /domain/service naming the entity under "component"
    When the gidlogconstruct analyzer checks the file
    Then the diagnostic "GID-154: the entity is named under a key other than its layer. Fix: use the layer as the key — logger.With(\"service\", \"<entity_name>\"); a free-form key (\"component\") is not filterable in the logs" is reported

  Scenario: positive — a CamelCase entity name
    Given "logger.With(\"service\", \"DeviceAccess\")" in the constructor
    When the gidlogconstruct analyzer checks the file
    Then the diagnostic "GID-154: the entity name \"DeviceAccess\" is not lower snake_case. Fix: spell it as the log fields do — \"device_access\", not \"DeviceAccess\" or \"deviceAccess\"" is reported

  Scenario: positive — a camelCase name inside a slog attribute
    Given "logger.With(slog.String(\"service\", \"deviceAccess\"))"
    When the gidlogconstruct analyzer checks the file
    Then the diagnostic "GID-154: the entity name \"deviceAccess\" is not lower snake_case. …" is reported

  Scenario: positive — a name inside a logrus.Fields literal
    Given "logger.WithFields(logrus.Fields{\"service\": \"snapshotStore\"})"
    When the gidlogconstruct analyzer checks the file
    Then the diagnostic "GID-154: the entity name \"snapshotStore\" is not lower snake_case. …" is reported

  # --- Class 2: negative ---

  Scenario: negative — the canonical naming
    Given "func NewSnapshot(logger *logrus.Entry) *Snapshot" returning "&Snapshot{logger: logger.WithField(\"service\", \"snapshot\")}" in /domain/service
    When the gidlogconstruct analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — the client layer names itself under "client"
    Given "logger.With(\"client\", \"debug_sink\")" in a constructor under /client
    When the gidlogconstruct analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — a constructor without a logger
    Given "func NewSnapshot(repo Repository) *Snapshot" where the struct has no logger field
    When the gidlogconstruct analyzer checks the file
    Then no diagnostic is reported

  # --- Class 3: boundary ---

  Scenario: boundary — the composition root
    Given "func New(logger *slog.Logger) *Application" in /internal/app/api
    When the gidlogconstruct analyzer checks the file
    Then no diagnostic is reported
    # Wiring hands the logger down to components that name themselves.

  Scenario: boundary — the server layer keys on "handler"
    Given a constructor in /server/grpc/handler naming the entity under "server"
    When the gidlogconstruct analyzer checks the file
    Then the diagnostic "GID-154: the entity is named under a key other than its layer. Fix: use the layer as the key — logger.With(\"handler\", \"<entity_name>\") …" is reported
    # The key of the /server/** layer is "handler", not the segment itself.

  Scenario: boundary — a package outside every known layer
    Given a constructor naming the entity under an arbitrary key in /pkg/tracing
    When the gidlogconstruct analyzer checks the file
    Then only the entity-name spelling is checked
    # With no layer there is no key to require.

  Scenario: boundary — a computed entity name
    Given "logger.With(\"service\", name)" where name is a variable
    When the gidlogconstruct analyzer checks the file
    Then no name-spelling diagnostic is reported
    # Only literals are judged.

  Scenario: boundary — WithGroup instead of With
    Given "logger.WithGroup(\"service\")" in the constructor
    When the gidlogconstruct analyzer checks the file
    Then it counts as naming the entity
    # Any of the enrichment calls satisfies the requirement.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — an enrichment call on a type that is not a logger
    Given "b.WithField(\"service\", \"snapshot\")" on a builder type
    When the gidlogconstruct analyzer checks the file
    Then it does not count as naming the entity
    # The receiver is resolved through internal/lgr.

  Scenario: non-applicability — package main
    Given a constructor with a logger in package main
    When the gidlogconstruct analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — generated code
    Given a file carrying the "Code generated … DO NOT EDIT." header
    When the gidlogconstruct analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-154; see also GID-104, GID-214)
#  [x] Layer chosen: go/analysis (package logconstruct)
#  [x] Message is defined ("GID-154: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest (logrus, slog and layer fixtures)
#  [x] Rule enabled in .golangci.yml
