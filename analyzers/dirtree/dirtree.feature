# language: en

Feature: GID-158 — a folder holds only the subfolders the tree allows (giddirtree)
  As a developer
  I want the service's folder tree fixed in the linter
  So that a new top-level folder is a deliberate decision rather than a place
  where code quietly accumulates outside the layers

  # One analyzer over the package path, no type info needed.
  # The tree is a map: a folder (segments separated by /) -> the subfolders
  # allowed directly inside it. A folder that is not a key is unrestricted.
  # The default tree is the canonical service structure (ARCHITECTURE.md):
  # internal -> app, client, dal, domain, event, job, metric, schedule, server;
  # internal/dal -> entity, repository; internal/dal/repository -> convert,
  # build; internal/domain -> model, service, usecase;
  # internal/domain/service -> convert; internal/server -> grpc, http.
  # A key is matched anywhere in the import path (pathseg.Index), so the module
  # prefix does not have to be spelled out.
  # settings.tree (.golangci.yml) *replaces* the default tree.
  # Keys are walked in sorted order, so the diagnostics of one package come out
  # deterministically; the message is reported on the package clause of every
  # non-generated file of the package.
  # GID-224 rides along: schedule is a transport leaf of the tree.

  # Scope: only a module laid out as a service — internal/modlayout walks up to
  # the package's go.mod and looks for a layer directory (domain, dal, server,
  # app, usecase, repository) at the module root or under internal/. A flat
  # library module has no layer to point at, so the rule stays silent there.

  # --- Class 1: positive ---

  Scenario: positive — an unknown folder in internal/
    Given the package "svc/internal/cache"
    When the giddirtree analyzer checks the package
    Then the diagnostic "GID-158: folder \"cache\" is not allowed in internal/ (allowed: app, client, dal, domain, event, job, metric, schedule, server); perhaps it should be a service or usecase; configure the tree via settings.tree" is reported on the package clause
    # internal/ carries an extra hint: the new folder is usually a service or a
    # usecase in disguise.

  Scenario: positive — a technology folder in internal/dal
    Given the package "svc/internal/dal/redis"
    When the giddirtree analyzer checks the package
    Then the diagnostic "GID-158: folder \"redis\" is not allowed in internal/dal/ (allowed: entity, repository); configure the tree via settings.tree" is reported

  Scenario: positive — a custom tree from settings
    Given settings "tree: {pkg: [api, contract], pkg/api: [v1, v2]}" and the packages "custom/pkg/util" and "custom/pkg/api/v3"
    When the giddirtree analyzer checks the packages
    Then "GID-158: folder \"util\" is not allowed in pkg/ (allowed: api, contract); …" and "GID-158: folder \"v3\" is not allowed in pkg/api/ (allowed: v1, v2); …" are reported

  # --- Class 2: negative ---

  Scenario: negative — the canonical structure
    Given the packages "svc/internal/domain/service" and "svc/internal/dal/repository/convert"
    When the giddirtree analyzer checks the packages
    Then no diagnostic is reported

  Scenario: negative — a folder that is not a key of the tree
    Given the package "svc/pkg/anything" with the default tree
    When the giddirtree analyzer checks the package
    Then no diagnostic is reported
    # Only folders listed as keys are restricted.

  # --- Class 3: boundary ---

  Scenario: boundary — a package deeper than a listed folder
    Given the package "svc/internal/domain/model/filter"
    When the giddirtree analyzer checks the package
    Then no diagnostic is reported
    # internal/domain/model is not a key, so what it contains is unrestricted;
    # its parent internal/domain allows "model" itself.

  Scenario: boundary — the key folder itself
    Given the package "svc/internal/dal"
    When the giddirtree analyzer checks the package
    Then no diagnostic is reported
    # There is no next segment to check.

  Scenario: boundary — settings.tree replaces the default tree
    Given settings "tree: {pkg: [api, contract]}" and the package "custom/internal/cache"
    When the giddirtree analyzer checks the package
    Then no diagnostic is reported
    # Replacement, not extension: with a custom tree the default keys are gone.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — a package outside every key
    Given the package "example.com/tools/gen"
    When the giddirtree analyzer checks the package
    Then no diagnostic is reported

  Scenario: non-applicability — generated code
    Given every file of the package carrying the "Code generated … DO NOT EDIT." header
    When the giddirtree analyzer checks the package
    Then no diagnostic is reported

  Scenario: non-applicability — a library module
    Given the package "gitlab.gid.team/.../libs/trino.git/internal/pool" in a module with no layer directories
    When the giddirtree analyzer checks the package
    Then no diagnostic is reported
    # The tree describes a service; a library lays its packages out freely.

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-158/GID-224)
#  [x] Layer chosen: go/analysis (package dirtree)
#  [x] Message is defined ("GID-158: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest (default and custom tree)
#  [x] Rule enabled in .golangci.yml
