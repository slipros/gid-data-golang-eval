# language: en

Feature: GID-217 — configurable ban of specific library symbols
  As a developer
  I want to ban specific symbols of third-party libraries (e.g. gdpostgres.TQuery)
  So that the team uses direct conn methods (Select, ScanRow, NamedStruct, Transaction), as agreed in repo.md

  Scenario: positive — call of a banned symbol
    Given the package gdpostgres "gitlab.gid.team/gid-data/tech/golang/libs/postgres.git" is imported
    And the code contains the call "gdpostgres.TQuery[int](conn, query)"
    When the analyzer checks the file
    Then a "GID-217" diagnostic is reported with a hint to use direct conn methods

  Scenario: positive — generic instantiation is caught the same way
    Given the package gdpostgres is imported
    And the code contains the instantiation "gdpostgres.TQuery[string](conn, query)"
    When the analyzer checks the file
    Then a "GID-217" diagnostic is reported on the TQuery selector

  Scenario: negative — allowed conn methods
    Given the package gdpostgres is imported
    And the code contains the calls "gdpostgres.Select(...)" and "gdpostgres.NamedStruct(...)"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — a same-named symbol from another package
    Given the package otherdb "example.com/otherdb" with its own TQuery function is imported
    And the code contains the call "otherdb.TQuery[int](query)"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — symbol configured in settings with a custom Msg
    Given settings.symbols defines the symbol otherdb.TQuery with Msg "otherdb.TQuery is banned by the project"
    And the code contains the call "otherdb.TQuery[int](query)"
    When the analyzer checks the file
    Then a "GID-217" diagnostic is reported with the text "otherdb.TQuery is banned by the project"

  Scenario: boundary — Pkg given as a path-segment suffix
    Given settings.symbols defines Pkg "libs/postgres.git" and Name "Select" without Msg
    And the symbol's import path ends with those segments
    When the analyzer checks the file
    Then a diagnostic is reported with the generic wording "symbol postgres.Select is banned by gidbansymbol"

  Scenario: non-applicability — package without an import of the banned library
    Given the file has no import of a banned package
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — generated code
    Given the file is marked "// Code generated ... DO NOT EDIT." and calls gdpostgres.TQuery
    When the analyzer checks the file
    Then no diagnostic is reported
