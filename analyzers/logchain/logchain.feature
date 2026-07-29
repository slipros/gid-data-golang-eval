# language: en

Feature: GID-156 — a logger chain puts one call per line (gidlogchain)
  As a developer
  I want every call of a logger chain on its own line, the first one included
  So that a log statement is read as a list of what it carries, and a diff
  touching one link stays one line

  # One analyzer over call expressions, LoadModeTypesInfo: the chain is
  # recognised through internal/lgr, so only real logger calls count.
  # Entry point: a terminal call (lgr.IsTerminal — Info/Error/… for logrus,
  # InfoContext/… for slog); from it lgr.Chain walks back to the base
  # expression and collects the selectors.
  # Threshold: a chain of at least 2 calls. A single inline call is allowed.
  # The check is positional: walking the chain in source order, each call must
  # start on a line strictly below the previous one — which is why the first
  # call must leave the base expression's line too.
  # Stack-agnostic: logrus With*/Info and slog With/InfoContext alike.
  # Generated code (ast.IsGenerated) is skipped.

  # --- Class 1: positive ---

  Scenario: positive — the whole logrus chain inline
    Given "s.logger.WithContext(ctx).WithError(err).Error(\"failed\")" on one line
    When the gidlogchain analyzer checks the file
    Then the diagnostic "GID-156: a logger chain must put one call per line, including the first. Fix: break each call onto a new line" is reported

  Scenario: positive — the first call sits on the logger's line
    Given "s.logger.WithContext(ctx)." followed by the rest of the chain on its own lines
    When the gidlogchain analyzer checks the file
    Then the diagnostic "GID-156: a logger chain must put one call per line, including the first. …" is reported
    # "including the first": the chain starts below the logger, not next to it.

  Scenario: positive — the slog stack
    Given "s.logger.With(slog.String(\"step\", \"start\")).InfoContext(ctx, \"start\")" on one line
    When the gidlogchain analyzer checks the file
    Then the diagnostic "GID-156: …" is reported

  # --- Class 2: negative ---

  Scenario: negative — one call per line
    Given the chain "s.logger.\n\tWithContext(ctx).\n\tWithError(err).\n\tWithField(\"some\", field).\n\tInfo(\"some text\")"
    When the gidlogchain analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — a single inline call
    Given "s.logger.Info(\"started\")"
    When the gidlogchain analyzer checks the file
    Then no diagnostic is reported
    # One call is not a chain.

  # --- Class 3: boundary ---

  Scenario: boundary — exactly two calls, both inline
    Given "s.logger.WithError(err).Error(\"failed\")" on one line
    When the gidlogchain analyzer checks the file
    Then the diagnostic "GID-156: …" is reported
    # Two calls is where the chain begins.

  Scenario: boundary — the chain is broken in the middle only
    Given "s.logger.\n\tWithContext(ctx).WithError(err).\n\tError(\"failed\")"
    When the gidlogchain analyzer checks the file
    Then the diagnostic "GID-156: …" is reported
    # Every call is checked against the previous line, not just the first one.

  Scenario: boundary — one diagnostic per chain
    Given a chain with several calls sharing lines
    When the gidlogchain analyzer checks the file
    Then a single diagnostic is reported on the first call that broke the order
    # The chain is reformatted as a whole; repeating the message per link adds
    # nothing.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — a chain on a type that is not a logger
    Given "b.WithHeader(h).WithBody(x).Do(ctx)" on a builder type
    When the gidlogchain analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — generated code
    Given a file carrying the "Code generated … DO NOT EDIT." header
    When the gidlogchain analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-156)
#  [x] Layer chosen: go/analysis (package logchain)
#  [x] Message is defined ("GID-156: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest (logrus and slog fixtures)
#  [x] Rule enabled in .golangci.yml
