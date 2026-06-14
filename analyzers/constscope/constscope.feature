# language: en

Feature: GID-194 — constants are declared where they are used
  As a developer
  I want package-level constants outside the model/entity layers not to appear,
  but to be declared inside the function where they are directly used
  So that the constant's scope matches the scope of its usage,
  and shared constants live in one place — /domain/model or /dal/entity

  # Semantics (fixed by the requirement of 2026-06-07):
  # - a const inside a function is the target state, always fine;
  # - /domain/model/** and /dal/entity/** (including subpackages) are the home of
  #   shared constants, package-level is legal there (aligned with GID-123/167/169);
  # - an unexported package-level const used by ≥2 functions of the package
  #   stays package-level (similar to the GID-133 exception for helpers);
  # - usage outside a function body (package-level var/const/type, a signature,
  #   a test or generated file) makes the constant non-movable;
  # - an exported const outside model/entity is a violation: external usage
  #   is invisible to the analyzer, and shared constants live in model/entity;
  # - an iota block can only be moved as a whole — it is evaluated as a group;
  # - exclusions: //nolint:gidconstscope or settings.exclude (constant names).

  Scenario: positive — an exported constant in a service package
    Given the package path ends with the segments "domain/service"
    And the package-level constant "DefaultPageSize" is declared
    When the analyzer checks the package
    Then a "GID-194" diagnostic is reported on the constant "DefaultPageSize"

  Scenario: positive — a constant used by only one method
    Given the package path ends with the segments "domain/service"
    And the package-level constant "snapshotPrefix" is used only by the method "Snapshot.Render"
    When the analyzer checks the package
    Then a "GID-194" diagnostic is reported with a hint to declare it inside the function

  Scenario: negative — a constant shared by two methods of the package
    Given the package-level constant "snapshotTable" is used by the methods "Render" and "Table"
    When the analyzer checks the package
    Then no diagnostic is reported

  Scenario: negative — a constant declared inside a function
    Given the constant "endpoint" is declared in the body of the method "Endpoint"
    When the analyzer checks the package
    Then no diagnostic is reported

  Scenario: boundary — an iota group used entirely by one function
    Given the block "const (stateIdle = iota; stateBusy)" is declared
    And both constants are used only by the function "stateName"
    When the analyzer checks the package
    Then a single "GID-194" diagnostic is reported on the const block

  Scenario: boundary — an iota group used by different functions
    Given the block "const (colorRed = iota; colorBlue)" is declared
    And the constants are used by the functions "isRed" and "isBlue"
    When the analyzer checks the package
    Then no diagnostic is reported

  Scenario: boundary — an iota group with an exported constant
    Given the block "const (ModePrimary = iota; modeSecondary)" is declared
    When the analyzer checks the package
    Then a "GID-194" diagnostic is reported only about the export of "ModePrimary"
    And localizing the block is not suggested

  Scenario: boundary — usage in a package-level var
    Given the constant "defaultLabel" is used in the initialization of a package-level var
    When the analyzer checks the package
    Then no diagnostic is reported
    # the constant cannot be moved inside a function

  Scenario: boundary — usage in a function signature
    Given the constant "bufSize" defines an array length in a function parameter
    When the analyzer checks the package
    Then no diagnostic is reported
    # the signature is evaluated outside the function body

  Scenario: boundary — an unused constant
    Given the package-level constant "orphan" is not used anywhere
    When the analyzer checks the package
    Then no GID-194 diagnostic is reported
    # unused code is the domain of the unused linter

  Scenario: boundary — usage only from a test or generated file
    Given the constant is used only in "*_test.go" or "zz_generated.go"
    When the analyzer checks the package
    Then no diagnostic is reported

  Scenario: boundary — generated and test files are not checked
    Given a package-level constant is declared in "zz_generated.go" or "*_test.go"
    When the analyzer checks the package
    Then no diagnostic is reported

  Scenario: non-applicability — /domain/model and subpackages
    Given the package path ends with the segments "domain/model" or "domain/model/filter"
    And package-level constants are declared, including exported ones
    When the analyzer checks the package
    Then no GID-194 diagnostic is reported

  Scenario: non-applicability — /dal/entity
    Given the package path ends with the segments "dal/entity"
    And package-level constants are declared
    When the analyzer checks the package
    Then no GID-194 diagnostic is reported

  Scenario: non-applicability — a name in settings.exclude
    Given the linter setting "exclude: [LegacyExported]"
    And the package-level constant "LegacyExported" is declared
    When the analyzer checks the package
    Then no diagnostic is reported on "LegacyExported"

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (GID-194, no-global-const)
#  [x] Layer chosen: go/analysis (TypesInfo.Uses needed to count usages)
#  [x] Severity and message are defined ("GID-194: constant %q is used only in %q — ...")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
