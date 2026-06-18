# language: en

Feature: GID-176 / GID-177 — error handling by layer (errwrap)
  As a developer
  I want errors from the application boundary to collect a stack and context (Wrap),
  non-static errors inside the application to be enriched without a second stack (WithMessage),
  and static errors to always carry a stack when returned (WithStack)
  So that the error trace is complete and the stack is collected exactly once — at the boundary

  # Two analyzers in one errwrap package (like errplace), LoadModeTypesInfo:
  #   - WrapAnalyzer  → linter giderrwrap  (GID-176)
  #   - StaticAnalyzer → linter gidstaticerr (GID-177)
  # pkg/errors is recognized by the import path github.com/pkg/errors (a stub in testdata).
  # Generated code (ast.IsGenerated) is skipped.
  # The layer is matched by path segments via internal/pathseg.

  # ============================================================
  # GID-176 (giderrwrap)
  # ============================================================

  # === Part 1: the application boundary (/client/** and /dal/repository) ===
  # The boundary is an interface-method call (an injected external dependency,
  # e.g. r.conn.call()). A function returning error must not pass through such a
  # non-static error without errors.Wrap. A call to a local package function
  # (a pure SQL builder build.Select(...)) or a concrete-type method is NOT the
  # boundary and may be enriched with WithMessage.

  # --- Class 1: positive (the violation is caught) ---

  Scenario: positive — pass-through of an error from an interface-method call at the boundary
    Given a package in "/dal/repository" with a function returning error where "err := r.conn.call(); return err"
    When the giderrwrap analyzer checks the file
    Then the diagnostic "GID-176: an error from the app boundary must be wrapped with errors.Wrap. Fix: collect stack and context; to map a sentinel, reassign then wrap once: if IsNoResult(err) { err = ErrNoResult }; return errors.Wrap(err, ...)" is reported on "err"

  Scenario: positive — pass-through in a multi-value return (return n, err)
    Given a package in "/dal/repository" with the function "n, err := r.conn.callRow(); return n, err"
    When the giderrwrap analyzer checks the file
    Then the diagnostic "GID-176: an error from the app boundary must be wrapped with errors.Wrap …" is reported on "err"

  Scenario: positive — WithStack/WithMessage add no context
    Given a package in "/dal/repository" with "err := r.conn.call(); return errors.WithStack(err)" (and likewise WithMessage)
    When the giderrwrap analyzer checks the file
    Then the diagnostic "GID-176: an error from the app boundary must be wrapped with errors.Wrap (WithStack adds no context). Fix: …" is reported

  Scenario: positive — the /client boundary (interface-method call)
    Given a package in "/client" with the function "err := c.transport.do(); return err"
    When the giderrwrap analyzer checks the file
    Then the diagnostic "GID-176: an error from the app boundary must be wrapped with errors.Wrap …" is reported on "err"

  # --- Class 2: negative (clean code passes) ---

  Scenario: negative — the error from an interface-method call is wrapped with Wrap
    Given a package in "/dal/repository" with "err := r.conn.call(); return errors.Wrap(err, \"select\")"
    When the giderrwrap analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — map a boundary error to a sentinel, then a single Wrap
    Given a package in "/dal/repository" with "err := r.conn.call(); if isNoResult(err) { err = entity.ErrNotFound }; return errors.Wrap(err, \"select\")"
    When the giderrwrap analyzer checks the file
    Then no diagnostic is reported
    # Reassigning err to a sentinel before one errors.Wrap avoids wrapping twice.

  Scenario: negative — an error from a local package function (a pure builder) is not the boundary
    Given a package in "/dal/repository" with "_, err := buildQuery(); return errors.WithMessage(err, \"build query\")"
    When the giderrwrap analyzer checks the file
    Then no diagnostic is reported
    # buildQuery is a package function, not an interface-method call.

  Scenario: negative — a method on a concrete type is not an interface-method call
    Given a package in "/dal/repository" with "err := r.concreteHelper(); return errors.WithMessage(err, \"helper\")"
    When the giderrwrap analyzer checks the file
    Then no diagnostic is reported

  # --- Class 3: boundary (looks like a violation but is allowed) ---

  Scenario: boundary — returning a static error at the boundary (the domain of GID-177)
    Given a package in "/dal/repository" with "if err != nil { return entity.ErrNotFound }"
    When the giderrwrap analyzer checks the file
    Then no diagnostic is reported
    # Returning a package-level error var is not pass-through; it is an error exchange, checked by GID-177.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — a function without error in the results
    Given a package in "/dal/repository" with a function returning int (no error)
    When the giderrwrap analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — not a boundary layer (pkg/util)
    Given a package in "/pkg/util" with "err := w.call(); return err"
    When the giderrwrap analyzer checks the file
    Then no diagnostic is reported
    # Part 1 applies only in /client/** and /dal/repository.

  # === Part 2: inside the application (/domain/**) ===
  # errors.Wrap of an incoming non-static error is forbidden — the stack was already collected at the boundary.

  # --- Class 1: positive ---

  Scenario: positive — Wrap of a non-static error in /domain/service
    Given a package in "/domain/service" with "err := s.call(); return errors.Wrap(err, \"ctx\")"
    When the giderrwrap analyzer checks the file
    Then the diagnostic "GID-176: the stack is already collected at the boundary. Fix: use errors.WithMessage instead of errors.Wrap for an incoming error" is reported

  Scenario: positive — Wrap of an error passed as a function parameter
    Given a package in "/domain/service" with "func (s *Service) f(err error) error { return errors.Wrap(err, \"ctx\") }"
    When the giderrwrap analyzer checks the file
    Then the diagnostic "GID-176: the stack is already collected at the boundary …" is reported

  # --- Class 2: negative ---

  Scenario: negative — WithMessage for an incoming error
    Given a package in "/domain/service" with "err := s.call(); return errors.WithMessage(err, \"ctx\")"
    When the giderrwrap analyzer checks the file
    Then no diagnostic is reported

  # --- Class 3: boundary ---

  Scenario: boundary — Wrap of a static error from model — allowed
    Given a package in "/domain/service" with "return errors.Wrap(model.ErrSnapshotNotFound, \"ctx\")"
    When the giderrwrap analyzer checks the file
    Then no diagnostic is reported
    # Wrap of a static (package-level) error is needed — it collects the stack first.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — WithStack of an incoming error in domain (not Wrap)
    Given a package in "/domain/service" with "err := s.call(); return errors.WithStack(err)"
    When the giderrwrap analyzer checks the file
    Then no diagnostic is reported
    # Part 2 catches only errors.Wrap; WithStack is outside its scope.

  # ============================================================
  # GID-177 (gidstaticerr)
  # ============================================================
  # Everywhere (except testdata/generated): a return of a static error without a wrapper.

  # --- Class 1: positive ---

  Scenario: positive — return of a package-level error var without a wrapper
    Given a function with "return model.ErrSnapshotNotFound"
    When the gidstaticerr analyzer checks the file
    Then the diagnostic "GID-177: a static error is returned without a stack. Fix: wrap with errors.WithStack (or errors.Wrap if you need context)" is reported

  Scenario: positive — return of the address of a named error type (&BigError{})
    Given a function with "return &model.BigError{Code: 1}"
    When the gidstaticerr analyzer checks the file
    Then the diagnostic "GID-177: a static error is returned without a stack …" is reported

  Scenario: positive — return of a composite literal of an error type (BigError{})
    Given a function with "return model.BigError{Code: 2}"
    When the gidstaticerr analyzer checks the file
    Then the diagnostic "GID-177: a static error is returned without a stack …" is reported

  # --- Class 2: negative ---

  Scenario: negative — a static error wrapped with WithStack or Wrap
    Given a function with "return errors.WithStack(model.ErrSnapshotNotFound)" (and likewise errors.Wrap)
    When the gidstaticerr analyzer checks the file
    Then no diagnostic is reported

  # --- Class 3: boundary ---

  Scenario: boundary — returning an incoming non-static error
    Given the function "func (s *Service) f(err error) error { return err }"
    When the gidstaticerr analyzer checks the file
    Then no diagnostic is reported
    # A non-static error (parameter/local) is not the domain of GID-177 (it is GID-176).

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — an exempted constructor collects the stack (settings.exclude)
    Given settings.exclude contains "gderror.NewUnhandledValueError" and "return gderror.NewUnhandledValueError(x)"
    When the gidstaticerr analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — an error var declaration (not a return)
    Given "var ErrSnapshotNotFound = errors.New(\"…\")" in error.go
    When the gidstaticerr analyzer checks the file
    Then no diagnostic is reported
    # GID-177 checks only return expressions, not declarations.

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-176/177)
#  [x] Layer chosen: go/analysis (package errwrap: giderrwrap + gidstaticerr)
#  [x] Messages are defined ("GID-176: …" / "GID-177: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability — for each analyzer
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml
