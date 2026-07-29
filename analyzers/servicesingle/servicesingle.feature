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
