# language: en

Feature: GID-250 — a test lives in the same package as the code under test (gidtestpackage)
  As a developer
  I want a _test.go file to declare "package <pkg>", not "package <pkg>_test"
  So that a unit test reaches the unexported code directly instead of going through the public API

  # One analyzer, the check is over the package clause of a _test.go file.
  # The team convention is the opposite of the standard golangci-lint
  # "testpackage" linter, which demands the external package — that linter is
  # in the disable list of the config for exactly this reason.
  # The package path of an external test package carries the same "_test"
  # suffix; settings.exclude-paths is written against the real directory.
  # Generated code (ast.IsGenerated) is skipped.

  # --- Class 1: positive ---

  Scenario: positive — an external test package
    Given the file "blackbox/svc_test.go" declares "package blackbox_test"
    When the gidtestpackage analyzer checks the file
    Then the diagnostic "GID-250: an external test package \"blackbox_test\" keeps the test away from the unexported code it tests. Fix: declare \"blackbox\" in this file — a test lives in the same package as the code under test" is reported

  # --- Class 2: negative ---

  Scenario: negative — the test is in the package under test
    Given the file "svc/svc_test.go" declares "package svc" and calls the unexported helper()
    When the gidtestpackage analyzer checks the file
    Then no diagnostic is reported

  # --- Class 3: boundary ---

  Scenario: boundary — a non-test file is never flagged
    Given the file "svc/svc.go" declares "package svc"
    When the gidtestpackage analyzer checks the file
    Then no diagnostic is reported
    # A package named "*_test" exists only as the external test package of a _test.go file.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — a directory exempted through settings.exclude-paths
    Given settings.exclude-paths contains "exempt/blackbox" and that directory keeps a black-box suite
    When the gidtestpackage analyzer checks the file
    Then no diagnostic is reported
    # A deliberate black-box suite over the public API stays possible.

  Scenario: non-applicability — generated code
    Given a generated _test.go file marked "// Code generated ... DO NOT EDIT."
    When the gidtestpackage analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-250)
#  [x] Layer chosen: go/analysis (package testpackage)
#  [x] Message is defined ("GID-250: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml
