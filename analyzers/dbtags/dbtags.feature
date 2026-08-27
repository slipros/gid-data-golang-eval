# language: en

Feature: GID-125 — entity fields carry a DB column mapping tag
  As a developer
  I want every exported field of a DAL entity to name the column it maps to
  So that the entity-to-column mapping is written down and not guessed
  Sources: entity.md + requirement 2026-06-07; applicability gate 2026-08-27
  after the consent-webhook-trigger incident — a gRPC-backed /dal/repository
  was judged as a data layer and produced 28 findings on entity.Document,
  Profile and Webhook, structures that never reach SQL

  Linter: giddbtags. Judged: /dal/entity (final path segments), exported
  fields, non-generated files. Default tag: db; settings.tags replaces the
  list (the ClickHouse library uses ch).

  Applicability gate (internal/sqlstack): the rule only judges a module that
  actually speaks SQL — at least one of the module's OWN non-test files
  imports the SQL stack (database/sql, sqlx, pgx, lib/pq, the MySQL driver,
  squirrel, clickhouse-go, bun, gorm, ent; settings.sql-imports replaces the
  list). The verdict is per module and reads imports IN CODE, not go.mod — a
  dependency there can be transitive. A /dal says where a repository WOULD
  live, not that there is a database behind it.

  # --- Positive class: the violation is caught ---

  Scenario: an exported entity field without a mapping tag
    Given "/dal/entity" of a module importing sqlx declares "Account.Email" with no db tag
    When the analyzer checks the file
    Then a "GID-125" diagnostic is reported telling to add the tag

  Scenario: a field carrying another tag only
    Given an entity field carries `json:"created_at"` and no db tag
    When the analyzer checks the file
    Then a "GID-125" diagnostic is reported

  Scenario: the SQL import sits outside the dal
    Given the composition root of the module imports "database/sql" and "/dal/entity" declares an untagged field
    When the analyzer checks the file
    Then a "GID-125" diagnostic is reported — the verdict is per module, not per package

  # --- Negative class: clean code passes ---

  Scenario: every exported field carries the db tag
    Given "/dal/entity" declares "Job" with a db tag on each field
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: the ch tag of the ClickHouse library
    Given settings.tags is "['db', 'ch']" and the field carries `ch:"name"`
    When the analyzer checks the file
    Then no diagnostic is reported

  # --- Boundary class: the applicability gate ---

  Scenario: the module reaches the database through an in-house wrapper
    Given the repository imports "gid.team/libs/pgstore" and nothing of the default stack
    When the analyzer checks the file with the default settings
    Then no diagnostic is reported

  Scenario: the in-house wrapper is named in the settings
    Given settings.sql-imports is "['gid.team/libs/pgstore']" and the repository imports it
    When the analyzer checks the file
    Then a "GID-125" diagnostic is reported for the untagged field

  Scenario: the SQL stack is reached for by a _test.go file only
    Given the module's only sqlx import sits in "repository_test.go"
    When the analyzer checks the file
    Then no diagnostic is reported — a fixture may reach for the driver in a service that stores nothing

  Scenario: a package with no go.mod above it
    Given an analysistest fixture in GOPATH style, outside any module of its own
    When the analyzer checks the file
    Then the entity is judged — without a module root to read, the rule keeps its pre-gate behaviour

  # --- Non-applicability class ---

  Scenario: the module has no database at all
    Given "/dal/repository" of the module speaks gRPC and fills entities from protobuf
    When the analyzer checks "/dal/entity" of that module
    Then no diagnostic is reported — a tag would document a mapping that never reaches a column

  Scenario: a package outside /dal/entity
    Given "/domain/model" declares an untagged struct
    When the analyzer checks the file
    Then no diagnostic is reported (GID-168 judges the db tag there, in the other direction)

  Scenario: an unexported or embedded field
    Given an entity declares "cursor.offset" and an embedded type
    When the analyzer checks the file
    Then no diagnostic is reported — neither is mapped directly

  Scenario: a generated file
    Given a file with the "Code generated … DO NOT EDIT." marker declares an entity
    When the analyzer checks the file
    Then no diagnostic is reported
