# language: en

Feature: GID-278 — pure logic over model values is a model method
  As a developer
  I want a service/usecase function that works only with model values and
  touches neither its receiver nor its package to become a public method of the model type
  So that a service that only carries a comparison or a predicate over two
  model values is not created, and the behavior lives where the data lives

  # Semantics (owner requirement of 2026-09-30, incident: SegmentChange.SegmentChanged(current, snapshot *model.SegmentMetric) bool):
  # - scope: the roots of /domain/service and /domain/usecase (EndsWith); subpackages (convert/) are not affected;
  # - trigger: a function or method (exported or not) that
  #     * has a parameter T or *T, T a named non-interface type of the model layer (struct, enum) — the future receiver;
  #     * takes only plain data: no interface, func or channel anywhere in a parameter type
  #       (context.Context, a repository, a callback mean the call talks to the outside);
  #     * does not use its receiver (unnamed and "_" count as unused);
  #     * refers to no symbol of its own package, signature included;
  # - the owner type must be declared in the module under analysis: a method cannot be added to a foreign module's type;
  # - widens GID-195: the private function with exactly one parameter stays GID-195's, never reported twice;
  # - not flagged: a method required by an interface declared in the same package; a generic function;
  #   a declaration without a body; a _test.go file (a double mirrors the interface it fakes, GID-250);
  # - exclusions: //nolint:gidpuremodel or settings.exclude ("Function" | "Type.Method").

  Scenario: positive — an exported method over two model values
    Given the package path ends with the segments "domain/service"
    And the method "(s *SegmentChange) SegmentChanged(current, snapshot *model.SegmentMetric) bool" does not access "s"
    When the analyzer checks the package
    Then a "GID-278" diagnostic is reported with a hint to make it a public method of model.SegmentMetric

  Scenario: positive — an unnamed receiver and a single model value
    Given the methods "(*SegmentChange) Same(a, b *model.SegmentMetric)" and "(n *Notifier) Title(m *model.SegmentMetric)" are declared
    When the analyzer checks the package
    Then a "GID-278" diagnostic is reported on each

  Scenario: positive — a private method with two parameters, an enum with plain data, free functions
    Given "grew(before, after model.SegmentMetric)", "Reached(st model.Status, limit int)", "SegmentDiffers(a, b *model.SegmentMetric)" are declared
    When the analyzer checks the package
    Then a "GID-278" diagnostic is reported on each

  Scenario: positive — the usecase root
    Given the package path ends with the segments "domain/usecase"
    And a method over two model values that does not access its receiver is declared
    When the analyzer checks the package
    Then a "GID-278" diagnostic is reported

  Scenario: negative — the method uses its receiver
    Given the method "Decorate" reads "n.prefix"
    When the analyzer checks the package
    Then no diagnostic is reported

  Scenario: negative — a private function with one model parameter
    Given the function "onlyChecksum(m *model.SegmentMetric)" is declared
    When the analyzer checks the package
    Then no GID-278 diagnostic is reported
    # GID-195 judges it

  Scenario: negative — the function depends on its own package
    Given "Tag" uses a package constant and "Wrap" returns a type of the package
    When the analyzer checks the package
    Then no diagnostic is reported
    # moving it would create a reverse dependency model → service

  Scenario: negative — the call talks to the outside
    Given "Touch(ctx context.Context, m *model.SegmentMetric)", "Check(v model.Validator, m *model.SegmentMetric)" and "Each(m *model.SegmentMetric, fn func(string))" are declared
    When the analyzer checks the package
    Then no diagnostic is reported

  Scenario: negative — the method is required by an interface of the package
    Given the interface "Differ" declares "Differs(a, b *model.SegmentMetric) bool"
    And "checksumDiffer" implements it
    When the analyzer checks the package
    Then no diagnostic is reported on "checksumDiffer.Differs"

  Scenario: boundary — nothing to attach the method to
    Given "Sum(a, b int)", "Names([]model.SegmentMetric)", "Join(...model.SegmentMetric)" and "Validate(a, b model.Validator)" are declared
    When the analyzer checks the package
    Then no diagnostic is reported

  Scenario: boundary — a parameter of the package's own type, a generic function, a bodiless declaration
    Given "OptionsName(o *Options, m *model.SegmentMetric)", "Pick[T any](a, b T, m *model.SegmentMetric)" and "external(a, b *model.SegmentMetric) bool" are declared
    When the analyzer checks the package
    Then no diagnostic is reported

  Scenario: non-applicability — /dal/repository and the convert subpackage of service
    Given the package path ends with "dal/repository" or "domain/service/convert"
    And a method over two model values that does not access its receiver is declared
    When the analyzer checks the package
    Then no GID-278 diagnostic is reported

  Scenario: non-applicability — a _test.go file
    Given a test double in a _test.go file of domain/service with a method over two model values
    When the analyzer checks the package
    Then no diagnostic is reported

  Scenario: non-applicability — settings.exclude
    Given the linter setting "exclude: [Legacy.Changed, LegacyChanged]"
    When the analyzer checks the package
    Then no diagnostic is reported on "Legacy.Changed" and "LegacyChanged"
    And the non-excluded "Legacy.Reported" is still reported
