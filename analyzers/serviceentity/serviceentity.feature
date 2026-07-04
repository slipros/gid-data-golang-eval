# language: en

Feature: GID-236 — a service works with exactly one entity
  As a service architect
  I want a domain service to inject only the repository of its own entity
  So that cross-entity orchestration stays in usecase, not scattered across services

  # Semantics (service.md: "Each service works with exactly one specific entity";
  # "By default a service calls the repository of its own entity, unless
  # explicitly stated otherwise"). Complements GID-148 (servicesingle: a service
  # does not depend on another service) — this rule catches a service injecting
  # a repository of a FOREIGN entity.
  #
  # - Scope: packages whose path ends with the segments domain/service
  #   (the root of the layer, not its subpackages).
  # - For every struct declared in the package (except *Options types), for
  #   every field whose type is a named interface declared in the SAME
  #   package (GID-134: interfaces live at the consumer — an interface from
  #   another package is not checked) with a name ending in one of
  #   settings.suffixes (default ["Repository"]): the entity is that name
  #   with the suffix stripped; if it differs from the struct's own name —
  #   a violation on the field (embedded fields count via the interface's
  #   type name).
  # - Exclusions: //nolint:gidserviceentity (service.md itself allows an
  #   explicit exception) or settings.exclude ("Struct" as a whole |
  #   "Struct.Field").

  Scenario: positive — a named field injects a foreign entity's repository
    Given the package path ends with the segments "domain/service"
    And the struct "Upload" has a field "jobs" of type "JobRepository", an interface declared in the same package
    When the analyzer checks the package
    Then a "GID-236" diagnostic is reported on the field "jobs"

  Scenario: positive — an embedded foreign-entity repository
    Given the struct "Delivery" embeds the interface "JobRepository"
    When the analyzer checks the package
    Then a "GID-236" diagnostic is reported on the embedded field

  Scenario: negative — the service's own repository, plus Options and a non-repository interface
    Given the struct "Snapshot" has a field "repository" of type "SnapshotRepository"
    And it also has a field "validator" of type "SnapshotValidator" (no "Repository" suffix)
    And it also has a field "opts" of type "Options"
    When the analyzer checks the package
    Then no diagnostic is reported for these fields

  Scenario: boundary — a repository entity that only shares a name prefix with the owner
    Given the struct "Snapshot" has a field "files" of type "SnapshotFileRepository"
    When the analyzer checks the package
    Then a "GID-236" diagnostic is reported on the field "files"
    # "SnapshotFile" (entity after stripping the suffix) != "Snapshot" — still foreign

  Scenario: boundary — an *Options type is skipped entirely
    Given the struct "SnapshotOptions" has a field "Jobs" of type "JobRepository"
    When the analyzer checks the package
    Then no diagnostic is reported
    # *Options types are configuration, not entity structs — skipped upfront

  Scenario: negative — an interface declared in another package is not checked
    Given the struct "Report" has a field "files" of type "repository.FileRepository" from another package
    When the analyzer checks the package
    Then no diagnostic is reported
    # GID-134 scope: only interfaces declared next to the consumer are in play

  Scenario: non-applicability — the same shape outside /domain/service
    Given the package path ends with the segments "domain/usecase"
    And a struct there has a field of a foreign-entity repository interface
    When the analyzer checks the package
    Then no GID-236 diagnostic is reported
    # a usecase legitimately orchestrates several entities

  Scenario: settings.suffixes — a custom repository-name suffix is honored
    Given the linter setting "suffixes: [Repository, Store]" is set in .golangci.yml
    And the struct "Upload" has a field "jobs" of type "JobStore"
    When the analyzer checks the package
    Then a "GID-236" diagnostic is reported on the field "jobs"

  Scenario: settings.exclude — a whole struct is skipped
    Given the linter setting "exclude: [LegacySnapshot]" is set in .golangci.yml
    And the struct "LegacySnapshot" has a foreign-entity repository field
    When the analyzer checks the package
    Then no diagnostic is reported

  Scenario: settings.exclude — a single field is skipped, other violations still caught
    Given the linter setting "exclude: [Delivery.jobs]" is set in .golangci.yml
    And the struct "Delivery" has a field "jobs" (excluded) and a field "others" of a foreign-entity repository
    When the analyzer checks the package
    Then no diagnostic is reported on "jobs"
    And a "GID-236" diagnostic is reported on "others"

  Scenario: pointwise escape — //nolint:gidserviceentity
    Given a field injecting a foreign entity's repository is annotated "//nolint:gidserviceentity"
    When golangci-lint runs
    Then the diagnostic is suppressed on that line
    # service.md explicitly allows "unless stated otherwise" — the nolint comment is that statement

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the RULES.md registry (GID-236, service-one-entity)
#  [x] Layer chosen: go/analysis (interface/struct type identity, same-package check)
#  [x] Severity and message are defined ("GID-236: service %q uses repository %q of another entity ...")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml (owned by the orchestrator)
