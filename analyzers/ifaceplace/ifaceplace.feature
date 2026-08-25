# language: en

Feature: GID-134 — interfaces live where they are used
  As a developer
  I want a dependency interface to be declared next to its consumer
  So that packages do not pull in foreign abstractions and stay loosely coupled
  Sources: styleguide.md#interfaces + requirement of 2026-06-07

  Linter: gidifaceplace. LoadMode: TypesInfo (types.Interface and the package
  of the interface declaration via Named.Obj().Pkg() are needed).

  Scope: struct fields and parameters/results of functions and methods of any package.
  Only a NAMED interface type in these positions is checked. We look at the package
  of the interface declaration and decide:
    - the same package — OK;
    - stdlib or an external library — OK; a service's "own" package is told apart
      from a library by path segments (pathseg): the path contains a layer segment
      (dal, domain, client, server, event, app, metric) — it is our package,
      otherwise a library;
    - an interface from the model layer (/domain/model, including subpackages) — OK, but
      only if the consumer is in /domain/service or /domain/usecase;
    - any other "own" package — violation.

  For GID-134, anonymous interfaces remain outside this package-placement check.
  GID-269 separately rejects non-empty anonymous interfaces used directly as
  struct field types. error, any/interface{}, generic constraints, and generated
  code (ast.IsGenerated) are not affected.

  Test files (srcfile.IsTest): a _test.go helper's parameters/results are a
  *use* of the interface, forced by the production constructor it wires up
  (handler.NewCreate(v, s CreateService)) — skipped. A struct type a test
  *declares* on its own picks its field types freely, not forced by anything —
  it stays in scope even in a _test.go file (incident 2026-08-06,
  resource-registry: yandex_audience_wire_test.go typed a wiring helper's
  parameters with the handler's own consumer-side interfaces because
  NewCreate demanded exactly that type).

  # --- Positive class: the violation is caught ---

  Scenario: service uses an interface from a foreign server package (field)
    Given a struct field in /domain/service has the type "grpc.Notifier" from /server/grpc
    When the analyzer checks the file
    Then a "GID-134" diagnostic is reported with the interface name and the declaration package

  Scenario: service uses an interface from a foreign server package (parameter)
    Given a method in /domain/service takes "grpc.Notifier"
    When the analyzer checks the file
    Then a "GID-134" diagnostic is reported

  Scenario: service uses an interface from a foreign server package (result)
    Given a method in /domain/service returns "grpc.Notifier"
    When the analyzer checks the file
    Then a "GID-134" diagnostic is reported

  # --- Negative class: clean code passes ---

  Scenario: the interface is declared in the same package — ok
    Given the interface "LocalRepository" from the same package is used in /domain/service
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: an interface from /domain/model with a service consumer — ok
    Given "model.JobRepository" is used in /domain/service
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: the library io.Reader (stdlib) — ok
    Given a method in /domain/service takes "io.Reader"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: an interface of an external library (no layer segments) — ok
    Given a method in /domain/service takes "extlib.Encoder" from example.com/extlib
    When the analyzer checks the file
    Then no diagnostic is reported

  # --- Boundary class ---

  Scenario: a model interface in /dal/repository — violation
    Given "model.JobRepository" is used in /dal/repository
    When the analyzer checks the file
    Then a "GID-134" diagnostic is reported (the model exception is only for service/usecase)

  Scenario: a model interface in /domain/usecase — ok
    Given "model.JobRepository" is used in /domain/usecase
    When the analyzer checks the file
    Then no diagnostic is reported

  # --- Non-applicability class ---

  Scenario: error in a result — not an interface with a package, skipped
    Given a method returns "error"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: an anonymous interface{ Foo() } in a parameter — outside both rules
    Given a method takes "interface{ Foo() }"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: any / interface{} in a parameter — skipped
    Given a method takes "any"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-interface types (struct, string) — skipped
    Given a method takes "model.Job" and "string"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a _test.go wiring helper's parameter is forced by a production constructor — skipped
    Given a _test.go helper function in /app/api takes "grpc.Notifier" from /server/grpc,
      mirroring the type handler.NewCreate(validator, service CreateService) demands
    When the analyzer checks the file
    Then no diagnostic is reported (the parameter type is not the test's free choice)

  Scenario: a _test.go wiring helper's result is forced by a production constructor — skipped
    Given a _test.go helper function in /app/api returns "grpc.Notifier" from /server/grpc
    When the analyzer checks the file
    Then no diagnostic is reported

  # --- Boundary class: a _test.go file is not exempted wholesale ---

  Scenario: a struct declared in a _test.go file still gets its fields checked
    Given a struct type declared inside a _test.go file in /app/api has a field
      typed "grpc.Notifier" from /server/grpc
    When the analyzer checks the file
    Then a "GID-134" diagnostic is reported (the field type was the test's own choice, not forced)

  Scenario: the same helper shape in a non-test file is still a violation
    Given a non-test file in /app/api declares a function with the exact shape of the
      exempted _test.go helper (a parameter typed "grpc.Notifier")
    When the analyzer checks the file
    Then a "GID-134" diagnostic is reported (the _test.go exception does not leak into production code)

  # --- GID-269: no inline interface declarations in struct fields ---

  Scenario: a struct field declares a non-empty anonymous interface — violation
    Given a struct field has the type "interface{ Resolve() error }"
    When the analyzer checks the file
    Then a "GID-269" diagnostic tells the developer to declare a named interface next to the struct

  Scenario: a struct field uses a named local interface — clean
    Given a struct field has the type "LocalRepository"
    When the analyzer checks the file
    Then no "GID-269" diagnostic is reported

  Scenario: an empty interface struct field — boundary excluded
    Given a struct field has the type "interface{}"
    When the analyzer checks the file
    Then no "GID-269" diagnostic is reported

  Scenario: an anonymous interface outside a struct field — not applicable
    Given a method parameter has the type "interface{ Foo() }"
    When the analyzer checks the file
    Then no "GID-269" diagnostic is reported

# --- Checklist when adding the rules ---
#  [x] IDs and descriptions are recorded in the registry (RULES.md, GID-134 and GID-269)
#  [x] Layer chosen: go/analysis (AST shape and TypesInfo are needed)
#  [x] Severity and messages are defined ("GID-134: ...", "GID-269: ...")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rules enabled through gidifaceplace in both embedded configs and .golangci.yml
