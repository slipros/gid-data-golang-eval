# language: en

Feature: GID-159 — caching lives in a repository decorator (gidcacheplace)
  As a developer
  I want the cache to be a caching repository in /dal/repository wrapping the
  main one
  So that the domain layer keeps its business logic free of storage concerns
  and the cache can be swapped (in-memory LRU, redis, …) without touching it

  # One analyzer over the file's import declarations, no type info needed.
  # Scope: packages of the domain layer only, matched through
  # internal/pathseg.HasLayer — anchored to the module root, so a package
  # nested under another layer (svc/server/grpc/domain/…) is not the domain
  # layer and is left alone.
  # Trigger: an import of a known cache library. The default list holds
  # go-redis, go-redis (legacy), golang-lru, ristretto, bigcache, freecache,
  # go-cache and gomemcache; matching is by path prefix, so versioned suffixes
  # (/v2, /v9) are covered.
  # settings.packages (.golangci.yml) *replaces* the default list — it does not
  # extend it — which is how an in-house cache library is registered.
  # Generated code (ast.IsGenerated) is skipped.

  # --- Class 1: positive ---

  Scenario: positive — redis imported in a service
    Given "redis \"github.com/redis/go-redis/v9\"" imported in /domain/service
    When the gidcacheplace analyzer checks the file
    Then the diagnostic "GID-159: importing the cache library \"github.com/redis/go-redis/v9\" in the domain layer is forbidden. Fix: implement caching as a caching repository in /dal/repository that wraps the main one" is reported on the import

  Scenario: positive — an LRU cache imported in a usecase
    Given "lru \"github.com/hashicorp/golang-lru/v2\"" imported in /domain/usecase
    When the gidcacheplace analyzer checks the file
    Then the diagnostic "GID-159: importing the cache library \"github.com/hashicorp/golang-lru/v2\" in the domain layer is forbidden. …" is reported

  Scenario: positive — an in-house library listed in settings.packages
    Given settings "packages: [example.com/inhouse/cache]" and that import in /domain/service
    When the gidcacheplace analyzer checks the file
    Then the diagnostic "GID-159: importing the cache library \"example.com/inhouse/cache\" in the domain layer is forbidden. …" is reported

  # --- Class 2: negative ---

  Scenario: negative — the cache library imported by a repository
    Given "redis \"github.com/redis/go-redis/v9\"" imported in /dal/repository
    When the gidcacheplace analyzer checks the file
    Then no diagnostic is reported
    # That is exactly where the caching decorator belongs.

  Scenario: negative — the domain layer imports no cache library
    Given a /domain/service package importing only context and the model
    When the gidcacheplace analyzer checks the file
    Then no diagnostic is reported

  # --- Class 3: boundary ---

  Scenario: boundary — a "domain" directory nested under another layer
    Given "redis \"github.com/redis/go-redis/v9\"" imported in svc/server/grpc/domain
    When the gidcacheplace analyzer checks the file
    Then no diagnostic is reported
    # The layer is a path segment anchored to the module root, not a substring
    # found anywhere in the path.

  Scenario: boundary — settings.packages replaces the default list
    Given settings "packages: [example.com/inhouse/cache]" and "github.com/redis/go-redis/v9" imported in /domain/service
    When the gidcacheplace analyzer checks the file
    Then no diagnostic is reported
    # Replacement, not extension: the configured list is the whole list.

  Scenario: boundary — a versioned major of a listed library
    Given "github.com/redis/go-redis/v9" while the list holds "github.com/redis/go-redis"
    When the gidcacheplace analyzer checks the file
    Then the diagnostic "GID-159: …" is reported
    # Prefix matching covers every major version.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — a package outside the domain layer
    Given the cache library imported in /server/grpc/handler
    When the gidcacheplace analyzer checks the file
    Then no diagnostic is reported
    # The rule states where caching lives, and the handler is not the domain.

  Scenario: non-applicability — generated code
    Given a file carrying the "Code generated … DO NOT EDIT." header
    When the gidcacheplace analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-159)
#  [x] Layer chosen: go/analysis (package cacheplace)
#  [x] Message is defined ("GID-159: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
