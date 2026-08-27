# language: en

Feature: GID-272 — a function or method in the domain layer takes at most 3 substantive arguments
  As the styleguide owner
  I want a domain function or method to accept at most 3 substantive arguments
  So that a function stays a conversion instead of becoming assembly — many parameters
  mean the caller is doing the grouping the code should do, and a type carrying the related
  fields is missing.

  # Layer: go/analysis (package arglimit, linter gidarglimit), LoadModeTypesInfo
  # (context.Context is recognised by its type).
  # Config: settings.max-args (default 3 — a violation begins at 4),
  # settings.exclude ("Function" | "Type.Method").
  #
  # Detect: a FuncDecl in a package under /domain/** (internal/domain/... and
  # pkg/<module>/domain/... alike, matched by path segments — internal/pathseg),
  # package-level or a method (the receiver is not an argument), with more than
  # max-args substantive parameters. Each named parameter is one argument, an
  # unnamed one is one too, a variadic tail is one; context.Context parameters
  # are technical and do not count (owner decision).
  #
  # Not reported: a constructor (the New/new prefix by GID-104 — it takes as
  # many dependencies as its entity has), a _test.go file (GID-250), a package
  # outside /domain/**.
  #
  # Why (incident 2026-08-27, consent-webhook-trigger):
  #   func WebhooksTriggersV2FromConsentEventV2(
  #       organizationID string, triggers map[string][]Trigger,
  #       disabled map[string]bool, events map[string][]Event,
  #       fallback map[string]string, limit int,
  #   ) []TriggerV2
  # Six arguments, four of them maps keyed by the same organizationID — the
  # four maps are one thing (the events of one organization) and ask to be
  # grouped into a type. Measured on six udmp repositories (1485 functions and
  # methods in /domain/**): 28 violations at the >=4 boundary (1.9%), so the
  # rule bites rarely and never on the canonical shapes.

  # --- Class 1: positive (the violation is caught) ---

  Scenario: positive — exactly 4 arguments
    Given a function with the signature "func four(a, b int, c, d string)"
    When the gidarglimit analyzer checks the file
    Then a "GID-272" diagnostic is reported on "four" naming the count 4

  Scenario: positive — a method with 4 arguments; the receiver is not an argument
    Given a method with the signature "func (m *Model) methodFour(a, b, c, d int)"
    When the gidarglimit analyzer checks the file
    Then a "GID-272" diagnostic is reported on "methodFour"

  Scenario: positive — the incident shape: six arguments, four maps keyed by one id
    Given a function "WebhooksTriggersV2FromConsentEventV2" with 6 parameters
    When the gidarglimit analyzer checks the file
    Then a "GID-272" diagnostic is reported on "WebhooksTriggersV2FromConsentEventV2"

  Scenario: positive — a variadic tail counts as one argument
    Given a function with the signature "func variadic(a, b, c int, rest ...string)"
    When the gidarglimit analyzer checks the file
    Then a "GID-272" diagnostic is reported on "variadic" naming the count 4

  # --- Class 2: negative (clean code passes) ---

  Scenario: negative — 0 to 3 arguments
    Given a function with the signature "func three(a, b int, c string)"
    When the gidarglimit analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — context.Context does not count
    Given a function with the signature "func withCtx(ctx context.Context, a, b, c int)"
    When the gidarglimit analyzer checks the file
    Then no diagnostic is reported

  # --- Class 3: boundary ---

  Scenario: boundary — exactly 3 arguments is the maximum allowed
    Given a function with the signature "func maxAllowed(a, b, c int)"
    When the gidarglimit analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — exactly 4 arguments is already a violation
    Given a function with the signature "func four(a, b int, c, d string)"
    When the gidarglimit analyzer checks the file
    Then a "GID-272" diagnostic is reported on "four"

  Scenario: boundary — 3 arguments plus ctx stay 3 and pass
    Given a function with the signature "func withCtx(ctx context.Context, a, b, c int)"
    When the gidarglimit analyzer checks the file
    Then no diagnostic is reported

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — a constructor is not judged
    Given a function with the signature "func NewModel(a, b, c, d, e, f int)"
    When the gidarglimit analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — a _test.go file is not judged
    Given a function "buildFixture" with 5 parameters in a "_test.go" file
    When the gidarglimit analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — a package outside /domain/** is not judged
    Given a function "buildEntity" with 5 parameters in the package "svc/dal/entity"
    When the gidarglimit analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — settings.exclude clears a function or a method
    Given settings.exclude is ["legacyConvert", "Converter.legacyMethod"]
    When the gidarglimit analyzer checks a file where both take 4 arguments
    Then no diagnostic is reported on either of them

  Scenario: non-applicability — a custom threshold is honoured
    Given settings.max-args is 1
    When the gidarglimit analyzer checks a function with the signature "func two(a, b int)"
    Then a "GID-272" diagnostic is reported on "two" naming the count 2
