# language: en

Feature: GID-263 — code that carries logic carries a test (gidhastest)
  As a developer
  I want every non-trivial exported function of the module to be exercised by a
  test of its own package
  So that logic does not reach production with nobody having run it once, and
  the gap is named at the declaration rather than in a coverage report nobody
  opens

  # One analyzer over the package's declarations, LoadModeTypesInfo (a usage is
  # resolved through pass.TypesInfo.Uses, so a same-named method of another type
  # is not mistaken for the candidate).
  #
  # The linter does NOT run tests: it reads the package's own _test.go files and
  # asks whether the candidate is mentioned there at all. That is a one-sided
  # proxy for coverage — "never mentioned" proves "never executed", while
  # "mentioned" proves nothing about branches. The rule is deliberately built on
  # the side that cannot produce a false alarm.
  #
  # The analyzer never touches the file system. It judges the pass it is given
  # and nothing else.
  #
  # Package variants. With run.tests: true go/packages hands out two variants of
  # a package that has tests — the base one (production files only) and
  # "pkg [pkg.test]" (production + _test.go). golangci-lint keeps only the second
  # (filterDuplicatePackages), and that is the pass this rule judges.
  #
  # A pass WITHOUT _test.go files is ambiguous — a package with no tests, or a
  # package whose tests were withheld (the base variant; every package under
  # run.tests: false) — and nothing inside the pass separates the two. So such a
  # pass is left alone. Consequences, both deliberate: under run.tests: false the
  # rule reports nothing at all, and a package with NO test file is out of reach
  # in every mode. Listing those packages needs no source analysis anyway:
  #   go list -f '{{if not .TestGoFiles}}{{.ImportPath}}{{end}}' ./...
  #
  # Candidates: top-level exported FuncDecls of production files whose body is
  # NON-TRIVIAL. Trivial (never judged) is a body of at most one statement whose
  # expression holds no binary operator and at most one call — a getter
  # (return s.id), an enum String (return string(e)), a delegation
  # (return s.repo.Create(ctx, m)) and a one-line constructor. Testing those
  # tests the mock, not the code.
  #
  # Skipped whole: generated files (ast.IsGenerated), the composition root
  # (package main and /app/**), the synthesized pkg.test package, and packages
  # matched by settings.exclude-paths.
  #
  # _test.go is not a source of candidates: an exported helper of a test suite
  # is scaffolding and has no test of its own (the "skips _test.go" side of the
  # RULES.md test-file split).

  # --- Class 1: positive ---

  Scenario: positive — a method of a tested package that no test mentions
    Given package "service" with "segment_test.go" exercising "Create"
    And the exported method "func (s *Segment) Rebuild(ctx context.Context) error" with a branching body
    When the gidhastest analyzer checks the package
    Then the diagnostic "GID-263: exported method Segment.Rebuild is not exercised by any test of this package. Fix: add func TestSegment_Rebuild(t *testing.T) calling it" is reported on the declaration

  # --- Class 2: negative ---

  Scenario: negative — the method is called by a test
    Given the exported method "Create" called as "svc.Create(ctx, in)" inside "TestSegment_Create"
    When the gidhastest analyzer checks the package
    Then no diagnostic is reported

  Scenario: negative — the method is only handed over as a value
    Given the exported method "Rebuild" passed to a helper as "svc.Rebuild" without being called
    When the gidhastest analyzer checks the package
    Then no diagnostic is reported
    # A mention is a mention: the test drives it through the helper.

  Scenario: negative — the function runs through the initializer of a var the test uses
    Given "var DefaultLimits = NewLimits(10, 20)" and a test naming DefaultLimits
    When the gidhastest analyzer checks the package
    Then no diagnostic is reported for "NewLimits"
    # It runs at package initialization, so "not exercised" would be false. One
    # step, and only through a var: "reachable from a test" is real coverage,
    # which no linter settles without running the tests.

  # --- Class 3: boundary ---

  Scenario: boundary — a trivial getter is not a candidate
    Given the exported method "func (s *Segment) ID() uuid.UUID { return s.id }" mentioned by no test
    When the gidhastest analyzer checks the package
    Then no diagnostic is reported

  Scenario: boundary — a one-line delegation is not a candidate
    Given the exported method "func (s *Segment) Delete(ctx context.Context, id uuid.UUID) error { return s.repo.Delete(ctx, id) }" mentioned by no test
    When the gidhastest analyzer checks the package
    Then no diagnostic is reported
    # Its test would assert that the mock was called.

  Scenario: boundary — an unexported function is not a candidate
    Given the unexported function "func normalize(s string) string" with a branching body and no test
    When the gidhastest analyzer checks the package
    Then no diagnostic is reported
    # It is reached through the exported surface that the rule does judge.

  Scenario: boundary — a pass that holds no test file
    Given a package whose pass.Files hold no _test.go — it has none, or the run withheld them (run.tests: false)
    When the gidhastest analyzer checks the package
    Then no diagnostic is reported
    # Separating the two cases takes a look at the file system, and looking there
    # to judge tests the run excluded works around the setting instead of
    # honouring it.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — generated code
    Given a file carrying the "Code generated by protoc-gen-go. DO NOT EDIT." marker
    When the gidhastest analyzer checks the package
    Then no diagnostic is reported

  Scenario: non-applicability — the composition root
    Given package "main" and package "app" wiring the dependencies together
    When the gidhastest analyzer checks the package
    Then no diagnostic is reported
    # Wiring is proved by the service starting, not by a unit test.

  Scenario: non-applicability — an exported helper declared in a _test.go file
    Given "func NewSegmentFixture(t *testing.T) *Segment" declared in "helpers_test.go"
    When the gidhastest analyzer checks the package
    Then no diagnostic is reported

  Scenario: non-applicability — the method is on settings.exclude
    Given settings.exclude holding "Segment.Rebuild"
    When the gidhastest analyzer checks the package
    Then no diagnostic is reported

  Scenario: non-applicability — the package is under settings.exclude-paths
    Given settings.exclude-paths holding "internal/generated"
    When the gidhastest analyzer checks a package under that path
    Then no diagnostic is reported
