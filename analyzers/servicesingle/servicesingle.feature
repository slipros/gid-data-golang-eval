# language: en

Feature: GID-148 — a service serves one entity and does not depend on another service (gidservicesingle)
  As a developer
  I want a domain service devoted to a single entity
  So that orchestrating several entities happens in a usecase, where the
  transaction and the order of steps are visible in one place

  # One analyzer over struct declarations, LoadModeTypesInfo (the field type is
  # resolved, so a pointer and a named type are handled alike).
  # Scope: the root of /domain/service, matched through
  # internal/pathseg.EndsWith.
  # Deterministic trigger: a struct field whose type is another struct of the
  # same package — in this layout every service of the layer lives in one
  # package, so such a field is a service holding a service.
  # Exempt: a type whose name ends with Options — that is the entity's own
  # options struct (GID-152), not a dependency.
  # Generated code (ast.IsGenerated) is skipped.

  # --- Class 1: positive ---

  Scenario: positive — a pointer to another service
    Given "type Upload struct { snapshots *Snapshot }" in /domain/service
    When the gidservicesingle analyzer checks the file
    Then the diagnostic "GID-148: service \"Upload\" depends on service \"Snapshot\". Fix: a service serves one entity, orchestrate multiple services in usecase" is reported on the field

  Scenario: positive — another service held by value
    Given "type Upload struct { jobs Job }"
    When the gidservicesingle analyzer checks the file
    Then the diagnostic "GID-148: service \"Upload\" depends on service \"Job\". …" is reported

  # --- Class 2: negative ---

  Scenario: negative — the service holds a repository and a logger
    Given "type Snapshot struct { repo Repository; logger *logrus.Entry }"
    When the gidservicesingle analyzer checks the file
    Then no diagnostic is reported
    # The repository comes in as an interface, from another package.

  Scenario: negative — orchestration in a usecase
    Given "type Upload struct { snapshots *service.Snapshot; jobs *service.Job }" in /domain/usecase
    When the gidservicesingle analyzer checks the file
    Then no diagnostic is reported
    # That is exactly what a usecase is for.

  # --- Class 3: boundary ---

  Scenario: boundary — an Options field
    Given "type Snapshot struct { opts *SnapshotOptions }"
    When the gidservicesingle analyzer checks the file
    Then no diagnostic is reported
    # A name ending with Options is the entity's own options struct.

  Scenario: boundary — a struct from another package
    Given "type Snapshot struct { client *client.Trino }"
    When the gidservicesingle analyzer checks the file
    Then no diagnostic is reported
    # Only same-package structs mean a service-on-service dependency here.

  Scenario: boundary — an interface field of the same package
    Given "type Upload struct { snapshots SnapshotProvider }" where SnapshotProvider is an interface
    When the gidservicesingle analyzer checks the file
    Then no diagnostic is reported
    # An interface declared next to its consumer is GID-134's shape, and it is
    # not a concrete service.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — a package outside /domain/service
    Given the same struct declared in /dal/repository
    When the gidservicesingle analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — generated code
    Given a file carrying the "Code generated … DO NOT EDIT." header
    When the gidservicesingle analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-148)
#  [x] Layer chosen: go/analysis (package servicesingle)
#  [x] Message is defined ("GID-148: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml

  Scenario: positive — a service constructs a different service in a method
    Given a Job method creates Snapshot using "&Snapshot{repo: s.repo}", "Snapshot{}", "new(Snapshot)" or "NewSnapshot(s.repo)"
    When the gidservicesingle analyzer checks the file
    Then GID-148 names Job and Snapshot and asks to move construction and orchestration to usecase

  Scenario: boundary — aliases and value receivers
    Given a value-receiver Job method creates a type alias of Snapshot
    When the gidservicesingle analyzer checks the file
    Then GID-148 names the underlying Snapshot service

  Scenario: negative — constructors, own type and Options
    Given package-level New functions assemble services and a Job method creates Job or SnapshotOptions
    When the gidservicesingle analyzer checks the file
    Then no construction diagnostic is reported

  Scenario: non-applicability — a test harness constructs services in a method
    Given a test harness method creates Snapshot in a _test.go file
    When the gidservicesingle analyzer checks the file
    Then no diagnostic is reported

  Scenario: positive — interface injection in composition root
    Given a service constructor takes a local interface
    And app supplies a concrete service or usecase through that interface
    When gidservicesingle checks the wiring package
    Then GID-148 asks to move orchestration to usecase

  Scenario: boundary — a single-assignment interface alias
    Given a concrete service is assigned to an interface variable and then aliased
    When another service is constructed with that alias
    Then GID-148 is reported

  Scenario: negative — repositories and unknown interfaces
    Given a service receives a repository or an interface parameter of unknown implementation
    When the analyzer runs
    Then no speculative wiring diagnostic is reported

  Scenario: non-applicability — converter helpers and options
    Given a type in domain/service/convert or a type named Options or ending in Options
    When it is constructed or passed to a service
    Then it is not classified as a domain service dependency
