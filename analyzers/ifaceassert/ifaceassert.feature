# language: en

Feature: GID-274 — no compile-time interface assertion the package already proves (gidifaceassert)
  As a developer
  I want the contract check to live in one place
  So that a change to an interface is carried by the wiring the compiler already
  verifies, and not by a hand-written line repeating it

  # One analyzer, type info needed. Two steps over the package:
  #   1. collect the package-level assertions `var _ Iface = value` — the blank
  #      name, an explicit interface type with at least one method, a value of a
  #      non-interface type;
  #   2. walk the package for a conversion of that very type to that very
  #      interface — an argument of a call, the right-hand side of an
  #      assignment, a declaration with an explicit type, a field of a composite
  #      literal, a returned value.
  # An assertion with a conversion behind it is reported and the diagnostic
  # names the file:line of the conversion; an assertion without one is left
  # alone — that is the library and the DI-container case, where the assertion
  # is the only compile-time check there is.
  # FP-safe by construction: a context the analyzer does not understand is a
  # conversion it does not see, which costs a diagnostic, never a false one.
  # Generated code (ast.IsGenerated) is not judged, and a _test.go file
  # (srcfile.IsTest) is out of the rule on both sides: it is not judged, and it
  # proves nothing about an assertion in production code either.

  # --- Class 1: positive ---

  Scenario: positive — the type is wired into a constructor of the same package
    Given "var _ service.DatasetSnapshotRepository = (*repository.DatasetSnapshot)(nil)"
    And "service.NewDatasetSnapshot(snapshotRepo)" is called in the package with snapshotRepo of type *repository.DatasetSnapshot
    When the gidifaceassert analyzer checks the package
    Then the diagnostic "GID-274: redundant compile-time assertion: the package already passes this value as service.DatasetSnapshotRepository at wiring.go:52, so the compiler checks the contract there. Fix: delete the \"var _ service.DatasetSnapshotRepository = (*repository.DatasetSnapshot)(nil)\" line" is reported on the assertion

  Scenario: positive — the conversion happens in a field of a composite literal
    Given "var _ service.SnapshotManifestRepository = (*repository.SnapshotManifest)(nil)"
    And "service.Registry{Manifests: manifestRepo}" is built in the package
    When the gidifaceassert analyzer checks the package
    Then the diagnostic "GID-274: …" is reported

  Scenario: positive — the conversion happens in an assignment
    Given "var _ service.TableSchemaRepository = (*repository.DatasetSnapshot)(nil)"
    And "schemaRepo = snapshotRepo" assigns to a variable of type service.TableSchemaRepository
    When the gidifaceassert analyzer checks the package
    Then the diagnostic "GID-274: …" is reported

  Scenario: positive — the assertion of a value type wired as a value
    Given "var _ usecase.LatestPageStore = latestPageStore{}"
    And "usecase.NewLatestPage(..., latestPageStore{}, ...)" is called in the package
    When the gidifaceassert analyzer checks the package
    Then the diagnostic "GID-274: …" is reported

  # --- Class 2: negative ---

  Scenario: negative — the package converts the type nowhere
    Given "var _ service.DatasetSnapshotRepository = (*Snapshot)(nil)" in a library package
    And the consumer of Snapshot lives outside the module
    When the gidifaceassert analyzer checks the package
    Then no diagnostic is reported

  Scenario: negative — the wiring goes through a DI container
    Given "var _ usecase.LatestPageStore = objectStore{}"
    And the package only hands the constructor to "c.Provide(newObjectStore)", a parameter of type any
    When the gidifaceassert analyzer checks the package
    Then no diagnostic is reported

  Scenario: negative — another type is wired under the same interface
    Given "var _ usecase.LatestPageCatalog = (*service.DatasetCatalog)(nil)"
    And the package converts only catalogValue to usecase.LatestPageCatalog
    When the gidifaceassert analyzer checks the package
    Then no diagnostic is reported

  # --- Class 3: boundary ---

  Scenario: boundary — asserted for the pointer, wired as a value
    Given "var _ usecase.LatestPageParquet = (*latestPageParquet)(nil)"
    And "usecase.NewLatestPage(..., latestPageParquet{})" passes the value type
    When the gidifaceassert analyzer checks the package
    Then the diagnostic "GID-274: …" is reported, because the method set of the pointer holds the method set of the value

  Scenario: boundary — asserted for the value, wired as a pointer
    Given "var _ usecase.LatestPageCatalog = catalogValue{}"
    And "usecase.UseCatalog(&catalogValue{})" passes the pointer
    When the gidifaceassert analyzer checks the package
    Then no diagnostic is reported, because the pointer's method set is the wider one

  Scenario: boundary — the conversion is to a wider interface
    Given "var _ SnapshotPort = (*Store)(nil)"
    And the package converts *Store to SnapshotAndSchemaPort, an interface holding SnapshotPort
    When the gidifaceassert analyzer checks the package
    Then no diagnostic is reported: the narrower contract does follow, but the diagnostic would name a contract the reader cannot find on the line it points at

  Scenario: boundary — the conversion lives only in a test file
    Given "var _ usecase.LatestPageCatalog = (*service.DatasetCatalog)(nil)" in wiring.go
    And the only conversion of *service.DatasetCatalog is in wiring_test.go
    When the gidifaceassert analyzer checks the package
    Then no diagnostic is reported: a conversion in a test file is checked by go test, not by go build

  Scenario: boundary — the assertion of the empty interface
    Given "var _ any = (*Snapshot)(nil)"
    When the gidifaceassert analyzer checks the package
    Then no diagnostic is reported: an interface without methods states no contract

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — an interface asserted against an interface
    Given "var _ usecase.LatestPageSnapshot = (SnapshotPort)(nil)"
    And the package converts SnapshotPort to usecase.LatestPageSnapshot
    When the gidifaceassert analyzer checks the package
    Then no diagnostic is reported: the wiring shape the rule is about starts from a concrete type

  Scenario: non-applicability — an assertion in a test file
    Given "var _ usecase.LatestPageStore = testStore{}" in wiring_test.go
    And "usecase.NewLatestPage(nil, nil, testStore{}, nil)" is called two lines below it
    When the gidifaceassert analyzer checks the package
    Then no diagnostic is reported: the double satisfies an interface it does not own, and the assertion is how the test states which one

  Scenario: non-applicability — an assertion of a production type in a test file
    Given "var _ usecase.LatestPageStore = latestPageStore{}" in wiring_test.go
    And the production code converts latestPageStore in wiring.go
    When the gidifaceassert analyzer checks the package
    Then no diagnostic is reported: the file the assertion lives in is not judged

  Scenario: non-applicability — generated code
    Given "var _ usecase.LatestPageStore = generatedStore{}" in a file with the "Code generated … DO NOT EDIT." header
    When the gidifaceassert analyzer checks the package
    Then no diagnostic is reported

  Scenario: non-applicability — a non-blank declaration
    Given "var store usecase.LatestPageStore = latestPageStore{}"
    When the gidifaceassert analyzer checks the package
    Then no diagnostic is reported: this is a variable, not an assertion

  Scenario: non-applicability — the interface is on settings.exclude
    Given "var _ Store = objectStore{}" and settings.exclude holding "Store"
    When the gidifaceassert analyzer checks the package
    Then no diagnostic is reported

  # --- Exceptions ---
  # //nolint:gidifaceassert — a pinpoint exception on the assertion line.
  # settings.exclude — "Interface" (any type) | "Type.Interface".
