# language: en

Feature: GID-268 — a constructor does not take one dependency twice under different interfaces (giddupdep)
  As a developer
  I want a constructor to name every collaborator once
  So that the dependency list says how many entities the module really has,
  and two interfaces satisfied by one entity are merged instead of duplicated

  # One analyzer over the package's call expressions, type info needed: the
  # parameter types decide, and the arguments are matched by the objects behind
  # them (a variable, a field, a no-argument getter), not by source text.
  # Trigger: a call to a constructor (the New/new prefix by GID-104) where two
  # fixed parameters get the same value and their types are two DIFFERENT named
  # interfaces declared in ONE package of the caller's module.
  # The fix belongs to the interfaces, not to the call: merge them into one
  # interface carrying both method sets and take the dependency once.
  # Generated code (ast.IsGenerated) is skipped.

  # --- Class 1: positive ---

  Scenario: positive — the same getter twice
    Given "marketplace.NewModule(lab.SavedDatasetService(), lab.ShowcaseService(), lab.ShowcaseService())"
    And NewModule takes ShowcasesWithAttributeDataTypesService and FilterShowcasesByAccessService, both declared in marketplace
    When the giddupdep analyzer checks the file
    Then the diagnostic "GID-268: constructor NewModule receives the same value in parameters #2 showcases and #3 filter — one dependency passed twice under different interfaces. Fix: merge ShowcasesWithAttributeDataTypesService and FilterShowcasesByAccessService into a single interface and take the dependency as one parameter" is reported on the duplicated argument

  Scenario: positive — the same variable twice
    Given "showcases := lab.ShowcaseService()" and "marketplace.NewModule(saved, showcases, showcases)"
    When the giddupdep analyzer checks the file
    Then the diagnostic "GID-268: …" is reported

  Scenario: positive — the same struct field twice
    Given "marketplace.NewModule(d.saved, d.showcase, d.showcase)"
    When the giddupdep analyzer checks the file
    Then the diagnostic "GID-268: …" is reported

  # --- Class 2: negative ---

  Scenario: negative — every interface gets its own entity
    Given "marketplace.NewModule(lab.SavedDatasetService(), lab.ShowcaseService(), lab.FilterService())"
    When the giddupdep analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — the same field name on two different structs
    Given "marketplace.NewModule(left.saved, left.showcase, right.showcase)"
    When the giddupdep analyzer checks the file
    Then no diagnostic is reported
    # Matching is by object, not by source text: left.showcase and
    # right.showcase are two values.

  # --- Class 3: boundary ---

  Scenario: boundary — both parameters are one and the same interface
    Given "newPair(store, store)" where both parameters are Fetcher
    When the giddupdep analyzer checks the file
    Then no diagnostic is reported
    # There is nothing to merge: the duplicate is a wiring question (a primary
    # and a backup may legitimately be the same value), not an interface one.

  Scenario: boundary — the interfaces belong to another module
    Given "newPipe(store, store)" with parameters io.Writer and io.Reader
    When the giddupdep analyzer checks the file
    Then no diagnostic is reported
    # io.Copy(rw, rw) passes one value under two stdlib interfaces on purpose,
    # and neither interface is ours to merge.

  Scenario: boundary — the interfaces are declared in two packages
    Given "newMixed(store, store)" with parameters local.Fetcher and remote.Auditor
    When the giddupdep analyzer checks the file
    Then no diagnostic is reported
    # Merging would have to pick one of the two packages as the home of the
    # merged interface — a design decision, not a mechanical fix.

  Scenario: boundary — a method-less interface
    Given "newBlank(store, store)" with two parameters of type Anything (interface{})
    When the giddupdep analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — a getter taking an argument
    Given "marketplace.NewModule(saved, lab.ShowcaseFor(\"a\"), lab.ShowcaseFor(\"a\"))"
    When the giddupdep analyzer checks the file
    Then no diagnostic is reported
    # A call taking arguments may return a fresh value every time; only a
    # no-argument getter is treated as naming one entity.

  Scenario: boundary — nil dependencies
    Given "marketplace.NewModule(nil, nil, nil)"
    When the giddupdep analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — the duplicate lands in the variadic tail
    Given "newFeed(store, other{}, store, store)" where the tail is "...Listener"
    When the giddupdep analyzer checks the file
    Then no diagnostic is reported
    # The tail holds no parameter of its own to name in a fix.

  Scenario: boundary — the arguments come from one multi-value call
    Given "newVault(both(store))" where both() returns (Fetcher, Storer)
    When the giddupdep analyzer checks the file
    Then no diagnostic is reported

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — the callee is not a constructor
    Given "marketplace.Register(lab.ShowcaseService(), lab.ShowcaseService())"
    When the giddupdep analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — a test wiring one double into every parameter
    Given "marketplace.NewModule(double, double, double)" in a _test.go file
    When the giddupdep analyzer checks the file
    Then no diagnostic is reported
    # A double is written to satisfy every interface of the constructor
    # (GID-250 keeps it in the same package); the fix, if any, belongs to the
    # production wiring the rule already sees.

  Scenario: non-applicability — generated wiring
    Given a file carrying the "Code generated … DO NOT EDIT." header
    When the giddupdep analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-268)
#  [x] Layer chosen: go/analysis (package dupdep)
#  [x] Message is defined ("GID-268: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml, gid-golangci.yml and gid-golangci-rules.yml
