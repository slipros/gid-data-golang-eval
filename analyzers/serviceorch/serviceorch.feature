# language: en

Feature: GID-260 — a service does not orchestrate (gidserviceorch)
  As a developer
  I want a domain service to be one entity and its repository
  So that composing several sources stays in a usecase, where the reader
  expects to find it

  # Scope: the root of /domain/service (pathseg.EndsWith), LoadModeTypesInfo.
  # Two markers, both read off the service struct (except *Options):
  #  1. more than one repository field — a named type (through a pointer too)
  #     whose name ends with a settings.suffixes entry (default "Repository");
  #     reported once, on the type name;
  #  2. a transaction field — a func value returning exactly error whose last
  #     parameter is itself a func returning exactly error
  #     (model.InTransactionFunc, GID-175); matched by signature, reported on
  #     the field.
  # Generated code (ast.IsGenerated) and _test.go files are skipped.

  # --- Class 1: positive ---

  Scenario: positive — two repositories in one service
    Given a struct "Integration" in /domain/service with fields "coreRepo CoreSnapshotRepository" and "repo SnapshotRepository"
    When the gidserviceorch analyzer checks the file
    Then the diagnostic "GID-260: service \"Integration\" depends on 2 repositories (CoreSnapshotRepository, SnapshotRepository) — a service is one entity and its repository. Fix: split it into a service per entity and compose them in a usecase (or //nolint:gidserviceorch when explicitly intended)" is reported on the type name

  Scenario: positive — a transaction held by a service
    Given a struct "Writer" in /domain/service with the field "tx model.InTransactionFunc"
    When the gidserviceorch analyzer checks the file
    Then the diagnostic "GID-260: service \"Writer\" holds a transaction — a service does not coordinate several writes. Fix: keep the transaction in a usecase, which calls the services it composes (or //nolint:gidserviceorch when explicitly intended)" is reported on the field

  Scenario: positive — the incident shape, both markers at once
    Given a struct with "tx model.InTransactionFunc", "coreRepo CoreSnapshotRepository" and "repo SnapshotRepository"
    When the gidserviceorch analyzer checks the file
    Then both diagnostics are reported — one on the type name, one on the tx field
    # resource-registry pkg/integration/push/firebase/domain/service/integration.go,
    # where the //nolint:gidserviceentity read "orchestration IS the essence of
    # this service".

  # --- Class 2: negative ---

  Scenario: negative — one repository is the norm
    Given a struct "Snapshot" with "repo SnapshotRepository", "validator SnapshotValidator" and "opts Options"
    When the gidserviceorch analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — an *Options type is not a service
    Given a struct "SnapshotOptions" holding two repositories
    When the gidserviceorch analyzer checks the file
    Then no diagnostic is reported

  # --- Class 3: boundary ---

  Scenario: boundary — a plain callback is not a transaction runner
    Given the field "notify func(ctx context.Context, id string) error"
    When the gidserviceorch analyzer checks the file
    Then no diagnostic is reported
    # The last parameter is not a func returning error.

  Scenario: boundary — a func returning more than error is not a transaction runner
    Given the field "load func(ctx context.Context, fn func(ctx context.Context) error) (string, error)"
    When the gidserviceorch analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — a custom repository suffix
    Given settings.suffixes ["Repository", "Store"] and a struct with "repo SnapshotRepository" and "store CoreSnapshotStore"
    When the gidserviceorch analyzer checks the file
    Then the two-repository diagnostic is reported

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — a usecase orchestrates by design
    Given the same struct in /domain/usecase
    When the gidserviceorch analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — a _test.go harness
    Given a "_test.go" fixture struct holding tx and two repositories
    When the gidserviceorch analyzer checks the file
    Then no diagnostic is reported
    # A harness bundling several doubles is composition of the test (GID-250).

  Scenario: non-applicability — settings.exclude
    Given settings.exclude ["LegacyWriter", "Mixed.tx"]
    When the gidserviceorch analyzer checks the file
    Then "LegacyWriter" is silent entirely, and "Mixed" keeps only its two-repository diagnostic
    # The excluded field drops out of the check; the rest of the struct is
    # judged as usual — the setting drives the exclusion, it is not a blanket
    # exemption for the shape.

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-260)
#  [x] Layer chosen: go/analysis (package serviceorch)
#  [x] Message is defined ("GID-260: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
