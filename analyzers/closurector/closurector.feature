# language: en

Feature: GID-273 — a function building a closure is a constructor of it
  As a developer
  I want a function that assembles and returns a closure to carry the New/new prefix
  So that the call site reads as construction, not as fetching a value
  Source: owner requirement 2026-08-27 — consent-webhook-trigger declares
  "func (w *WebhookTriggerV2) contactFilter(scope model.WebhookTriggerV2ContactFilterScope)
  func(identifier *model.ConsentEventV2UserIdentifier) bool": the method returns
  nothing of the domain, it builds a predicate, so it is newContactFilter

  Linter: gidclosurector. LoadMode: TypesInfo (a named type over a func type is
  resolved through types.Signature).

  Judged: the result list holds a function type (a bare func type or a named
  type over one) AND the body builds it — a returned function literal, directly
  or through a local variable bound to one. A literal nested inside another
  closure belongs to that closure, not to the function under judgement.

  Not judged: a name already carrying the New/new prefix (GID-104), the options
  convention of GID-126 (a With-prefixed builder, a result of a named type with
  an Options/Option/Opt suffix — settings.option-suffixes replaces the list), a
  _test.go file (GID-250), generated code.
  Exceptions: //nolint:gidclosurector, settings.exclude ("Function" | "Type.Method").

  # --- Positive class: the violation is caught ---

  Scenario: a method assembling a predicate
    Given "/domain/usecase" declares "func (w *WebhookTriggerV2) contactFilter(scope …) func(identifier *model.ConsentEventV2UserIdentifier) bool" returning a function literal
    When the analyzer checks the file
    Then a "GID-273" diagnostic is reported telling to rename it to "newContactFilter"

  Scenario: the closure travels through a local variable
    Given "func statusFilter(status string) func(string) bool" assigns a function literal to "match" and returns it
    When the analyzer checks the file
    Then a "GID-273" diagnostic is reported

  Scenario: an exported builder
    Given "func ScopePredicate(scope string) model.Predicate" returns a function literal
    When the analyzer checks the file
    Then a "GID-273" diagnostic is reported telling to rename it to "NewScopePredicate"

  Scenario: the literal is returned from inside a branch
    Given "func branchFilter(enabled bool) func() bool" returns a literal in the if branch and nil otherwise
    When the analyzer checks the file
    Then a "GID-273" diagnostic is reported

  # --- Negative class: clean code passes ---

  Scenario: the name already carries the constructor prefix
    Given "func newContactFilter(scope string) func(string) bool" returns a function literal
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: an accessor handing out a stored callback
    Given "func (w *WebhookTriggerV2) Filter() func(identifier *model.ConsentEventV2UserIdentifier) bool" returns "w.filter"
    When the analyzer checks the file
    Then no diagnostic is reported — the callback was built elsewhere

  Scenario: a function returning no function type
    Given "func (w *WebhookTriggerV2) Scopes() []string"
    When the analyzer checks the file
    Then no diagnostic is reported

  # --- Boundary class ---

  Scenario: the options convention by prefix
    Given "func WithRetry(count int) func(identifier *model.ConsentEventV2UserIdentifier)" returns a literal
    When the analyzer checks the file
    Then no diagnostic is reported — a With-prefixed builder is named by GID-126

  Scenario: the options convention by result type
    Given "func retryOption(value string) model.SendOption" returns a literal
    When the analyzer checks the file
    Then no diagnostic is reported — the Option suffix names the convention

  Scenario: settings.option-suffixes replaces the defaults
    Given settings.option-suffixes is "['Setting']" and the package declares "func retryOption(value string) func(string)"
    When the analyzer checks the file
    Then a "GID-273" diagnostic is reported — the default Option suffix no longer exempts

  Scenario: the constructor prefix is a whole word
    Given "func newest(scope string) func() string" returns a function literal
    When the analyzer checks the file
    Then a "GID-273" diagnostic is reported — "newest" is not the New/new prefix

  Scenario: the literal belongs to a nested closure
    Given "func passthrough(base func() string) func() string" declares a closure that returns a literal, and itself returns "base"
    When the analyzer checks the file
    Then no diagnostic is reported — this function builds nothing

  # --- Non-applicability class ---

  Scenario: a _test.go helper returning a cleanup func
    Given "usecase_test.go" declares "func setup(scope string) func()"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: generated code
    Given a file with the "Code generated … DO NOT EDIT." marker declares "func genFilter(scope string) func() string"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: settings.exclude in both forms
    Given settings.exclude is "['Sender.contactFilter', 'statusFilter']"
    When the analyzer checks the file
    Then no diagnostic is reported for either

  Scenario: pinpoint exclusion via //nolint
    Given a violating declaration carries "//nolint:gidclosurector"
    When golangci-lint runs the rule
    Then no diagnostic is reported for that declaration
