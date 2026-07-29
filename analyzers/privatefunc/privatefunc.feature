# language: en

Feature: GID-133 — a private helper of one entity is its method (gidprivatefunc)
  As a developer
  I want a private function that serves a single entity to be that entity's
  method
  So that the entity's behaviour is read in one block instead of leaking into
  package-level helpers, and only genuinely shared helpers stay functions

  # One analyzer over the package's declarations, LoadModeTypesInfo (usages are
  # resolved through pass.TypesInfo.Uses, so a same-named local variable is not
  # mistaken for the function).
  # Scope: /dal/repository, /domain/service, /domain/usecase, matched through
  # internal/pathseg.EndsWith — the layer's root package.
  # Candidates: unexported package-level functions. init is exempt, and so is
  # anything declared in a _test.go file: shared test builders cannot be entity
  # methods, and the rule is about production code.
  # Ownership: a usage inside a method counts for its receiver type; a usage
  # inside a New<Entity> constructor counts for that entity (the constructor is
  # the entity's code). A usage from another package-level function establishes
  # no owner.
  # Two findings: no owner at all — the function belongs to the package; a
  # single owner — it belongs to that entity. Two or more owners is the
  # legitimate shared helper.
  # Generated code (ast.IsGenerated) is skipped.

  # --- Class 1: positive ---

  Scenario: positive — a helper used by one entity
    Given "func decorate(s string) string" called only from methods of "Snapshot" in /domain/service
    When the gidprivatefunc analyzer checks the file
    Then the diagnostic "GID-133: private function \"decorate\" is used only by entity \"Snapshot\". Fix: make it a method" is reported

  Scenario: positive — a helper used by nobody
    Given "func orphan() string" called from no method and no constructor
    When the gidprivatefunc analyzer checks the file
    Then the diagnostic "GID-133: private function \"orphan\" belongs to the package. Fix: make it a struct method (only a function shared by several entities may stay package-level)" is reported

  # --- Class 2: negative ---

  Scenario: negative — a helper shared by two entities
    Given "func normalize(s string) string" called from a method of "Snapshot" and from a method of "Job"
    When the gidprivatefunc analyzer checks the file
    Then no diagnostic is reported
    # Two owners is exactly the exception the rule grants.

  Scenario: negative — the logic already is a method
    Given "func (s *Snapshot) decorate() string" instead of a package function
    When the gidprivatefunc analyzer checks the file
    Then no diagnostic is reported

  # --- Class 3: boundary ---

  Scenario: boundary — used only from a constructor
    Given "func defaultOptions() Options" called only from "NewSnapshot"
    When the gidprivatefunc analyzer checks the file
    Then the diagnostic "GID-133: private function \"defaultOptions\" is used only by entity \"Snapshot\". …" is reported
    # A New<Entity> constructor is the entity's code, so it establishes
    # ownership just as a method does.

  Scenario: boundary — used only from another package-level function
    Given "func decorate(s string) string" called only from "func render(s string) string"
    When the gidprivatefunc analyzer checks the file
    Then the "belongs to the package" diagnostic is reported
    # A caller with no entity establishes no owner.

  Scenario: boundary — a helper declared in a _test.go file
    Given "func newTestSnapshot(t *testing.T) *Snapshot" in service_test.go
    When the gidprivatefunc analyzer checks the file
    Then no diagnostic is reported
    # Tests live in the same package (GID-250); their builders are
    # package-level by design.

  Scenario: boundary — init
    Given "func init()" in the service package
    When the gidprivatefunc analyzer checks the file
    Then no diagnostic is reported
    # It cannot be a method; banning init altogether is gochecknoinits' job
    # (GID-208).

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — a package outside the three layers
    Given "func decorate(s string) string" in /pkg/util
    When the gidprivatefunc analyzer checks the file
    Then no diagnostic is reported
    # A utility package is functions by nature.

  Scenario: non-applicability — an exported function
    Given "func Decorate(s string) string" in /domain/service
    When the gidprivatefunc analyzer checks the file
    Then no diagnostic is reported
    # The rule is about private helpers; an exported one is API and is judged
    # by other rules.

  Scenario: non-applicability — generated code
    Given a file carrying the "Code generated … DO NOT EDIT." header
    When the gidprivatefunc analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-133)
#  [x] Layer chosen: go/analysis (package privatefunc)
#  [x] Message is defined ("GID-133: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
