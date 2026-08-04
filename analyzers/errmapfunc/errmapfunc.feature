# language: en

Feature: GID-242 — a dedicated error-mapper function is forbidden
  As the styleguide owner
  I want error mapping (error -> error/status) to happen inline, at the place the error occurs
  So that no shared error-mapper translates errors from layer to layer and gets called from
  everywhere, hiding the actually bounded set of errors a real call site can produce.
  A function is a MAPPER only when it BOTH classifies its own error parameter (errors.Is /
  errors.As, or any bool-predicate over an error such as a driver's IsNoResult) AND returns an
  error of its own; a bool-predicate (isNotFound / isRetryable / isCustom) merely classifies and
  is legitimate, not a mapper, and neither is an observer that only logs and returns the error
  it received.

  # Layer: go/analysis (package errmapfunc, linter giderrmapfunc), LoadModeTypesInfo.
  # No exceptions — the rule is absolute (owner's decision). Config: settings.packages —
  # the classifier package paths whose Is/As calls count; default ["errors", "github.com/pkg/errors"].
  #
  # Detect: a top-level FuncDecl F such that ALL of
  #   - F has a NAMED parameter of type error, AND
  #   - F's body CLASSIFIES that parameter, in either shape:
  #     (a) errors.Is(<that parameter>, ...) / errors.As(<that parameter>, ...) — where errors is
  #         any of the configured classifier packages (default: stdlib "errors" +
  #         github.com/pkg/errors, which forwards Is/As to stdlib since v0.9.1; gid.team code uses
  #         pkg/errors, GID-146) — with that parameter as the first argument, anywhere;
  #     (b) a call to ANY bool-predicate over an error — a function or method whose SIGNATURE is
  #         func(error, ...) bool — with that parameter as the first argument
  #         (gdpostgres.IsUniqueViolation(err), IsNoResult(err), s.isRetryable(err)). Matched by
  #         signature, in any package (including F's own), so no whitelist is involved, AND
  #   - F's result list includes error (F returns error, or (T, error), ...), AND
  #   - F PRODUCES an error of its own (discriminator #3).
  # All of them together → reported on F's declaration. The package is matched on the RESOLVED
  # callee (typeutil.Callee -> f.Pkg().Path()), so import aliases (pkgerrors "github.com/pkg/errors",
  # stderrors "errors") are handled automatically. A project-internal errors facade that re-exports
  # Is/As is added via settings.packages — no code change needed (and since v0.9.1 of the rule, a
  # facade Is/As is a func(error, ...) bool anyway, so shape (b) already covers it).
  #
  # Discriminator #1 (return type, owner refinement 2026-07-12): only functions that RETURN
  # error are mappers. A bool-predicate over the error parameter is a legitimate classifier.
  # Discriminator #2 (parameter vs local): the classification must branch on F's own PARAMETER,
  # not on a local variable produced inside the body (the inline handler/repository shape).
  # Discriminator #3 (produces its own error, added 2026-08-04 with shape (b)): F must replace
  # the error — assign to its own error parameter, or return, in some branch, an error expression
  # that is not that bare parameter. A function that classifies only to decide how to log/count
  # and always returns the same value maps nothing.
  #
  # Why shape (b) (incident 2026-08-04, resource-registry internal/dal/repository/errors.go):
  # a storage driver publishes its error classification as bool-predicates, not as sentinels for
  # errors.Is — so `func MapError(err error) error { switch { case gdpostgres.IsUniqueViolation(err):
  # ... } }` never mentions errors.Is and passed the shape-(a)-only detector untouched.

  # --- Class 1: positive (the violation is caught) ---

  Scenario: positive — a mapper classifies the error parameter via errors.Is and returns error
    Given the top-level function "func mapErr(err error) error { switch { case errors.Is(err, ErrX): return status.Error(codes.NotFound, \"not found\"); default: return status.Error(codes.Internal, \"internal error\") } }"
    When the giderrmapfunc analyzer checks the file
    Then the diagnostic "GID-242: a dedicated error-mapper function is forbidden — it classifies its own error parameter (errors.Is/errors.As, or a bool-predicate such as IsNoResult(err)) and returns an error of its own (maps error to error/status). Map the bounded set of errors inline, at the call site (in the repository method/handler where the error occurs): if IsNoResult(err) { err = entity.ErrNoResult }; return errors.Wrap(err, \"select x\"). A bool-predicate (func isNotFound(err error) bool) is a legitimate classifier, not a mapper" is reported on "mapErr"

  Scenario: positive — a mapper classifies via errors.As (type-assert) and returns error
    Given the top-level function "func mapErrAs(err error) error { var t *CustomErr; if errors.As(err, &t) { return status.Error(codes.Internal, t.Msg) }; return err }"
    When the giderrmapfunc analyzer checks the file
    Then the diagnostic "GID-242: a dedicated error-mapper function is forbidden …" is reported on "mapErrAs"

  Scenario: positive — a mapper with a (T, error) result — the error result still makes it a mapper
    Given the top-level function "func mapErrTuple(err error) (int, error) { if errors.Is(err, ErrX) { return 0, status.Error(codes.NotFound, \"not found\") }; return 0, nil }"
    When the giderrmapfunc analyzer checks the file
    Then the diagnostic "GID-242: a dedicated error-mapper function is forbidden …" is reported on "mapErrTuple"

  Scenario: positive — a mapper via github.com/pkg/errors.Is (the gid.team default, not stdlib "errors")
    Given the top-level function "func mapPkgErr(err error) error { if pkgerrors.Is(err, ErrX) { return status.Error(codes.NotFound, \"not found\") }; return err }" with pkgerrors imported as github.com/pkg/errors
    When the giderrmapfunc analyzer checks the file
    Then the diagnostic "GID-242: a dedicated error-mapper function is forbidden …" is reported on "mapPkgErr"
    # pkgerrors.Is has package path github.com/pkg/errors, not "errors" — the whitelist must include both;
    # this is the real-code case the stdlib-only whitelist was silently missing.

  Scenario: positive — a mapper via github.com/pkg/errors.As
    Given the top-level function "func mapPkgErrAs(err error) error { var t *CustomErr; if pkgerrors.As(err, &t) { return status.Error(codes.Internal, t.Msg) }; return err }" with pkgerrors imported as github.com/pkg/errors
    When the giderrmapfunc analyzer checks the file
    Then the diagnostic "GID-242: a dedicated error-mapper function is forbidden …" is reported on "mapPkgErrAs"

  Scenario: positive — shape (b): a repository mapper built on a DRIVER's bool-predicates, no errors.Is anywhere
    Given the top-level function "func MapError(err error) error { switch { case driver.IsUniqueViolation(err): return pkgerrors.WithStack(ErrX); case driver.IsNoResult(err): return pkgerrors.WithStack(ErrX); default: return err } }"
    When the giderrmapfunc analyzer checks the file
    Then the diagnostic "GID-242: a dedicated error-mapper function is forbidden …" is reported on "MapError"
    # The incident this shape was added for: a driver classifies through predicates, so a
    # detector keyed on errors.Is/As sees nothing to match and the mapper ships.

  Scenario: positive — shape (b): the predicate lives in the mapper's own package
    Given the top-level function "func mapLocalPredicate(err error) error { if isRetryable(err) { return status.Error(codes.Unavailable, \"retry\") }; return err }"
    When the giderrmapfunc analyzer checks the file
    Then the diagnostic "GID-242: a dedicated error-mapper function is forbidden …" is reported on "mapLocalPredicate"
    # Keeping the classifier local does not make the mapper legitimate — predicates are matched
    # by signature, in any package.

  Scenario: positive — shape (b): the error is replaced by ASSIGNING to the parameter, not by returning another expression
    Given the top-level function "func mapByAssign(err error) error { if driver.IsNoResult(err) { err = ErrX }; return err }"
    When the giderrmapfunc analyzer checks the file
    Then the diagnostic "GID-242: a dedicated error-mapper function is forbidden …" is reported on "mapByAssign"
    # Discriminator #3 counts an assignment to the parameter, so the sentinel-then-return mapper
    # (whose every return is a bare `return err`) is still caught.

  Scenario: positive — shape (b): a method-shaped mapper
    Given the method "func (r *Repo) mapError(err error) error { if driver.IsUniqueViolation(err) { return ErrX }; return err }"
    When the giderrmapfunc analyzer checks the file
    Then the diagnostic "GID-242: a dedicated error-mapper function is forbidden …" is reported on "mapError"

  # --- Class 2: negative (clean code passes) ---

  Scenario: negative — an observer classifies via a driver predicate but produces no error of its own
    Given the function "func logAndReturn(err error) error { if driver.IsTemporary(err) { fmt.Println(\"temporary\") } else { fmt.Println(\"permanent\") }; return err }"
    When the giderrmapfunc analyzer checks the file
    Then no diagnostic is reported
    # Discriminator #3: it decides how to OBSERVE the error (log/metric) and hands the very same
    # value back. Nothing is mapped — without this discriminator, shape (b) would report it.

  Scenario: negative — inline mapping at the call site: the predicate classifies a LOCAL variable
    Given the method "func (r *Repository) Get() error { if err := r.conn.Select(); err != nil { if driver.IsNoResult(err) { err = ErrX }; return pkgerrors.Wrap(err, \"select\") }; return nil }"
    When the giderrmapfunc analyzer checks the file
    Then no diagnostic is reported
    # This is exactly the shape the rule demands instead of a mapper: classify where the error is
    # produced, reassign the sentinel, wrap once (GID-176/GID-244).

  Scenario: negative — a bool-predicate classifies via errors.Is but does not map (returns bool)
    Given the function "func isRetryable(err error) bool { return errors.Is(err, ErrX) }"
    When the giderrmapfunc analyzer checks the file
    Then no diagnostic is reported
    # Classifies the error but returns bool, not error — a predicate, not a mapper. Legitimate.

  Scenario: negative — a bool-predicate classifies via errors.As but does not map (returns bool)
    Given the function "func isCustom(err error) bool { var t *CustomErr; return errors.As(err, &t) }"
    When the giderrmapfunc analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — a bool-predicate via github.com/pkg/errors.Is — return-type discriminator holds for pkg/errors too
    Given the function "func isPkgRetryable(err error) bool { return pkgerrors.Is(err, ErrX) }" with pkgerrors imported as github.com/pkg/errors
    When the giderrmapfunc analyzer checks the file
    Then no diagnostic is reported
    # pkg/errors is in the classification-API whitelist, but the function returns bool, not error — a predicate.

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
    # Has an error parameter and returns error, but never classifies it — not a mapper.

  Scenario: boundary — discriminator #3 holds for shape (a) too, not only for the new predicates
    Given the function "func countAndReturn(err error) error { if errors.Is(err, ErrX) { fmt.Println(\"known\") }; return err }"
    When the giderrmapfunc analyzer checks the file
    Then no diagnostic is reported
    # errors.Is on the parameter + returns error, but the same value comes back out: an observer,
    # not a mapper. The discriminator is applied uniformly to both classification shapes.

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

  # --- Config: settings.packages adds a project errors facade ---

  Scenario: config — a mapper via a project errors facade is flagged only when its package is in settings.packages
    Given settings.packages contains "myerrors" and the mapper "func mapWithFacade(err error) error { if myerrors.Is(err, ErrX) { return myerrors.New(\"mapped\") }; return err }"
    When the giderrmapfunc analyzer checks the file
    Then the diagnostic "GID-242: a dedicated error-mapper function is forbidden …" is reported on "mapWithFacade"
    # myerrors is neither "errors" nor github.com/pkg/errors, so settings.packages is what makes its
    # Is/As count as shape (a). Since shape (b) was added, a facade whose Is is a func(error, ...) bool
    # is caught even without the setting — the whitelist stays as the explicit, signature-independent
    # way to declare a facade (an As-style classifier with an unusual signature still needs it).
    # The facade bool-predicate (func isFacadeErr(err error) bool) stays legitimate — discriminator #1.
    # Covered by TestCustomPackages.

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-242)
#  [x] Layer chosen: go/analysis (package errmapfunc: giderrmapfunc)
#  [x] Message is defined ("GID-242: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
