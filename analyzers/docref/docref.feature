# language: en

Feature: GID-262 — a comment explains the code, not the development documentation
  As a developer
  I want comments to carry the constraint itself instead of a pointer into an ARD/PRD/backlog
  So that the reader of a function is not sent to a document that lives outside the repository and outlives no refactor

  Scenario: positive — a doc comment names the document
    Given the doc comment "AdCabinetResolver resolves ad cabinets through resource-registry (ARD Р-11)"
    When the analyzer checks the file
    Then a "GID-262" diagnostic is reported on the "ARD" marker with a hint to state the constraint itself

  Scenario: positive — a requirement id inside a comment
    Given the comment "collects unique cabinet ids before the call (@ФТ-11)"
    When the analyzer checks the file
    Then a "GID-262" diagnostic is reported on the "@ФТ-11" marker
    And its fix asks to move the id into the requirement map — a file of its own linked from the README, "ФТ-15 → TestCreate_DuplicateTitle_AlreadyExists"

  Scenario: boundary — the fix follows the class of the marker
    Given the comment "BACKLOG B-48: the guard rejects a write outside a transaction"
    When the analyzer checks the file
    Then the fix asks to state the constraint itself and leave the reference in the document — a backlog entry has no coverage map to preserve

  Scenario: positive — a task of the decomposition and a commit as the source of a decision
    Given the comments "один вызов на страницу (задача 29)" and "порядок шагов — коммит 34640e6"
    When the analyzer checks the file
    Then a "GID-262" diagnostic is reported on each of them

  Scenario: positive — a test doc comment is judged like any other
    Given the file svc_test.go with the comment "TestCabinets_Duplicates_ResolvedOnce — @ФТ-15: a repeated id costs one call"
    When the analyzer checks the file
    Then a "GID-262" diagnostic is reported

  Scenario: boundary — one diagnostic per comment, on the leftmost marker
    Given the comment "(BACKLOG B-48) — задача 43" carrying three markers
    When the analyzer checks the file
    Then exactly one diagnostic is reported, quoting "BACKLOG"

  Scenario: boundary — the marker list comes from settings
    Given settings.patterns holds "(?i)\bwiki\b" and settings.extra holds "\bUDMP-\d+\b"
    And the comments "starts the export (UDMP-3762)", "see the wiki page" and "(ARD Р-11, задача 29)"
    When the analyzer checks the file
    Then the first two are reported and the third one is not — the built-in list was replaced

  Scenario: boundary — an invalid regexp in settings
    Given settings.patterns holds the pattern "(unclosed"
    When the analyzer runs
    Then it fails with a compile error instead of silently dropping the marker

  Scenario: negative — the comment states the constraint itself
    Given the comment "reads cabinets in one batch call per page: a per-item resolve would cost N requests"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — text that only looks like a document reference
    Given the comments naming RFC3339, UTF-8, HTTP/2, TRACE_ENABLED and SkipVerify
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — a directive comment
    Given the comment "//nolint:giddocref // legitimate: the exception granted by ARD Р-11" and the comment "//go:generate mockery --name Registry"
    When the analyzer checks the file
    Then no diagnostic is reported — a tool directive is machine input, and the justification of a //nolint is the one place a decision may be named

  Scenario: non-applicability — generated code
    Given the file is marked "// Code generated ... DO NOT EDIT." and its comments carry "ARD Р-11", "@ФТ-15" and "ARD §12"
    When the analyzer checks the file
    Then no diagnostic is reported
