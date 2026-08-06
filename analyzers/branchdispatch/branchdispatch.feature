# language: en

Feature: GID-261 — a service method does not dispatch between two queries (gidbranchdispatch)
  As a developer
  I want one service method to be one operation
  So that a query is not hidden behind a field that is meaningful only half
  the time

  # Scope: the root of /domain/service (pathseg.EndsWith), LoadModeTypesInfo.
  # Trigger: an if/else — or an else-if chain, judged as a whole and reported
  # once on the outermost if — where every branch is a SINGLE call on the SAME
  # receiver (an assignment of the call, or a bare return of it) and at least
  # two distinct method names appear among the branches.
  # The receiver is compared as a written expression (types.ExprString), so
  # i.repo matches i.repo only.
  # Generated code (ast.IsGenerated) and _test.go files are skipped.

  # --- Class 1: positive ---

  Scenario: positive — the incident shape, assignment form
    Given "if org == \"\" { e, err = i.repo.Integration(ctx, id) } else { e, err = i.repo.IntegrationByOrganization(ctx, id, org) }"
    When the gidbranchdispatch analyzer checks the file
    Then the diagnostic "GID-261: method Integration.Get picks between i.repo.Integration and i.repo.IntegrationByOrganization by a condition on its input — one method, several operations. Fix: split it into a method per query (Integration and IntegrationByOrganization), each taking the arguments its own query needs, and let the caller choose" is reported on the if
    # resource-registry, modules firebase and yandex_audience (coreIntegration).

  Scenario: positive — return form
    Given "if org == \"\" { return i.repo.Integration(ctx, id) } else { return i.repo.IntegrationByOrganization(ctx, id, org) }"
    When the gidbranchdispatch analyzer checks the file
    Then the diagnostic is reported

  Scenario: positive — an else-if chain
    Given three branches calling i.repo.Integration, i.repo.IntegrationByOrganization and i.repo.Integration
    When the gidbranchdispatch analyzer checks the file
    Then one diagnostic is reported, on the outermost if
    # Three ways into one dependency is the same defect as two, only bigger.

  # --- Class 2: negative ---

  Scenario: negative — the same method with different arguments
    Given "if id == \"\" { e, err = i.repo.Integration(ctx, fallback) } else { e, err = i.repo.Integration(ctx, id) }"
    When the gidbranchdispatch analyzer checks the file
    Then no diagnostic is reported
    # One operation with a prepared argument, not a dispatch.

  Scenario: negative — different receivers
    Given "if core { e, err = i.coreRepo.Integration(ctx, id) } else { e, err = i.repo.Integration(ctx, id) }"
    When the gidbranchdispatch analyzer checks the file
    Then no diagnostic is reported
    # Two dependencies, not two ways into one (that pair is GID-260's business).

  # --- Class 3: boundary ---

  Scenario: boundary — a branch doing more than the call
    Given an else branch that calls i.repo.IntegrationsByFilter, checks its error and picks an element
    When the gidbranchdispatch analyzer checks the file
    Then no diagnostic is reported
    # A branch that is not a bare call is not a dispatch arm.

  Scenario: boundary — a guard with no else
    Given "if org == \"\" { return i.repo.Integration(ctx, id) }" followed by a plain return
    When the gidbranchdispatch analyzer checks the file
    Then no diagnostic is reported

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — a usecase
    Given the same method in /domain/usecase
    When the gidbranchdispatch analyzer checks the file
    Then no diagnostic is reported
    # The rule scope is /domain/service (owner decision 2026-08-06).

  Scenario: non-applicability — a _test.go double
    Given a test double dispatching over its own state
    When the gidbranchdispatch analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — settings.exclude
    Given settings.exclude ["Integration.Get", "Legacy"]
    When the gidbranchdispatch analyzer checks the file
    Then those two methods are silent, and "Read" is still reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-261)
#  [x] Layer chosen: go/analysis (package branchdispatch)
#  [x] Message is defined ("GID-261: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
