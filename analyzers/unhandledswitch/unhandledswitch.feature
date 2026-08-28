# language: en

Feature: GID-274 — enum/discriminator switches return an unhandled-value error
  As a developer
  I want every switch over a finite enum or discriminator to state its unknown-value behavior
  So that a newly added or invalid value does not disappear silently

  # The analyzer uses type information, not identifier names: a named string
  # type with a package-level constant of that same type is finite.
  # A default is compliant only when its sole statement is the direct return:
  # errors.WithStack(gderror.NewUnhandledValueError(the switch tag expression)).
  # The calls must be static package-qualified functions with those symbol
  # names, and the constructor must receive the exact switch tag value.

  # --- Class 1: positive (the violation is caught) ---

  Scenario: enum switch without default
    Given a function switches on a named enum with typed constants
    And the switch has one or more value cases but no default clause
    When the gidunhandledswitch analyzer checks the file
    Then a "GID-274" diagnostic is reported on the switch

  Scenario: discriminator field switch without default
    Given a function switches on a named discriminator field such as "event.Kind"
    And the switch has no default clause
    When the gidunhandledswitch analyzer checks the file
    Then a "GID-274" diagnostic is reported on the switch

  Scenario: default returns a plain fallback
    Given a named enum switch has a default clause that returns nil
    When the gidunhandledswitch analyzer checks the file
    Then a "GID-274" diagnostic is reported on the switch

  Scenario: default returns the bare unhandled-value error
    Given a named enum switch has "default: return gderror.NewUnhandledValueError(value)"
    When the gidunhandledswitch analyzer checks the file
    Then a "GID-274" diagnostic is reported on the switch

  Scenario: default wraps an unhandled error for a different value
    Given a named enum switch has "default: return errors.WithStack(gderror.NewUnhandledValueError(other))"
    When the gidunhandledswitch analyzer checks the file
    Then a "GID-274" diagnostic is reported on the switch

  Scenario: default uses another error handler
    Given a named enum switch has "default: return errors.New(\"unknown\")"
    When the gidunhandledswitch analyzer checks the file
    Then a "GID-274" diagnostic is reported on the switch

  Scenario: default performs work before returning the required error
    Given a named enum switch has an extra statement before the unhandled-value return
    When the gidunhandledswitch analyzer checks the file
    Then a "GID-274" diagnostic is reported on the switch

  Scenario: default calls the wrapper through a function value
    Given a named enum switch calls a local variable containing errors.WithStack
    When the gidunhandledswitch analyzer checks the file
    Then a "GID-274" diagnostic is reported on the switch

  # --- Class 2: negative (clean code passes) ---

  Scenario: enum switch with the exact unhandled-value return
    Given a named enum switch has "default: return errors.WithStack(gderror.NewUnhandledValueError(value))"
    When the gidunhandledswitch analyzer checks the file
    Then no diagnostic is reported

  Scenario: discriminator field switch passes its exact field value
    Given a named discriminator switch has "default: return errors.WithStack(gderror.NewUnhandledValueError(event.Kind))"
    When the gidunhandledswitch analyzer checks the file
    Then no diagnostic is reported

  # --- Class 3: boundary (looks similar but is outside the deterministic rule) ---

  Scenario: plain string switch
    Given a function switches on a value of the predeclared type "string"
    When the gidunhandledswitch analyzer checks the file
    Then no diagnostic is reported
    # The possible values of a plain string are not a finite declared set.

  Scenario: plain integer switch
    Given a function switches on a value of the predeclared type "int"
    When the gidunhandledswitch analyzer checks the file
    Then no diagnostic is reported

  Scenario: named integer enum
    Given a function switches on a named integer type with typed constants
    When the gidunhandledswitch analyzer checks the file
    Then no diagnostic is reported
    # This project defines enums as named string types (GID-123); integer enums are out of scope.

  Scenario: named string without typed constants
    Given a function switches on a named string type with no package-level constants
    When the gidunhandledswitch analyzer checks the file
    Then no diagnostic is reported

  Scenario: an expressionless switch
    Given a function uses "switch { case condition: ...}" without a tag expression
    When the gidunhandledswitch analyzer checks the file
    Then no diagnostic is reported

  # --- Class 4: non-applicability ---

  Scenario: generated code
    Given a generated file contains an enum switch without a default
    When the gidunhandledswitch analyzer checks the file
    Then no diagnostic is reported

  Scenario: test code
    Given a _test.go file contains an enum switch without a default
    When the gidunhandledswitch analyzer checks the file
    Then no diagnostic is reported

# --- Checklist for adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md)
#  [x] Layer chosen: go/analysis (type info identifies declared enum values)
#  [x] Severity and message are defined
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
