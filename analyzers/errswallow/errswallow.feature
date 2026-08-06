# language: en

Feature: GID-258 — in /domain/**, a function that handles an error must be able to return one
  As the styleguide owner
  I want a failure inside the application to reach the caller instead of ending in a log line
  So that an outage and an empty answer are distinguishable upstream. Inside /domain/** the
  signature is ours to change, so a function that checks `err != nil` and has no error result is
  deciding, on the callee's behalf, that the failure does not matter — and the caller, reading the
  signature, has no way to learn otherwise.

  # Layer: go/analysis (package errswallow, linter giderrswallow), LoadModeTypesInfo.
  # Scope: packages under /domain/** (internal/pathseg.HasLayer, anchored to the module root).
  # Config: settings.exclude ("Function" | "Type.Method").
  #
  # Detect: a top-level FuncDecl in /domain/** whose result list holds NO error and whose body
  # compares an error value with nil (err != nil, err == nil, if err := f(); err != nil). The
  # comparison is the proof the code KNOWS the call can fail. Reported on the declaration — the
  # fix is the signature, not the branch.
  #
  # Why scope is /domain/**: a transport handler answers to a foreign contract (http.HandlerFunc,
  # a Kafka message handler, a gRPC interceptor) and has nowhere to return an error to; judging
  # those would demand a nolint on every one of them.
  #
  # Why (incident 2026-08-06, advertising-api internal/domain/service/ad_cabinet_resolver.go):
  #   func (a *AdCabinetResolver) resolveChunk(ctx, chunk, result) {
  #       resp, err := a.registry.IntegrationsByIDs(ctx, req)
  #       if err != nil { a.logger.…Warn("resource registry unavailable…"); return }
  # A registry outage left the caller with exactly what an empty chunk leaves it with. The same
  # shape sat one level up in AdCabinet, which returns (model.AdCabinet, bool) and drops the error
  # in both branches.
  #
  # Not reported: a function that already returns error (whether it also logs is GID-155's
  # business); an error explicitly discarded (_ = f()) or never compared to nil (errcheck's
  # business); a function LITERAL — a goroutine body, an errgroup closure, a defer — whose
  # signature belongs to whoever consumes it, so the fix this rule asks for does not exist inside
  # one (only the enclosing FuncDecl is judged); _test.go, where a test reports failures through
  # *testing.T rather than a result (GID-250).
  #
  # Soft degradation ("an outage must NOT surface as an error, the caller substitutes a default")
  # is a real functional requirement now and then — declared per function with
  # //nolint:giderrswallow or settings.exclude, so the decision is visible in the code instead of
  # living in a doc comment.

  # --- Class 1: positive (the violation is caught) ---

  Scenario: positive — the incident shape: the failure is logged and the function returns nothing
    Given the method "func (r *Resolver) resolveChunk(chunk []string) { items, err := r.registry.ByIDs(chunk); if err != nil { log(\"registry unavailable\"); return }; use(items) }" in /domain/service
    When the giderrswallow analyzer checks the file
    Then the diagnostic "GID-258: the function checks an error but cannot return one — the failure ends in a log line and the caller sees the same result as a successful empty answer. Fix: add error to the results and return it (func resolve(ctx, ids) (map[K]V, error)); if a functional requirement really demands soft degradation, declare it with //nolint:giderrswallow on this function" is reported on "resolveChunk"

  Scenario: positive — results without an error: (Cabinet, bool) cannot carry the failure either
    Given the method "func (r *Resolver) Cabinet(id string) (Cabinet, bool) { cabinet, err := r.registry.One(id); if err != nil { log(\"registry unavailable\"); return Cabinet{}, false }; return cabinet, true }" in /domain/service
    When the giderrswallow analyzer checks the file
    Then the diagnostic "GID-258: the function checks an error but cannot return one …" is reported on "Cabinet"
    # The caller reads "not found" for what was an outage.

  # --- Class 2: negative (clean code passes) ---

  Scenario: negative — the error is returned
    Given the method "func (r *Resolver) resolveChunkClean(chunk []string) (map[string]Cabinet, error) { items, err := r.registry.ByIDs(chunk); if err != nil { return nil, err }; … }"
    When the giderrswallow analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — a NAMED error result counts just as much
    Given the method "func (r *Resolver) resolveNamed(chunk []string) (err error) { _, err = r.registry.ByIDs(chunk); if err != nil { return err }; return nil }"
    When the giderrswallow analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — nothing failable is checked
    Given the method "func (r *Resolver) titles(cabinets []Cabinet) []string { … }"
    When the giderrswallow analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — the error is explicitly discarded, never compared to nil
    Given the method "func (r *Resolver) fireAndForget(ids []string) { _, _ = r.registry.ByIDs(ids) }"
    When the giderrswallow analyzer checks the file
    Then no diagnostic is reported
    # There is no handling to speak of; the discard itself is errcheck's business.

  # --- Class 3: boundary ---

  Scenario: boundary — the comparison written the other way round (err == nil)
    Given the method "func (r *Resolver) countKnown(ids []string) int { items, err := r.registry.ByIDs(ids); if err == nil { return len(items) }; return 0 }"
    When the giderrswallow analyzer checks the file
    Then the diagnostic "GID-258: the function checks an error but cannot return one …" is reported on "countKnown"
    # Either direction of the comparison proves the same thing: the code knows the call can fail.

  Scenario: boundary — the check lives inside a function LITERAL (a goroutine body)
    Given the method "func (r *Resolver) resolveAsync(ids []string) { go func() { if _, err := r.registry.ByIDs(ids); err != nil { log(\"async resolve failed\") } }() }"
    When the giderrswallow analyzer checks the file
    Then no diagnostic is reported
    # A literal's signature belongs to whoever consumes it — there is no result list to fix.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — the same shape outside /domain/**
    Given the method "func (h *Handler) Handle(payload string) { if err := h.sender.Send(payload); err != nil { log(\"send failed\") } }" in /server/grpc
    When the giderrswallow analyzer checks the file
    Then no diagnostic is reported
    # A transport handler answers to a foreign contract and has nowhere to return an error to.

  Scenario: non-applicability — a _test.go helper that logs and moves on
    Given the helper "func checkResolve(r *Resolver, ids []string) { if _, err := r.registry.ByIDs(ids); err != nil { log(\"fixture setup failed\") } }" in service_test.go
    When the giderrswallow analyzer checks the file
    Then no diagnostic is reported
    # A test reports failures through *testing.T, not through a result (GID-250).

  # --- Config: settings.exclude ---

  Scenario: config — soft degradation declared as a functional requirement
    Given settings.exclude contains "Resolver.degradeSilently" and that method logging the error and moving on
    When the giderrswallow analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-258)
#  [x] Layer chosen: go/analysis (package errswallow: giderrswallow)
#  [x] Message is defined ("GID-258: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
