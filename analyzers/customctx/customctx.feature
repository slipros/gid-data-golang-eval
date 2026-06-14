# language: en

Feature: GID-188 — ban on custom context types (no-custom-context)
  As a developer
  I want only context.Context to be used in the ctx position and in interface embedding
  So that custom contexts do not proliferate and data is passed via context.WithValue

  # Google rule: "custom contexts — no exceptions".
  # Analyzer gidcustomctx, LoadMode TypesInfo. Detection (pass.TypesInfo):
  #   1. a declaration of a named type (struct/interface) in the checked package
  #      whose method set covers context.Context (Deadline/Done/Err/Value)
  #      — via types.Implements against the stdlib context.Context interface;
  #   2. an interface type EMBEDDING context.Context (embedded in the declaration);
  #   3. a parameter of a function/function literal named ctx whose type is a
  #      named non-stdlib type (not context.Context).
  # The context.Context interface is taken from the package imports (direct/transitive);
  # if context is not imported anywhere, cases 1 and 2 do not apply.
  # Generated code (ast.IsGenerated) is skipped.
  # Targeted suppression — the standard //nolint:gidcustomctx.

  # === Class 1: positive (violations) ===

  Scenario: positive — an interface embeds context.Context
    Given the declaration "type MyContext interface { context.Context; Extra() string }"
    When the analyzer checks the file
    Then the diagnostic "GID-188: custom context type MyContext is forbidden. Fix: pass context.Context and store data via context.WithValue (helpers live in /domain/model, GID-165/166)." is reported

  Scenario: positive — a struct with the full set of context.Context methods
    Given the type "CtxStruct" with the methods "Deadline/Done/Err/Value" matching the context.Context signatures
    When the analyzer checks the file
    Then the diagnostic "GID-188: custom context type CtxStruct is forbidden ..." is reported

  Scenario: positive — a ctx parameter of a custom type
    Given the function "func f(ctx MyCtx)" where MyCtx is a named non-stdlib type
    When the analyzer checks the file
    Then the diagnostic "GID-188: parameter ctx has type <MyCtx>. Fix: use context.Context." is reported

  # === Class 2: negative (clean code) ===

  Scenario: negative — the ctx parameter is context.Context
    Given the function "func f(ctx context.Context)"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — a struct with a Done method but not the full set
    Given the type "PartialCtx" with the single method "Done() <-chan struct{}"
    When the analyzer checks the file
    Then no diagnostic is reported
    # The method set does not cover context.Context — types.Implements is false.

  # === Class 3: boundary ===

  Scenario: boundary — interface { context.Context } is matched once
    Given the declaration "type OnlyEmbed interface { context.Context }"
    When the analyzer checks the file
    Then exactly one diagnostic "GID-188: custom context type OnlyEmbed is forbidden ..." is reported
    # The embedding case (2) fires and prevents a duplicate match via types.Implements (1).

  Scenario: boundary — methods Deadline/Done/Err/Value with different signatures
    Given the type "FakeCtx" with the methods "Deadline() string, Done() bool, Err() string, Value(int) int"
    When the analyzer checks the file
    Then no diagnostic is reported
    # The names match, but the signatures are not equal to context.Context — not matched.

  Scenario: boundary — a ctx parameter of the stdlib type next to a non-ctx parameter
    Given the function "func f(ctx context.Context, other FakeCtx)"
    When the analyzer checks the file
    Then no diagnostic is reported
    # The name ctx is reserved for context.Context; the parameter other is not checked by name.

  # === Class 4: non-applicability ===

  Scenario: non-applicability — a package without context
    Given a package that does not import context and has no context-like types
    When the analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-188)
#  [x] Layer chosen: go/analysis (analyzer gidcustomctx in analyzers/customctx)
#  [x] Severity and message are defined ("GID-188: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml
