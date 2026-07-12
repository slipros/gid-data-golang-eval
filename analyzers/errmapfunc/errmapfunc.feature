# language: en

Feature: GID-242 — a dedicated error-mapper function is forbidden
  As the styleguide owner
  I want error mapping (error -> error/status) to happen inline, at the place the error occurs
  So that no shared error-mapper translates errors from layer to layer and gets called from
  everywhere, hiding the actually bounded set of errors a real call site can produce.
  A function is a MAPPER only when it BOTH classifies its own error parameter (errors.Is /
  errors.As) AND returns error; a bool-predicate (isNotFound / isRetryable / isCustom) merely
  classifies and is legitimate, not a mapper.

  # Layer: go/analysis (package errmapfunc, linter giderrmapfunc), LoadModeTypesInfo.
  # No settings, no exceptions — the rule is absolute (owner's decision).
  #
  # Detect: a top-level FuncDecl F such that ALL of
  #   - F has a NAMED parameter of type error, AND
  #   - F's body calls errors.Is(<that parameter>, ...) OR errors.As(<that parameter>, ...)
  #     (stdlib errors package, that parameter as the first argument) anywhere, AND
  #   - F's result list includes error (F returns error, or (T, error), ...).
  # All three together → reported on F's declaration.
  #
  # Discriminator #1 (return type, owner refinement 2026-07-12): only functions that RETURN
  # error are mappers. A bool-predicate over the error parameter is a legitimate classifier.
  # Discriminator #2 (parameter vs local): errors.Is/As must branch on F's own PARAMETER, not
  # on a local variable produced inside the body (the inline handler shape).

  # --- Class 1: positive (the violation is caught) ---

  Scenario: positive — a mapper classifies the error parameter via errors.Is and returns error
    Given the top-level function "func mapErr(err error) error { switch { case errors.Is(err, ErrX): return status.Error(codes.NotFound, \"not found\"); default: return status.Error(codes.Internal, \"internal error\") } }"
    When the giderrmapfunc analyzer checks the file
    Then the diagnostic "GID-242: a dedicated error-mapper function is forbidden — it classifies its own error parameter via errors.Is/errors.As and returns error (maps error to error/status). Map the bounded set of errors inline, at the call site (in the handler/interceptor where the error occurs). A bool-predicate (func isNotFound(err error) bool) is a legitimate classifier, not a mapper. Fix: remove the function, inline the switch errors.Is(...) into the caller" is reported on "mapErr"

  Scenario: positive — a mapper classifies via errors.As (type-assert) and returns error
    Given the top-level function "func mapErrAs(err error) error { var t *CustomErr; if errors.As(err, &t) { return status.Error(codes.Internal, t.Msg) }; return err }"
    When the giderrmapfunc analyzer checks the file
    Then the diagnostic "GID-242: a dedicated error-mapper function is forbidden …" is reported on "mapErrAs"

  Scenario: positive — a mapper with a (T, error) result — the error result still makes it a mapper
    Given the top-level function "func mapErrTuple(err error) (int, error) { if errors.Is(err, ErrX) { return 0, status.Error(codes.NotFound, \"not found\") }; return 0, nil }"
    When the giderrmapfunc analyzer checks the file
    Then the diagnostic "GID-242: a dedicated error-mapper function is forbidden …" is reported on "mapErrTuple"

  # --- Class 2: negative (clean code passes) ---

  Scenario: negative — a bool-predicate classifies via errors.Is but does not map (returns bool)
    Given the function "func isRetryable(err error) bool { return errors.Is(err, ErrX) }"
    When the giderrmapfunc analyzer checks the file
    Then no diagnostic is reported
    # Classifies the error but returns bool, not error — a predicate, not a mapper. Legitimate.

  Scenario: negative — a bool-predicate classifies via errors.As but does not map (returns bool)
    Given the function "func isCustom(err error) bool { var t *CustomErr; return errors.As(err, &t) }"
    When the giderrmapfunc analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — classifies via errors.Is but returns a plain int (HTTP status), not error
    Given the function "func mapToHTTPStatus(err error) int { switch { case errors.Is(err, ErrX): return http.StatusNotFound; default: return http.StatusInternalServerError } }"
    When the giderrmapfunc analyzer checks the file
    Then no diagnostic is reported
    # By the return-type discriminator, a function that does not return error is not a mapper.

  Scenario: negative — inline handling in a handler switches on a LOCAL variable, not a parameter
    Given the method "func (h *Handler) Handle() (int, error) { res, err := h.u.Do(); if err != nil { switch { case errors.Is(err, ErrX): return 0, status.Error(codes.NotFound, \"not found\") } }; return res, nil }"
    When the giderrmapfunc analyzer checks the file
    Then no diagnostic is reported
    # err here is a local variable — the result of an inner call inside the handler's own body — not
    # a parameter of Handle. The discriminator is whether errors.Is/As inspects F's own parameter.

  Scenario: negative — returns error but never calls errors.Is/errors.As (a plain wrapper)
    Given the function "func wrap(err error) error { return fmt.Errorf(\"wrap: %w\", err) }"
    When the giderrmapfunc analyzer checks the file
    Then no diagnostic is reported

  # --- Class 3: boundary ---

  Scenario: boundary — a function with an error parameter that returns error but never classifies it
    Given the function "func passthrough(err error) error { return err }"
    When the giderrmapfunc analyzer checks the file
    Then no diagnostic is reported
    # Has an error parameter and returns error, but never calls errors.Is/As on it — not a mapper.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — no error parameter at all (a request validator)
    Given the function "func validate(req Req) error { if req.Name == \"\" { return status.Error(codes.InvalidArgument, \"name is required\") }; return nil }"
    When the giderrmapfunc analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — an unnamed error parameter (cannot be referenced by errors.Is/As)
    Given the function "func discard(error) error { return status.Error(codes.Internal, \"x\") }"
    When the giderrmapfunc analyzer checks the file
    Then no diagnostic is reported
    # An unnamed parameter has no identifier errors.Is/As could ever branch on inside the body.

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-242)
#  [x] Layer chosen: go/analysis (package errmapfunc: giderrmapfunc)
#  [x] Message is defined ("GID-242: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
