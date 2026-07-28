# language: en

Feature: GID-253 — a logger chain sets its fields in one call (gidlogfields)
  As a developer
  I want the fields of one log record collected into a single WithFields call
  So that the payload is read as one block and no entry is allocated per field

  # One analyzer over call expressions, LoadModeTypesInfo (the receiver type is
  # resolved through internal/lgr, so only real logrus chains count).
  # Scope: a chain of logrus method calls, with or without a terminal call.
  # Field calls are WithField and WithFields; WithContext/WithError carry the
  # context and the error (GID-155), not fields, and never count.
  # Threshold: 2 field calls in one chain — two pairs already fit one
  # logrus.Fields literal.
  # slog is out of scope: its With takes variadic pairs, so the batch form is
  # the only form. Generated code (ast.IsGenerated) is skipped.

  # --- Class 1: positive ---

  Scenario: positive — four WithField calls in one chain
    Given a chain "logger.WithContext(ctx).WithError(err).WithField(\"offset\", offset).WithField(\"target_topic\", t).WithField(\"fallback_level\", l).WithField(\"fallback_retry\", r).Error(msg)"
    When the gidlogfields analyzer checks the file
    Then the diagnostic "GID-253: a logger chain sets its fields in 4 separate calls — they belong in one. Fix: replace them with a single WithFields(logrus.Fields{\"offset\": offset, \"fallback_level\": level})" is reported on the first WithField

  Scenario: positive — the logger is an interface (logrus.FieldLogger)
    Given a chain "p.logger.WithError(err).WithField(\"topic\", topic).WithField(\"partition\", partition).Error(msg)"
    When the gidlogfields analyzer checks the file
    Then the diagnostic "GID-253: …" is reported

  # --- Class 2: negative ---

  Scenario: negative — the fields are passed in one WithFields
    Given a chain "logger.WithContext(ctx).WithError(err).WithFields(logrus.Fields{\"offset\": offset, \"fallback_level\": level}).Error(msg)"
    When the gidlogfields analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — a single pair through WithField
    Given a chain "logger.WithContext(ctx).WithField(\"offset\", offset).Info(msg)"
    When the gidlogfields analyzer checks the file
    Then no diagnostic is reported
    # One pair is exactly what WithField is for.

  # --- Class 3: boundary ---

  Scenario: boundary — exactly two field calls
    Given a chain "logger.WithField(\"a\", a).WithField(\"b\", b).Info(msg)"
    When the gidlogfields analyzer checks the file
    Then the diagnostic "GID-253: a logger chain sets its fields in 2 separate calls — …" is reported

  Scenario: boundary — WithFields followed by WithField
    Given a chain "logger.WithFields(logrus.Fields{\"offset\": offset}).WithField(\"fallback_level\", level).Info(msg)"
    When the gidlogfields analyzer checks the file
    Then the diagnostic "GID-253: …" is reported
    # Either mix of the two methods counts — the pairs still fit one literal.

  Scenario: boundary — a chain without a terminal call
    Given "entry := logger.WithField(\"a\", a).WithField(\"b\", b)" with no Info/Error at the end
    When the gidlogfields analyzer checks the file
    Then the diagnostic "GID-253: …" is reported

  Scenario: boundary — fields attached through separate statements
    Given "entry := logger.WithField(\"a\", a)" followed by "entry = entry.WithField(\"b\", b)"
    When the gidlogfields analyzer checks the file
    Then no diagnostic is reported
    # Each statement holds one chain with one field call; the rule does not
    # chase an entry across statements.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — WithField on a type that is not a logrus logger
    Given a chain "f.WithField(\"a\", a).WithField(\"b\", b).Info(msg)" on a local *fakeLogger type
    When the gidlogfields analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — WithContext and WithError are not field calls
    Given a chain "logger.WithContext(ctx).WithError(err).Error(msg)"
    When the gidlogfields analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — the slog stack
    Given a chain "logger.With(\"offset\", offset).With(\"fallback_level\", level).InfoContext(ctx, msg)"
    When the gidlogfields analyzer checks the file
    Then no diagnostic is reported
    # slog.With takes variadic pairs — there is no single-field method to repeat.

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-253)
#  [x] Layer chosen: go/analysis (package logfields)
#  [x] Message is defined ("GID-253: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
