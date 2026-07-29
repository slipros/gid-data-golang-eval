# language: en

Feature: GID-104 — a constructor is named New<Entity>, not bare New (gidctor)
  As a developer
  I want every constructor to carry the name of the entity it builds
  So that entities sharing one package (all services of a layer live together)
  do not collide on a single New

  # One analyzer over the file's top-level function declarations, no type info
  # needed: the name alone decides.
  # Trigger: a receiverless function named exactly "New".
  # Exception: the composition root — a package under the "app" layer, matched
  # through internal/pathseg (anchored to the module root, not a substring):
  # by the service template the application's own New() lives there.
  # Generated code (ast.IsGenerated) is skipped.

  # Scope: only a module laid out as a service — internal/modlayout walks up to
  # the package's go.mod and looks for a layer directory (domain, dal, server,
  # app, usecase, repository) at the module root or under internal/. A flat
  # library module has no layer to point at, so the rule stays silent there.

  # --- Class 1: positive ---

  Scenario: positive — bare New in a service package
    Given "func New() *Snapshot" in /domain/service
    When the gidctor analyzer checks the file
    Then the diagnostic "GID-104: a constructor must be named New<Entity>, not bare New. Fix: rename it to New<Entity> (bare New clashes with other entities in the package)" is reported on the function name

  Scenario: positive — a package whose path merely contains "app" as a word part
    Given "func New() *Worker" in /domain/service/app
    When the gidctor analyzer checks the file
    Then the diagnostic "GID-104: …" is reported
    # The exemption is the composition root, matched as a path segment anchored
    # to the module root — not any directory with "app" in the name.

  # --- Class 2: negative ---

  Scenario: negative — the constructor names its entity
    Given "func NewSnapshot() *Snapshot"
    When the gidctor analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — the composition root
    Given "func New() *Application" in /internal/app/api
    When the gidctor analyzer checks the file
    Then no diagnostic is reported

  # --- Class 3: boundary ---

  Scenario: boundary — a method named New
    Given "func (f *Factory) New() *Job"
    When the gidctor analyzer checks the file
    Then no diagnostic is reported
    # A method is already qualified by its receiver — there is nothing to clash.

  Scenario: boundary — an unexported "new"-prefixed factory
    Given "func newJob() *Job"
    When the gidctor analyzer checks the file
    Then no diagnostic is reported
    # Only the exact name "New" is the collision the rule is about.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — a function whose name merely starts with New
    Given "func Newest(items []Job) Job"
    When the gidctor analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — generated code
    Given a file carrying the "Code generated … DO NOT EDIT." header
    When the gidctor analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — a bare New in a library module
    Given "func New() *Pool" in a module with no layer directories
    When the gidctor analyzer checks the file
    Then no diagnostic is reported
    # pool.New() is the Go idiom: the package name qualifies the constructor,
    # and no layer packs several entities into one package for it to clash with.

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-104)
#  [x] Layer chosen: go/analysis (package ctornaming)
#  [x] Message is defined ("GID-104: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
