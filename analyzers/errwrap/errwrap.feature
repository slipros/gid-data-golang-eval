# language: en

Feature: GID-176 / GID-177 / GID-237 — error handling by layer (errwrap)
  As a developer
  I want every external call's error to collect a stack and context (Wrap), wherever it is made,
  a same-module non-static error inside the application to be enriched without a second stack (WithMessage),
  static errors to always carry a stack when returned (WithStack),
  and a service to never add a message to an incoming error (that belongs to usecase)
  So that the error trace is complete and the stack is collected exactly once — at the external call

  # Three analyzers in one errwrap package (like errplace), LoadModeTypesInfo:
  #   - WrapAnalyzer          → linter giderrwrap   (GID-176)
  #   - StaticAnalyzer        → linter gidstaticerr (GID-177)
  #   - ServiceMessageAnalyzer → linter gidwithmessage (GID-237)
  # pkg/errors is recognized by the import path github.com/pkg/errors (a stub in testdata).
  # Generated code (ast.IsGenerated) is skipped.
  # The layer is matched by path segments via internal/pathseg.

  # ============================================================
  # GID-176 (giderrwrap) — v2 (2026-07-04): "every external call's error is
  # always wrapped with errors.Wrap, wherever it is made." Two call shapes
  # count as an external call:
  #   (a) a DIRECT call to a function/method whose declaring package lies
  #       outside the current module (stdlib counts as external too) — a
  #       boundary in ANY layer, including /domain/**;
  #   (b) an INTERFACE-METHOD call on an injected dependency (e.g.
  #       c.conn.Select(...)) — a boundary only in the scoped layers:
  #       /client/**, /dal/repository, /event/** (a Kafka producer/consumer
  #       talks to an external system; canon: example_event.md uses Wrap).
  # ============================================================

  # === Part 1: external calls (mechanisms a + b) ===
  # A function returning error must not pass through a tracked external-call
  # error without errors.Wrap. A call to a local same-module package function
  # (a pure SQL builder build.Select(...)) or a concrete-type method — neither
  # (a) nor (b) — is NOT a boundary call and may be enriched with WithMessage
  # or passed through as is.

  # --- Class 1: positive (the violation is caught) ---

  Scenario: positive — pass-through of an error from an interface-method call at the boundary (mechanism b)
    Given a package in "/dal/repository" with a function returning error where "err := r.conn.call(); return err"
    When the giderrwrap analyzer checks the file
    Then the diagnostic "GID-176: an error from an external call must be wrapped with errors.Wrap. Fix: collect stack and context; to map a sentinel, reassign then wrap once: if IsNoResult(err) { err = ErrNoResult }; return errors.Wrap(err, ...)" is reported on "err"

  Scenario: positive — pass-through in a multi-value return (return n, err)
    Given a package in "/dal/repository" with the function "n, err := r.conn.callRow(); return n, err"
    When the giderrwrap analyzer checks the file
    Then the diagnostic "GID-176: an error from an external call must be wrapped with errors.Wrap …" is reported on "err"

  Scenario: positive — WithStack/WithMessage add no context
    Given a package in "/dal/repository" with "err := r.conn.call(); return errors.WithStack(err)" (and likewise WithMessage)
    When the giderrwrap analyzer checks the file
    Then the diagnostic "GID-176: an error from an external call must be wrapped with errors.Wrap (WithStack adds no context). Fix: …" is reported

  Scenario: positive — the /client boundary (interface-method call, mechanism b)
    Given a package in "/client" with the function "err := c.transport.do(); return err"
    When the giderrwrap analyzer checks the file
    Then the diagnostic "GID-176: an error from an external call must be wrapped with errors.Wrap …" is reported on "err"

  Scenario: positive (v2) — the /event/** boundary (interface-method call, mechanism b)
    Given a package in "/event/producer" with the function "err := p.client.Send(topic, msg); return err" (KafkaClient is the injected dependency)
    When the giderrwrap analyzer checks the file
    Then the diagnostic "GID-176: an error from an external call must be wrapped with errors.Wrap …" is reported on "err"
    # /event/** joins /client/** and /dal/repository as a scoped boundary layer — a Kafka producer/consumer talks to an external system.

  Scenario: positive (v2) — a direct external call (mechanism a) in ANY layer, e.g. /pkg/util
    Given a package in "/pkg/util" (not a scoped boundary layer) with "_, err := strconv.Atoi(x); return err"
    When the giderrwrap analyzer checks the file
    Then the diagnostic "GID-176: an error from an external call must be wrapped with errors.Wrap …" is reported on "err"
    # strconv is outside the current module (stdlib counts as external); mechanism (a) is checked everywhere, not just in the scoped layers.

  Scenario: positive (v2) — a direct external call (mechanism a) inside /event/**, on top of the interface-call boundary (mechanism b)
    Given a package in "/event/consumer" with "err := json.Unmarshal(data, &e); return err"
    When the giderrwrap analyzer checks the file
    Then the diagnostic "GID-176: an error from an external call must be wrapped with errors.Wrap …" is reported on "err"

  # --- Class 2: negative (clean code passes) ---

  Scenario: negative — the error from an interface-method call is wrapped with Wrap
    Given a package in "/dal/repository" with "err := r.conn.call(); return errors.Wrap(err, \"select\")"
    When the giderrwrap analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — map a boundary error to a sentinel, then a single Wrap
    Given a package in "/dal/repository" with "err := r.conn.call(); if isNoResult(err) { err = entity.ErrNotFound }; return errors.Wrap(err, \"select\")"
    When the giderrwrap analyzer checks the file
    Then no diagnostic is reported
    # Reassigning err to a sentinel before one errors.Wrap avoids wrapping twice. Unchanged in v2.

  Scenario: negative — an error from a local package function (a pure builder) is not the boundary
    Given a package in "/dal/repository" with "_, err := buildQuery(); return errors.WithMessage(err, \"build query\")"
    When the giderrwrap analyzer checks the file
    Then no diagnostic is reported
    # buildQuery is a same-module package function — neither mechanism (a) nor (b).

  Scenario: negative — a method on a concrete type is not an interface-method call
    Given a package in "/dal/repository" with "err := r.concreteHelper(); return errors.WithMessage(err, \"helper\")"
    When the giderrwrap analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative (v2) — a direct external call anywhere is wrapped with Wrap
    Given a package in "/pkg/util" with "_, err := strconv.Atoi(x); return errors.Wrap(err, \"atoi\")"
    When the giderrwrap analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative (v2) — the /event/** interface-call boundary is wrapped with Wrap
    Given a package in "/event/producer" with "err := p.client.Send(topic, msg); return errors.Wrap(err, \"kafka send\")"
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

  Scenario: non-applicability — a same-module call outside the scoped boundary layers (no mechanism a or b)
    Given a package in "/pkg/util" with "err := w.call(); return err" (call() is a same-module concrete-type method)
    When the giderrwrap analyzer checks the file
    Then no diagnostic is reported
    # Not a scoped boundary layer (no client / dal/repository / event in the path) AND call() is neither
    # an external call (mechanism a) nor an interface-method call (mechanism b).

  Scenario: non-applicability (v2) — a local package function in /event/** may use WithMessage
    Given a package in "/event/consumer" with "_, err := buildKey(topic); return errors.WithMessage(err, \"build key\")"
    When the giderrwrap analyzer checks the file
    Then no diagnostic is reported

  # === Part 2: inside the application (/domain/**) ===
  # errors.Wrap of a SAME-MODULE incoming non-static error is forbidden — its
  # stack, if any, was already collected upstream. errors.Wrap of an
  # EXTERNAL-CALL error (mechanism a) is REQUIRED instead, not forbidden (v2):
  # the domain may be the first place that reaches out to an external
  # dependency (e.g. a DB connection called directly, without an injected
  # interface). Classification of the error's source drives the outcome:
  #   external direct call  → Wrap required (Part 1 above)
  #   same-module call / parameter / interface call in domain → Wrap forbidden (WithMessage optional)

  # --- Class 1: positive ---

  Scenario: positive — Wrap of a same-module non-static error in /domain/service
    Given a package in "/domain/service" with "err := s.call(); return errors.Wrap(err, \"ctx\")"
    When the giderrwrap analyzer checks the file
    Then the diagnostic "GID-176: the stack is already collected upstream for a same-module error. Fix: use errors.WithMessage instead of errors.Wrap for an incoming error" is reported

  Scenario: positive — Wrap of an error passed as a function parameter (parameter = same-module)
    Given a package in "/domain/service" with "func (s *Service) f(err error) error { return errors.Wrap(err, \"ctx\") }"
    When the giderrwrap analyzer checks the file
    Then the diagnostic "GID-176: the stack is already collected upstream for a same-module error …" is reported

  Scenario: positive (v2) — a direct external call in /domain/service requires Wrap
    Given a package in "/domain/service" with "_, err := strconv.Atoi(x); return err"
    When the giderrwrap analyzer checks the file
    Then the diagnostic "GID-176: an error from an external call must be wrapped with errors.Wrap …" is reported on "err"
    # This is Part 1 (mechanism a), which applies inside /domain/** too — not Part 2.

  # --- Class 2: negative ---

  Scenario: negative — WithMessage for a same-module incoming error
    Given a package in "/domain/service" with "err := s.call(); return errors.WithMessage(err, \"ctx\")"
    When the giderrwrap analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative (v2) — Wrap of an external-call error inside /domain/service is now required, not forbidden
    Given a package in "/domain/service" with "_, err := strconv.Atoi(x); return errors.Wrap(err, \"parse\")"
    When the giderrwrap analyzer checks the file
    Then no diagnostic is reported
    # Before v2 this exact shape (Wrap of a non-static error in domain) was forbidden; the source is now
    # classified as external-call (mechanism a), so Wrap is correct and required, and Part 2's ban does not apply.

  # --- Class 3: boundary ---

  Scenario: boundary — Wrap of a static error from model — allowed
    Given a package in "/domain/service" with "return errors.Wrap(model.ErrSnapshotNotFound, \"ctx\")"
    When the giderrwrap analyzer checks the file
    Then no diagnostic is reported
    # Wrap of a static (package-level) error is needed — it collects the stack first. Unchanged in v2.

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
  # Unaffected by the GID-176 v2 change.

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

  # ============================================================
  # GID-237 (gidwithmessage) — new (2026-07-04, service.md)
  # ============================================================
  # In /domain/service, errors.WithMessage/WithMessagef is banned: a service
  # converts the error and wraps it with errors.WithStack; adding a message to
  # an incoming error belongs to /domain/usecase.

  # --- Class 1: positive ---

  Scenario: positive — errors.WithMessage in a service
    Given a package in "/domain/service" with "err := s.call(); return errors.WithMessage(err, \"ctx\")"
    When the gidwithmessage analyzer checks the file
    Then the diagnostic "GID-237: errors.WithMessage is not used in a service — convert the error and wrap with errors.WithStack; WithMessage belongs to usecase" is reported

  # --- Class 2: negative ---

  Scenario: negative — errors.WithStack / errors.Wrap are fine in a service
    Given a package in "/domain/service" with "err := s.call(); return errors.WithStack(err)" (and likewise errors.Wrap)
    When the gidwithmessage analyzer checks the file
    Then no diagnostic is reported

  # --- Class 3: boundary ---

  Scenario: boundary — errors.WithMessagef (the formatted variant) is banned too
    Given a package in "/domain/service" with "err := s.call(); return errors.WithMessagef(err, \"ctx %d\", n)"
    When the gidwithmessage analyzer checks the file
    Then the diagnostic "GID-237: errors.WithMessage is not used in a service …" is reported

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — errors.WithMessage in /domain/usecase (where it belongs)
    Given a package in "/domain/usecase" with "err := u.call(); return errors.WithMessage(err, \"ctx\")"
    When the gidwithmessage analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — errors.WithMessage in /dal/repository (out of scope)
    Given a package in "/dal/repository" with "err := r.call(); return errors.WithMessage(err, \"ctx\")"
    When the gidwithmessage analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — settings.exclude exempts a specific method
    Given settings.exclude contains "Service.excludedMethod" and that method returns "errors.WithMessage(err, \"ctx\")"
    When the gidwithmessage analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-176/177/237)
#  [x] Layer chosen: go/analysis (package errwrap: giderrwrap + gidstaticerr + gidwithmessage)
#  [x] Messages are defined ("GID-176: …" / "GID-177: …" / "GID-237: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability — for each analyzer
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml
