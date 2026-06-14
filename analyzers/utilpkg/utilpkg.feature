# language: en
# Specification of rule GID-187 (no-util-package).
# Linter: gidutilpkg (go/analysis, LoadModeSyntax — no types needed).

Feature: GID-187 — ban on junk-drawer packages (util, utils, common, …)
  As a backend-go service developer
  I want packages to be named after what they provide
  So that the package name conveys its purpose rather than "everything goes here"

  # The package name is checked (pass.Pkg.Name() == the last path segment)
  # case-insensitively. The default blacklist:
  #   util, utils, common, helper, helpers, shared, misc, lib, base.
  # settings.names fully replaces the default.
  # The _test suffix of a test package is normalized (utils_test → utils).
  # One report per package — on the package clause of the first non-generated file;
  # generated files are skipped when choosing the position.
  # The comparison is by the full name: stringutil does NOT match util (it is not a suffix match).

  # --- Positive cases (the violation is caught) ---

  Scenario: positive — the util package
    Given a package named "util"
    When the analyzer checks the package
    Then the diagnostic "GID-187: package \"util\" is a junk drawer with no responsibility. Fix: name the package after what it provides" is reported on the package clause

  Scenario: positive — the helpers package
    Given a package named "helpers"
    When the analyzer checks the package
    Then the diagnostic "GID-187: package \"helpers\" …" is reported on the package clause

  Scenario: positive — the common package
    Given a package named "common"
    When the analyzer checks the package
    Then the diagnostic "GID-187: package \"common\" …" is reported on the package clause

  # --- Negative cases (clean code passes) ---

  Scenario: negative — the meaningful name convert
    Given a package named "convert"
    When the analyzer checks the package
    Then no diagnostic is reported

  Scenario: negative — stringutil (a name with the util suffix is not matched)
    Given a package named "stringutil"
    When the analyzer checks the package
    Then no diagnostic is reported

  # --- Boundary cases (configuration/case) ---

  Scenario: boundary — settings.names=["junk"] matches junk
    Given settings.names = ["junk"] and a package named "junk"
    When the analyzer checks the package
    Then the diagnostic "GID-187: package \"junk\" …" is reported on the package clause

  Scenario: boundary — settings.names=["junk"] does NOT match util
    Given settings.names = ["junk"] and a package named "util"
    When the analyzer checks the package
    Then no diagnostic is reported
    # (The custom list fully replaces the default.)

  Scenario: boundary — the name case does not matter (UTIL → util)
    Given a package with an uppercase name normalizing to "util"
    When the analyzer checks the package
    Then the diagnostic "GID-187: …" is reported
    # (The comparison is case-insensitive; in real Go the package name is written in code
    #  as is, so it is covered by strings.ToLower normalization.)

  # --- Non-applicability (the rule does not apply) ---

  Scenario: non-applicability — an ordinary domain package model
    Given a package named "model"
    When the analyzer checks the package
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-187)
#  [x] Layer chosen: go/analysis, LoadModeSyntax (no types needed — the package name suffices)
#  [x] Severity and message are defined ("GID-187: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest (+ custom settings.names)
#  [ ] Rule enabled in .golangci.yml
