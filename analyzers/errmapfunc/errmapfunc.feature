# language: en

Feature: GID-242 — a dedicated error-mapper function is forbidden
  As the styleguide owner
  I want error mapping (error -> error / status / code / message) to happen inline, at the place
  the error occurs
  So that no shared error-mapper translates errors from layer to layer and gets called from
  everywhere, hiding the actually bounded set of errors a real call site can produce.
  A function is a MAPPER only when it BOTH classifies its own error parameter (errors.Is /
  errors.As, or any bool-predicate over an error such as a driver's IsNoResult) AND hands back a
  translation of it — anything but a lone bool. A bool-predicate (isNotFound / isRetryable /
  isCustom) merely classifies and is legitimate, not a mapper, and neither is an observer that
  only logs and returns the error it received.

  # Layer: go/analysis (package errmapfunc, linter giderrmapfunc), LoadModeTypesInfo.
  # No exceptions for the DETECTION logic (owner's decision) — but a function's SIGNATURE can be
  # dictated by a framework rather than chosen by the service (a gdgrpcserver error converter
  # registered via WithErrorConverters is func(error) *status.Status, fixed by
  # interceptor.ErrorConverterFunc), leaving no call site to inline the mapping into. For that one
  # case, settings.exclude ("Function" | "Type.Method", the same mechanism as giderrtext/gidmapout)
  # names the function/method centrally, same as a functional-requirement exception elsewhere in
  # this repo. Config: settings.packages — the classifier package paths whose Is/As calls count;
  # default ["errors", "github.com/pkg/errors"]. settings.exclude — see the Config scenarios below.
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
  #   - F RETURNS something, and that something is not a lone bool (discriminator #1), AND
  #   - F PRODUCES a value of its own (discriminator #3).
  # All of them together → reported on F's declaration. The package is matched on the RESOLVED
  # callee (typeutil.Callee -> f.Pkg().Path()), so import aliases (pkgerrors "github.com/pkg/errors",
  # stderrors "errors") are handled automatically. A project-internal errors facade that re-exports
  # Is/As is added via settings.packages — no code change needed (and since v0.9.1 of the rule, a
  # facade Is/As is a func(error, ...) bool anyway, so shape (b) already covers it).
  #
  # Discriminator #1 (return type, owner refinement 2026-07-12, WIDENED 2026-08-06): the only
  # legitimate shape is a bool-predicate — a single bool result (named or not). Everything else a
  # classifier hands back IS the translation, whatever its Go type: error, *status.Status,
  # codes.Code, an HTTP status int, a message string. A function with no results at all is not a
  # mapper either (it translates the error into nothing).
  # Why widened (incident 2026-08-06, resource-registry internal/server/grpc/errors): the earlier
  # cut asked "does F return error?", reading the OUTPUT TYPE as the thing being ruled on. A
  # mapper split into `func Code(err error) codes.Code` + `func Converter(err error)
  # *status.Status` — both classifying err through the package's own IsNotFound/IsAlreadyExists
  # predicates — returns no error anywhere and passed clean; the package doc then cited that clean
  # run as proof the pair was legitimate.
  # Discriminator #2 (parameter vs local): the classification must branch on F's own PARAMETER,
  # not on a local variable produced inside the body (the inline handler/repository shape).
  # Discriminator #3 (produces its own value, added 2026-08-04 with shape (b), extended 2026-08-06):
  # F must replace the error — assign to its own error parameter, assign to a NAMED result, or
  # return, in some branch, an expression that is not that bare parameter. A function that
  # classifies only to decide how to log/count and always returns the same value maps nothing.
  # When F returns error, only its ERROR results are weighed (an observer with a (T, error)
  # signature returns a zero T beside the untouched err); when F returns no error, no result of
  # F can be the parameter, so every result counts.
  #
  # Why shape (b) (incident 2026-08-04, resource-registry internal/dal/repository/errors.go):
  # a storage driver publishes its error classification as bool-predicates, not as sentinels for
  # errors.Is — so `func MapError(err error) error { switch { case gdpostgres.IsUniqueViolation(err):
  # ... } }` never mentions errors.Is and passed the shape-(a)-only detector untouched.

  # --- Class 1: positive (the violation is caught) ---

  Scenario: positive — a mapper classifies the error parameter via errors.Is and returns error
    Given the top-level function "func mapErr(err error) error { switch { case errors.Is(err, ErrX): return status.Error(codes.NotFound, \"not found\"); default: return status.Error(codes.Internal, \"internal error\") } }"
    When the giderrmapfunc analyzer checks the file
    Then the diagnostic "GID-242: a dedicated error-mapper function is forbidden — it classifies its own error parameter (errors.Is/errors.As, or a bool-predicate such as IsNoResult(err)) and hands back a translation of it: an error, a *status.Status, a codes.Code, a message. Map the bounded set of errors inline, at the call site (in the repository method/handler where the error occurs): if IsNoResult(err) { err = entity.ErrNoResult }; return errors.Wrap(err, \"select x\"). Only a bool-predicate (func isNotFound(err error) bool) is a legitimate classifier, not a mapper" is reported on "mapErr"

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

  Scenario: positive — the mapper split in two: a codes.Code half and a *status.Status half, neither returning error
    Given the top-level functions "func Code(err error) codes.Code { switch { case isNotFound(err): return codes.NotFound; default: return codes.Internal } }" and "func Converter(err error) *status.Status { if isNotFound(err) { return status.New(codes.NotFound, err.Error()) }; return status.New(codes.Internal, err.Error()) }"
    When the giderrmapfunc analyzer checks the file
    Then the diagnostic "GID-242: a dedicated error-mapper function is forbidden …" is reported on both "Code" and "Converter"
    # The 2026-08-06 incident (resource-registry internal/server/grpc/errors). codes.Code is a
    # uint32 and *status.Status carries Err() rather than being an error, so an error-return-only
    # discriminator saw two clean functions — while the pair is a textbook shared mapper feeding
    # every handler through the server's error converter.

  Scenario: positive — the translation is an HTTP status code
    Given the top-level function "func mapToHTTPStatus(err error) int { switch { case errors.Is(err, ErrX): return http.StatusNotFound; default: return http.StatusInternalServerError } }"
    When the giderrmapfunc analyzer checks the file
    Then the diagnostic "GID-242: a dedicated error-mapper function is forbidden …" is reported on "mapToHTTPStatus"
    # Was a negative until 2026-08-06 (return-type discriminator, error only). The transport type
    # of the translation was never what the rule is about.

  Scenario: positive — a named result filled in by the classified branch, with a naked return
    Given the top-level function "func mapToNamedCode(err error) (code codes.Code) { code = codes.Internal; if driver.IsNoResult(err) { code = codes.NotFound }; return }"
    When the giderrmapfunc analyzer checks the file
    Then the diagnostic "GID-242: a dedicated error-mapper function is forbidden …" is reported on "mapToNamedCode"
    # There is no return expression to inspect — discriminator #3 counts the assignment to the
    # NAMED result instead.

  Scenario: positive — the translation is a message string
    Given the top-level function "func errorMessage(err error) string { if errors.Is(err, ErrX) { return \"not found\" }; return err.Error() }"
    When the giderrmapfunc analyzer checks the file
    Then the diagnostic "GID-242: a dedicated error-mapper function is forbidden …" is reported on "errorMessage"

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

  Scenario: negative — a NAMED bool result is still a predicate
    Given the function "func isKnown(err error) (ok bool) { ok = errors.Is(err, ErrX); return }"
    When the giderrmapfunc analyzer checks the file
    Then no diagnostic is reported
    # Discriminator #1 keys on the result TYPE, not on whether the result carries a name.

  Scenario: negative — a classifier with NO results translates the error into nothing
    Given the function "func logClassified(err error) { if driver.IsTemporary(err) { fmt.Println(\"temporary\"); return }; fmt.Println(\"permanent\") }"
    When the giderrmapfunc analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — an observer with a (T, error) signature returns a zero T beside the untouched error
    Given the function "func observeTuple(err error) (int, error) { if driver.IsTemporary(err) { fmt.Println(\"temporary\") }; return 0, err }"
    When the giderrmapfunc analyzer checks the file
    Then no diagnostic is reported
    # Discriminator #3 weighs only the ERROR results when F returns error — otherwise the zero T
    # would read as a translation and every observer with a value result would be reported.

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

  # --- Config: settings.exclude for a framework-mandated converter ---

  Scenario: config — a function on settings.exclude is not reported even though it has the full mapper shape
    Given settings.exclude contains "ValidationErrorConverter" and the function "func ValidationErrorConverter(err error) *status.Status { var t *ValidationErr; if !pkgerrors.As(err, &t) { return nil }; return status.New(codes.InvalidArgument, t.Field) }"
    When the giderrmapfunc analyzer checks the file
    Then no diagnostic is reported
    # ValidationErrorConverter is the canonical case settings.exclude exists for: registered in
    # gdgrpcserver.WithErrorConverters, whose interceptor.ErrorConverterFunc = func(error)
    # *status.Status fixes the signature — there is no call site to fold the mapping into
    # (resource-registry internal/server/grpc/integration, advertising-api
    # internal/server/grpc/advertising).

  Scenario: config — a function with the identical mapper shape but NOT on settings.exclude is still reported
    Given settings.exclude contains "ValidationErrorConverter" and, in the same file, the function "func Converter(err error) *status.Status { var t *ValidationErr; if !pkgerrors.As(err, &t) { return nil }; return status.New(codes.Internal, t.Field) }"
    When the giderrmapfunc analyzer checks the file
    Then the diagnostic "GID-242: a dedicated error-mapper function is forbidden …" is reported on "Converter"
    # Proves the setting names a specific function, not the shape — a second converter of the exact
    # same shape one line away is caught the same as before the setting existed. Covered together
    # with the scenario above by TestExclude.

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-242)
#  [x] Layer chosen: go/analysis (package errmapfunc: giderrmapfunc)
#  [x] Message is defined ("GID-242: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
