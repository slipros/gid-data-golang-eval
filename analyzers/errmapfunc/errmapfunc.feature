# language: en

Feature: GID-242 — a dedicated function that classifies its own error parameter via errors.Is is forbidden
  As the styleguide owner
  I want error classification/mapping to happen inline, at the place the error occurs
  So that no universal "handle everything the same way, from everywhere" helper accumulates
  unrelated errors and hides the actually bounded set of errors a real call site can produce —
  regardless of WHAT the function maps to (gRPC status, HTTP status, a log level, another
  error, a business code, ...): the forbidden shape is the function, not its target

  # Layer: go/analysis (package errmapfunc, linter giderrmapfunc), LoadModeTypesInfo.
  # No settings, no exceptions — the rule is absolute (owner's decision).
  #
  # Detect: a top-level FuncDecl F such that
  #   - F has a NAMED parameter of type error, AND
  #   - F's body calls errors.Is(<that parameter>, ...) (stdlib errors.Is) anywhere.
  # Both together → reported on F's declaration. What F does with the classification is
  # irrelevant — gRPC status is only ONE possible instance of the pattern, not a requirement.

  # --- Class 1: positive (the violation is caught) ---

  Scenario: positive — a dedicated function maps its error parameter to a plain HTTP status (no gRPC at all)
    Given the top-level function "func mapToHTTPStatus(err error) int { switch { case errors.Is(err, ErrX): return http.StatusNotFound; default: return http.StatusInternalServerError } }"
    When the giderrmapfunc analyzer checks the file
    Then the diagnostic "GID-242: a dedicated function that classifies its own error parameter via errors.Is is forbidden — handle the bounded set of errors inline, at the call site (in the handler/interceptor where the error occurs), whatever the target (status code, log level, another error, ...). Fix: remove the function, inline the switch errors.Is(...) into the caller" is reported on "mapToHTTPStatus"
    # No gRPC package is even imported here — the rule is not about gRPC.

  Scenario: positive — gRPC status is just another instance of the same forbidden shape
    Given the top-level function "func mapErr(err error) error { switch { case errors.Is(err, ErrX): return status.Error(codes.NotFound, \"not found\"); default: return status.Error(codes.Internal, \"internal error\") } }"
    When the giderrmapfunc analyzer checks the file
    Then the diagnostic "GID-242: a dedicated function that classifies its own error parameter via errors.Is is forbidden …" is reported on "mapErr"

  # --- Class 2: negative (clean code passes) ---

  Scenario: negative — inline handling in a handler switches on a LOCAL variable, not a parameter
    Given the method "func (h *Handler) Handle() (int, error) { res, err := h.u.Do(); if err != nil { switch { case errors.Is(err, ErrX): return 0, status.Error(codes.NotFound, \"not found\") } }; return res, nil }"
    When the giderrmapfunc analyzer checks the file
    Then no diagnostic is reported
    # err here is a local variable — the result of an inner call inside the handler's own body — not
    # a parameter of Handle. The discriminator is whether errors.Is inspects F's own parameter or not.

  Scenario: negative — not a mapper at all (no errors.Is anywhere)
    Given the function "func wrap(err error) error { return fmt.Errorf(\"wrap: %w\", err) }"
    When the giderrmapfunc analyzer checks the file
    Then no diagnostic is reported

  # --- Class 3: boundary ---

  Scenario: boundary — a function with an error parameter that never calls errors.Is on it
    Given the function "func passthrough(err error) error { return err }"
    When the giderrmapfunc analyzer checks the file
    Then no diagnostic is reported
    # Has an error parameter, but never classifies it via errors.Is — the required condition never holds.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — no error parameter at all (a request validator)
    Given the function "func validate(req Req) error { if req.Name == \"\" { return status.Error(codes.InvalidArgument, \"name is required\") }; return nil }"
    When the giderrmapfunc analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — an unnamed error parameter (cannot be referenced by errors.Is)
    Given the function "func discard(error) error { return status.Error(codes.Internal, \"x\") }"
    When the giderrmapfunc analyzer checks the file
    Then no diagnostic is reported
    # An unnamed parameter has no identifier errors.Is could ever branch on inside the body.

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-242)
#  [x] Layer chosen: go/analysis (package errmapfunc: giderrmapfunc)
#  [x] Message is defined ("GID-242: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml
