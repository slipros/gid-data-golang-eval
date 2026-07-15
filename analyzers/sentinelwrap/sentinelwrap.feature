# language: en

Feature: GID-244 — map a boundary error to a sentinel by reassign-then-wrap-once (sentinelwrap)
  As a developer
  I want a sentinel mapping to reassign err then wrap once, not wrap the sentinel in a guard branch
  So that the context message is written once and the boundary error is wrapped a single time

  # One analyzer, linter gidsentinelwrap, LoadModeTypesInfo.
  # pkg/errors is recognized by the import path github.com/pkg/errors (a stub in testdata).
  # Generated code (ast.IsGenerated) is skipped.
  #
  # The rule complements GID-176: GID-176 already blesses the reassign-then-Wrap
  # pattern ("if IsNoResult(err) { err = ErrNoResult }; return errors.Wrap(err, ...)")
  # but treats every errors.Wrap as correct, so it does NOT catch the anti-pattern
  # where a sentinel is wrapped directly inside a guard branch that only duplicates
  # the outer wrap. GID-244 catches exactly that shape.
  #
  # Match (deliberately narrow, to avoid false positives — all conditions required):
  #   Inside a single block:
  #     (guard)  if <pred>(err, ...) { return errors.Wrap(<staticErr>, "msg") }
  #                - the guard has NO else and NO init statement
  #                - the guard body is EXACTLY one return of a single value
  #                - <staticErr> is a package-level static error (or a named error
  #                  literal) — NOT the err variable
  #                - <pred>(err, ...) is a call whose first error-typed argument is
  #                  the err variable (e.g. epgx.IsNoResult(err), errors.Is(err, X))
  #     (mirror) return errors.Wrap(err, "msg")
  #                - the SAME err object as the guard predicate's argument
  #                - an IDENTICAL string literal message
  #   Only errors.Wrap is considered (not Wrapf); messages must be string literals.
  # Fix: collapse to `if <pred>(err) { err = <staticErr> }` + one shared errors.Wrap(err, "msg").

  # --- Class 1: positive (the violation is caught) ---

  Scenario: positive — sentinel wrapped in a guard branch duplicating the outer wrap
    Given a function with "err := t.conn.Exec(...); if err != nil { if IsNoResult(err) { return errors.Wrap(entity.ErrNoResult, \"update key\") }; return errors.Wrap(err, \"update key\") }"
    When the gidsentinelwrap analyzer checks the file
    Then the diagnostic "GID-244: a sentinel wrapped in a guard branch duplicates the outer errors.Wrap …" is reported on the guard "if"

  Scenario: positive — the predicate is errors.Is(err, target)
    Given a guard "if errors.Is(err, entity.ErrNoResult) { return errors.Wrap(entity.ErrNoResult, \"op\") }" beside "return errors.Wrap(err, \"op\")"
    When the gidsentinelwrap analyzer checks the file
    Then the diagnostic "GID-244 …" is reported on the guard "if"

  # --- Class 2: negative (clean code passes) ---

  Scenario: negative — the canonical reassign-then-wrap-once shape
    Given a function with "if IsNoResult(err) { err = entity.ErrNoResult }; return errors.Wrap(err, \"update key\")"
    When the gidsentinelwrap analyzer checks the file
    Then no diagnostic is reported
    # This is the target shape GID-244 pushes toward — never flagged.

  Scenario: negative — different context messages in the two branches
    Given a guard "if IsNoResult(err) { return errors.Wrap(entity.ErrNoResult, \"no rows\") }" beside "return errors.Wrap(err, \"exec failed\")"
    When the gidsentinelwrap analyzer checks the file
    Then no diagnostic is reported
    # Distinct messages mean the collapse is not a pure dedup — a deliberate choice, left alone.

  Scenario: negative — the guard wraps err (not a sentinel)
    Given a guard "if cond(err) { return errors.Wrap(err, \"op\") }" beside "return errors.Wrap(err, \"op\")"
    When the gidsentinelwrap analyzer checks the file
    Then no diagnostic is reported
    # Both branches wrap the same err — there is no sentinel to reassign.

  # --- Class 3: boundary (looks like a violation but is allowed) ---

  Scenario: boundary — the guard body has an extra statement (not a single return)
    Given a guard "if IsNoResult(err) { log(err); return errors.Wrap(entity.ErrNoResult, \"op\") }" beside "return errors.Wrap(err, \"op\")"
    When the gidsentinelwrap analyzer checks the file
    Then no diagnostic is reported
    # The guard does more than map-and-return; a mechanical collapse would drop the extra statement.

  Scenario: boundary — the predicate does not test the err variable (a bool flag)
    Given a guard "if useSentinel { return errors.Wrap(entity.ErrNoResult, \"op\") }" beside "return errors.Wrap(err, \"op\")"
    When the gidsentinelwrap analyzer checks the file
    Then no diagnostic is reported
    # Choosing the error by an unrelated flag cannot be rewritten as a reassignment of err.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — no mirror wrap of err in the block
    Given a guard "if IsNoResult(err) { return errors.Wrap(entity.ErrNoResult, \"op\") }" with no "return errors.Wrap(err, ...)" beside it
    When the gidsentinelwrap analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — Wrapf (formatted) is out of scope
    Given a guard "if IsNoResult(err) { return errors.Wrapf(entity.ErrNoResult, \"op %d\", n) }" beside "return errors.Wrapf(err, \"op %d\", n)"
    When the gidsentinelwrap analyzer checks the file
    Then no diagnostic is reported
    # v1 handles only errors.Wrap with string-literal messages.

  Scenario: non-applicability — settings.exclude exempts a specific method
    Given settings.exclude contains "Repo.excludedMethod"
    When the gidsentinelwrap analyzer checks that method
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-244)
#  [x] Layer chosen: go/analysis (package sentinelwrap: gidsentinelwrap)
#  [x] Message is defined ("GID-244: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml
