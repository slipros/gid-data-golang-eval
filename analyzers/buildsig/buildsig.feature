# language: en

Feature: GID-212 — contract of repository build functions
  As a developer
  I want repository build functions to share a single result contract
  And squirrel to live only in build packages
  So that query construction is predictable and isolated
  Source: repo.md
  Scope:
    - signature check: exported functions without a receiver
      in packages /dal/repository/build/**;
    - squirrel ban: importing github.com/Masterminds/squirrel in any package
      outside /dal/repository/build/**.
  Result-signature contract:
    - (string, []any, error) — a single query (sql, args, err); OR
    - (*<...>.Batch, error) — a batch operation (matched by the type name Batch,
      any package).
  Generated code is skipped.

  # --- Positive class: the violation is caught ---

  Scenario: a build function returns (string, error) — violation
    Given "func BuildBad(...) (string, error)" is declared in /dal/repository/build
    When the analyzer checks the file
    Then a "GID-212" diagnostic is reported with the text "a build function must return (sql string, args []any, err error) or (*batch.Batch, error)"

  Scenario: a build function returns *squirrel.SelectBuilder — violation
    Given "func BuildBuilder() *squirrel.SelectBuilder" is declared in /dal/repository/build
    When the analyzer checks the file
    Then a "GID-212" diagnostic is reported with the text "a build function must return ..."

  Scenario: squirrel import in /dal/repository — violation
    Given "github.com/Masterminds/squirrel" is imported in /dal/repository
    When the analyzer checks the file
    Then a "GID-212" diagnostic is reported with the text "squirrel is allowed only in repository build packages (/dal/repository/build)"

  Scenario: squirrel import in /domain/service — violation
    Given "github.com/Masterminds/squirrel" is imported in /domain/service
    When the analyzer checks the file
    Then a "GID-212" diagnostic is reported with the text "squirrel is allowed only in repository build packages ..."

  # --- Negative class: correct code passes ---

  Scenario: a build function returns (string, []any, error) — ok
    Given "func SelectJobs(...) (string, []any, error)" is declared in /dal/repository/build
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a build function returns (*batch.Batch, error) — ok
    Given "func InsertJobsBatch(...) (*batch.Batch, error)" is declared in /dal/repository/build
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: squirrel import in /dal/repository/build — ok
    Given "github.com/Masterminds/squirrel" is imported in /dal/repository/build
    When the analyzer checks the file
    Then no diagnostic is reported

  # --- Boundary class ---

  Scenario: an unexported helper in a build package with a different signature — not flagged
    Given "func helper(n int) int" is declared in /dal/repository/build
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a build function with no results — violation
    Given "func BuildVoid()" is declared in /dal/repository/build
    When the analyzer checks the file
    Then a "GID-212" diagnostic is reported with the text "a build function must return ..."

  Scenario: a generated file — skipped
    Given the file is marked "// Code generated ... DO NOT EDIT."
    When the analyzer checks the file
    Then no diagnostic is reported

  # --- Non-applicability class: the signature check does not apply outside build ---

  Scenario: a function with an arbitrary signature in /dal/repository — not flagged
    Given "func DoStuff(id string) (int, error)" is declared in /dal/repository
    When the analyzer checks the file
    Then no result diagnostic is reported

  Scenario: a function with an arbitrary signature in /domain/service — not flagged
    Given "func Process(id string) (bool, error)" is declared in /domain/service
    When the analyzer checks the file
    Then no result diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-212) — outside the scope of this change
#  [x] Layer chosen: go/analysis (go/types, LoadModeTypesInfo)
#  [x] Severity and message are defined ("GID-212: ...")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml (outside the scope of this task)
