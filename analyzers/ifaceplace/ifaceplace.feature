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

  Not affected: anonymous interfaces, error, any/interface{},
  generic constraints. Generated code (ast.IsGenerated) is skipped.

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

  Scenario: an anonymous interface{ Foo() } in a parameter — not named, skipped
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

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-134)
#  [x] Layer chosen: go/analysis (TypesInfo — interface types and the declaration package are needed)
#  [x] Severity and message are defined ("GID-134: ...")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml (outside the scope of this task)
