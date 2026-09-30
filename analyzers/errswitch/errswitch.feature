# language: en

Feature: GID-279 — a switch over an error sits inside `if err != nil`
  As a developer
  I want an error to be checked with if first and only then classified
  So that the success path is not one of the branches that classify failures

  # Semantics (owner decision of 2026-09-30; incident: ad-cabinet-connector,
  # switch { case err == nil: return id, nil; case errors.Is(err, …): … } with no `if err != nil` before it):
  # - scope: the whole module, generated code and _test.go files excluded;
  # - trigger: a switch that examines an error VARIABLE (a local variable or a field of an error INTERFACE
  #   type — a concrete error type is the target of errors.As, not an examined error) — in a case condition of a tagless switch (err == nil, errors.Is(err, X),
  #   errors.As(err, &t), ready && err == nil), as the tag (switch err { case ErrX: }) or as the operand of a
  #   type switch — and is not inside the body of an enclosing `if` that checked that same variable:
  #   `if err != nil`, a conjunct of && (`err != nil && ready`), an init form (`if err := f(); err != nil`),
  #   or the else branch of `if err == nil`;
  # - a guard for another variable, an `if` on something else, the else branch of `err != nil` do not count;
  # - the guard may be any level up: a loop, a closure or a nested if inside the guarded body is covered;
  # - errors.Is / errors.As / `ctx.Err() != nil` are fine as cases — inside the guard;
  # - not judged: a PARAMETER (the caller has checked it: errorResponse(ctx, f, err), isRetryable(err error)),
  #   a call result (`ctx.Err() != nil`, `c.failure() != nil`), a package-level sentinel on its
  #   own, a non-error nil, a switch with an error only in a clause body, a type switch over any;
  # - reported on the switch keyword; exclusions: //nolint:giderrswitch.

  Scenario: positive — no guard before the switch
    Given "switch { case err == nil: return id, nil; case errors.Is(err, ErrRate): …; default: return 0, err }" with no enclosing "if err != nil"
    When the analyzer checks the package
    Then a "GID-279" diagnostic names the variable and suggests checking the error with if first

  Scenario: positive — errors.Is, errors.As, a field, a compound condition, a tagged switch, a type switch
    Given each of them is unguarded
    When the analyzer checks the package
    Then a "GID-279" diagnostic is reported on each switch

  Scenario: positive — the switch is in a loop, under an if about something else, or in the wrong branch
    Given "for { switch { case err == nil: … } }", "if ready { switch { case errors.Is(err, X): … } }", the else of "if err != nil", and a guard on another variable
    When the analyzer checks the package
    Then a "GID-279" diagnostic is reported on each

  Scenario: negative — the classes are narrowed inside the guard
    Given "if err != nil { switch { case errors.Is(err, ErrRate): err = ErrRejected }; return 0, err }"
    When the analyzer checks the package
    Then no diagnostic is reported

  Scenario: negative — the guard in other forms
    Given "if _, err := f(); err != nil", "if err != nil && ready", "if err == nil { … } else { switch … }", "if (err != nil)" and a field guard "res.Err != nil"
    When the analyzer checks the package
    Then no diagnostic is reported

  Scenario: negative — deeper inside the guard
    Given a loop, a closure and a nested if inside the body of "if err != nil" contain a switch over err
    When the analyzer checks the package
    Then no diagnostic is reported

  Scenario: boundary — a parameter is not judged
    Given "func errorResponse(err error) int { switch { case errors.Is(err, ErrRate): … } }", a "switch err" predicate and a closure taking err
    When the analyzer checks the package
    Then no diagnostic is reported
    # the error has been checked where it was received

  Scenario: boundary — not an error variable
    Given "errors.As(err, &target)" with a concrete *rateError target, "case ctx.Err() != nil", "case c.failure() != nil", "case ErrRate == ErrNoTable", "case p == nil" for a pointer, "case kind == 1: return err" and "switch v.(type)" over any
    When the analyzer checks the package
    Then no diagnostic is reported

  Scenario: non-applicability — a _test.go file
    Given a test with "switch { case err == nil: return; default: t.Fatal(err) }"
    When the analyzer checks the package
    Then no diagnostic is reported
