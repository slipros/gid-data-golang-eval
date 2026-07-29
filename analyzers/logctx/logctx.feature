# language: en

Feature: GID-155 — a log call carries the context and, at Error level, the error (gidlogctx)
  As a developer
  I want every log record to carry the request context and every error log to
  carry the error
  So that a record can be traced back to its request and a failure is readable
  without digging for the original error

  # One analyzer over call expressions, LoadModeTypesInfo; the logger and the
  # shape of the chain come from internal/lgr, so both stacks are handled.
  # Two independent checks on a terminal log call:
  #   1. the enclosing function has a context.Context parameter -> the call
  #      must carry it: WithContext(ctx) in the chain (logrus) or the *Context
  #      variant of the terminal method, InfoContext(ctx, …) (slog, which has
  #      no WithContext);
  #   2. the terminal method starts with "Error" -> the call must carry the
  #      error: WithError(err) (logrus) or an argument of error type anywhere
  #      in the call — slog.Any("error", err).
  # Scope of "has ctx": the nearest enclosing function. A function literal is
  # walked with its own parameter list.
  # Only the Error level is checked for the error: "WithError whenever an error
  # is in scope" needs flow analysis and is not deterministic.
  # The diagnostic is placed on the first selector of the chain, so a
  # multi-line chain reports on its first line.
  # Generated code (ast.IsGenerated) is skipped.

  # --- Class 1: positive ---

  Scenario: positive — a log without the context in a function taking ctx
    Given "func (s *Snapshot) Do(ctx context.Context)" containing "s.logger.Info(\"start\")"
    When the gidlogctx analyzer checks the file
    Then the diagnostic "GID-155: a log call in a function with ctx must carry the context. Fix: add WithContext(ctx) (logrus) or call the *Context variant — InfoContext(ctx, …) (slog)" is reported

  Scenario: positive — an Error log without the error
    Given the chain "s.logger.\n\tWithContext(ctx).\n\tError(\"failed\")"
    When the gidlogctx analyzer checks the file
    Then the diagnostic "GID-155: an Error-level log must carry the error. Fix: add WithError(err) (logrus) or pass the error as an attribute — ErrorContext(ctx, msg, slog.Any(\"error\", err)) (slog)" is reported

  Scenario: positive — both at once
    Given "s.logger.Error(\"failed\")" in a function taking ctx
    When the gidlogctx analyzer checks the file
    Then both the context diagnostic and the error diagnostic are reported on the call

  Scenario: positive — slog ErrorContext without the error
    Given "s.logger.ErrorContext(ctx, \"failed\")"
    When the gidlogctx analyzer checks the file
    Then the diagnostic "GID-155: an Error-level log must carry the error. …" is reported
    # The context is carried by the *Context variant; the error is still missing.

  # --- Class 2: negative ---

  Scenario: negative — the full logrus shape
    Given the chain "s.logger.\n\tWithContext(ctx).\n\tWithError(err).\n\tError(\"failed\")"
    When the gidlogctx analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — the full slog shape
    Given "s.logger.ErrorContext(ctx, \"failed\", slog.Any(\"error\", err))"
    When the gidlogctx analyzer checks the file
    Then no diagnostic is reported

  # --- Class 3: boundary ---

  Scenario: boundary — a function without ctx
    Given "func (s *Snapshot) Name() string" containing "s.logger.Info(\"called\")"
    When the gidlogctx analyzer checks the file
    Then no diagnostic is reported
    # There is no context to carry.

  Scenario: boundary — a function literal without ctx inside a function with ctx
    Given "go func() { s.logger.Info(\"tick\") }()" inside a method taking ctx
    When the gidlogctx analyzer checks the file
    Then no context diagnostic is reported
    # The presence of ctx is that of the nearest enclosing function, which is
    # the literal.

  Scenario: boundary — an Info log without an error
    Given "s.logger.WithContext(ctx).Info(\"started\")" in a function with an error in scope
    When the gidlogctx analyzer checks the file
    Then no diagnostic is reported
    # Only Error-level calls are required to carry the error.

  Scenario: boundary — Errorf and other Error* levels
    Given "s.logger.WithContext(ctx).Errorf(\"failed: %s\", id)"
    When the gidlogctx analyzer checks the file
    Then the diagnostic "GID-155: an Error-level log must carry the error. …" is reported
    # Every terminal method starting with Error counts.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — a call on a type that is not a logger
    Given "tracer.Info(\"start\")" on a local type in a function with ctx
    When the gidlogctx analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — generated code
    Given a file carrying the "Code generated … DO NOT EDIT." header
    When the gidlogctx analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-155)
#  [x] Layer chosen: go/analysis (package logctx)
#  [x] Message is defined ("GID-155: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest (logrus and slog fixtures)
#  [x] Rule enabled in .golangci.yml
