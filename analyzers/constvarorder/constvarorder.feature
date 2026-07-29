# language: en

Feature: GID-130 — declaration order in a file: import, const, var, the rest (gidconstvarorder)
  As a developer
  I want const and var blocks at the top of the file
  So that the file's constants and package state are found in one known place
  instead of being scattered between types and functions

  # One analyzer over the file's top-level declarations, no type info needed.
  # Every declaration gets a rank: import (0), const (1), var (2), everything
  # else — types and functions (3). Walking the file, the rank must never
  # decrease; the declaration that drops below the maximum seen so far is the
  # one reported, so the misplaced block is named rather than everything after it.
  # Only const and var are reported: a type or a function can legitimately
  # follow anything.
  # Generated code (ast.IsGenerated) is skipped.

  # --- Class 1: positive ---

  Scenario: positive — a const block below a var block
    Given a file with "var First = 0" followed by "const AfterVar = 1"
    When the gidconstvarorder analyzer checks the file
    Then the diagnostic "GID-130: a const block must be at the top of the file, right after import and above var, types and functions. Fix: move it up" is reported on the const block

  Scenario: positive — a const block below a type declaration
    Given a file with "type Job struct{}" followed by "const AfterType = 2"
    When the gidconstvarorder analyzer checks the file
    Then the diagnostic "GID-130: a const block must be at the top of the file, …" is reported

  Scenario: positive — a var block below a function
    Given a file with "func Do()" followed by "var AfterFunc = 3"
    When the gidconstvarorder analyzer checks the file
    Then the diagnostic "GID-130: a var block must be at the top of the file, after const and above types and functions. Fix: move it up" is reported

  # --- Class 2: negative ---

  Scenario: negative — the canonical order
    Given a file with "import", then a const block, then a var block, then types and functions
    When the gidconstvarorder analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — const and var are absent
    Given a file holding only types and functions
    When the gidconstvarorder analyzer checks the file
    Then no diagnostic is reported

  # --- Class 3: boundary ---

  Scenario: boundary — several const blocks in a row, then several var blocks
    Given a file with two const blocks followed by two var blocks
    When the gidconstvarorder analyzer checks the file
    Then no diagnostic is reported
    # The rank must not decrease; an equal rank repeated is fine.

  Scenario: boundary — a var block above a const block
    Given a file with "var A = 1" followed by "const B = 2"
    When the gidconstvarorder analyzer checks the file
    Then only the const block is reported
    # The var block is where it belongs relative to import; the const is the
    # declaration that broke the order.

  Scenario: boundary — a const declared inside a function
    Given "const timeout = time.Second" declared in the body of a function
    When the gidconstvarorder analyzer checks the file
    Then no diagnostic is reported
    # Only top-level declarations are ranked; a function-scoped const is exactly
    # what GID-194 asks for.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — generated code
    Given a file carrying the "Code generated … DO NOT EDIT." header
    When the gidconstvarorder analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — a file holding nothing but the import block
    Given a file with a package clause and imports only
    When the gidconstvarorder analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-130)
#  [x] Layer chosen: go/analysis (package constvarorder)
#  [x] Message is defined ("GID-130: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
