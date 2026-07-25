# language: en

Feature: GID-249 — a data store is reached through the repository layer (gidstorageplace)
  As a developer
  I want a storage driver (SQL, key-value, NoSQL) to be imported only by /dal/** and the composition root
  So that a store is always behind a repository and /client/** stays what it is — an external-API adapter

  # One analyzer, LoadModeSyntax is enough (the check is over import paths).
  # The layer is matched by path segments via internal/pathseg.
  # A driver is matched by import path prefix (path == prefix or prefix + "/").
  # Settings (all four are additive, the built-in lists stay intact):
  #   packages         — extra drivers (an in-house redis wrapper);
  #   allow            — extra allowed layers on top of dal + app;
  #   exclude-packages — imports that must not count as a driver;
  #   exclude-paths    — packages skipped entirely ("internal/legacy/cache").
  # Generated code (ast.IsGenerated) and _test.go are skipped.

  # --- Class 1: positive ---

  Scenario: positive — a key-value driver in /client
    Given a package in "/client/ratelimit" importing "github.com/redis/go-redis/v9"
    When the gidstorageplace analyzer checks the file
    Then the diagnostic "GID-249: package \"…/client/ratelimit\" reaches a data store directly — driver \"github.com/redis/go-redis/v9\" belongs to the repository layer. Fix: move the storage code to /dal/repository (its stored shapes to /dal/entity) and inject the repository through an interface; the pool itself is opened in /app. A /client package is for an external API (HTTP/gRPC/SMTP), not for a database" is reported

  Scenario: positive — a SQL driver in /domain/service
    Given a package in "/domain/service" importing "github.com/jackc/pgx/v5/pgxpool"
    When the gidstorageplace analyzer checks the file
    Then the diagnostic "GID-249: …" is reported

  Scenario: positive — an in-house wrapper added through settings.packages
    Given settings.packages contains "git.example.com/go-library/eredis" and a package in "/client/magiclink" imports it
    When the gidstorageplace analyzer checks the file
    Then the diagnostic "GID-249: …" is reported

  # --- Class 2: negative ---

  Scenario: negative — the driver in /dal/repository (where it belongs)
    Given a package in "/dal/repository" importing "github.com/redis/go-redis/v9"
    When the gidstorageplace analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — the driver in the composition root /app
    Given a package in "/app/api" importing "github.com/jackc/pgx/v5/pgxpool"
    When the gidstorageplace analyzer checks the file
    Then no diagnostic is reported
    # The pool is opened exactly there and injected into repositories.

  Scenario: negative — a client of an external API
    Given a package in "/client/mail" importing "github.com/wneessen/go-mail"
    When the gidstorageplace analyzer checks the file
    Then no diagnostic is reported
    # SMTP/HTTP/gRPC is what a client is for; an object store (minio) is not a driver either.

  # --- Class 3: boundary ---

  Scenario: boundary — the driver in /dal/entity (a column type)
    Given a package in "/dal/entity" importing "github.com/jackc/pgx/v5/pgtype"
    When the gidstorageplace analyzer checks the file
    Then no diagnostic is reported
    # The whole dal layer is allowed, not just repository: an entity carries pgtype columns.

  Scenario: boundary — a package path that merely starts like a driver
    Given a package in "/client/dedup" importing "github.com/redis/go-redis-extra-tools"
    When the gidstorageplace analyzer checks the file
    Then no diagnostic is reported
    # Prefix matching is by path segment: "github.com/redis/go-redis" + "/" or an exact match.

  Scenario: boundary — the metrics collector of a driver library
    Given a package in "/metric" importing "github.com/redis/go-redis/pkg/prometheus"
    When the gidstorageplace analyzer checks the file
    Then no diagnostic is reported
    # A "pkg/prometheus" (or pkg/logrus, pkg/logger, pkg/metrics, pkg/otel,
    # pkg/tracing) subpackage of a driver library carries no storage access —
    # it is wired in /metric and /app.

  Scenario: boundary — database/sql imported only for its NULL value types
    Given a package in "/domain/service/convert" using "sql.NullString" and nothing else from database/sql
    When the gidstorageplace analyzer checks the file
    Then no diagnostic is reported
    # An entity converter over sql.NullString is not storage access; only
    # package-level symbols count — the fields of NullString itself do not.

  Scenario: boundary — database/sql used for a connection
    Given a package in "/domain/usecase" declaring a field of type "*sql.DB"
    When the gidstorageplace analyzer checks the file
    Then the diagnostic "GID-249: …" is reported

  Scenario: boundary — a driver in a _test.go file outside dal
    Given a package in "/client/dedup" whose "cache_test.go" imports "github.com/redis/go-redis/v9"
    When the gidstorageplace analyzer checks the file
    Then no diagnostic is reported
    # An integration test or a miniredis fixture may talk to the driver from any layer.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — a layer added through settings.allow
    Given settings.allow contains "job" and a package in "/job/sweep" imports "github.com/redis/go-redis/v9"
    When the gidstorageplace analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — an import dropped through settings.exclude-packages
    Given settings.exclude-packages contains "github.com/redis/go-redis/pkg/instrumentation" and a package in "/client/tools" imports it
    When the gidstorageplace analyzer checks the file
    Then no diagnostic is reported
    # The escape hatch for a driver subpackage the built-in observability list does not know.

  Scenario: non-applicability — a package skipped through settings.exclude-paths
    Given settings.exclude-paths contains "legacy/cache" and that package imports "github.com/redis/go-redis/v9"
    When the gidstorageplace analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — generated code
    Given a generated file in "/client/gen" importing "github.com/jackc/pgx/v5"
    When the gidstorageplace analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-249)
#  [x] Layer chosen: go/analysis (package storageplace)
#  [x] Message is defined ("GID-249: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml
