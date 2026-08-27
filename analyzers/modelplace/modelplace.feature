# language: en

Feature: GID-270 — data types live only in /domain/model
  As a developer
  I want data structures to be declared in /domain/model and nowhere else
  So that the domain vocabulary has one home and convert/service/usecase stay free of it
  Sources: owner requirement 2026-08-27 — «Не может convert возвращать какую-то
  бизнес сущность. Бизнес сущности объявляются только в пакете model»; «Все
  модели данных живут только в domain/model. Не могут они вот так просто
  объявляться в service, usecase, convert»

  Linter: gidmodelplace. LoadMode: TypesInfo (the declaring package of a
  receiver/result type is resolved through types.Named.Obj().Pkg()).

  Part A: a package whose final path segment is convert (the GID-247 boundary,
  pathseg.EndsWith) declares no types at all — only struct declarations are
  judged. A convert package nested under /domain is judged by part A only.

  Part B: in /domain/service and /domain/usecase (matched by path segments —
  internal/domain/... and pkg/<module>/domain/... alike), an exported struct
  is a data model when ALL hold: no methods in this package (a struct with
  behavior is the layer's entity), not returned by any New/new-prefixed
  function of the package (a pointer result counts — what a constructor
  assembles is the layer's entity), name not ending with an options suffix
  (settings.suffixes, default Options/Option/Config/Params/Settings).

  Not judged: unexported types, _test.go files (GID-250), generated code,
  non-struct types in convert (interface, named basic type).
  Exceptions: //nolint:gidmodelplace, settings.exclude (the "Struct" form).

  # --- Positive class: the violation is caught ---

  Scenario: a struct declared in a convert package (part A)
    Given a package "…/usecase/convert" declares "type Build struct"
    When the analyzer checks the file
    Then a "GID-270" diagnostic is reported telling to declare the type in /domain/model

  Scenario: an exported data struct in /domain/service (part B)
    Given "/domain/service" declares "type TriggerBuild struct" with three fields and no methods
    When the analyzer checks the file
    Then a "GID-270" diagnostic is reported telling to move the type to /domain/model

  Scenario: an exported data struct in /domain/usecase (part B)
    Given "/domain/usecase" declares "type AltCraftTriggerBuild struct" with no methods
    When the analyzer checks the file
    Then a "GID-270" diagnostic is reported telling to move the type to /domain/model

  Scenario: a struct in a convert package nested under /domain is reported once
    Given "/domain/usecase/convert" declares an exported struct
    When the analyzer checks the file
    Then exactly one "GID-270" diagnostic is reported (part A only)

  # --- Negative class: clean code passes ---

  Scenario: /domain/model declares the data struct — the canonical place
    Given "/domain/model" declares "type User struct"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a package outside /domain/** declares a struct
    Given "/server/http" declares "type Request struct"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a convert package holds no type declarations
    Given a convert package with mapping functions only
    When the analyzer checks the file
    Then no diagnostic is reported

  # --- Boundary class: the exemptions carry the precision ---

  Scenario: a struct with a method in /domain/service
    Given "/domain/service" declares "type Processor struct" with a method "func (p *Processor) Run()"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a struct returned by a constructor (pointer result)
    Given "/domain/service" declares "func NewIntegration(repo Repository) *Integration"
    When the analyzer checks the file
    Then no diagnostic is reported for "Integration"

  Scenario: a struct returned by a lowercase-new constructor
    Given "/domain/service" declares "func newPair(a, b string) *Pair"
    When the analyzer checks the file
    Then no diagnostic is reported for "Pair"

  Scenario: a struct returned by a value-returning constructor
    Given "/domain/usecase" declares "func NewReport() Report"
    When the analyzer checks the file
    Then no diagnostic is reported for "Report"

  Scenario: a struct named by the options convention
    Given "/domain/service" declares "type ServerOptions struct" (also Option, Config, Params, Settings)
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: settings.suffixes replaces the defaults
    Given settings.suffixes is "['Spec']" and "/domain/service" declares "type LegacyOptions struct"
    When the analyzer checks the file
    Then a "GID-270" diagnostic is reported for "LegacyOptions" — the default Options suffix no longer exempts

  # --- Non-applicability class ---

  Scenario: an unexported struct in /domain/service
    Given "/domain/service" declares "type draft struct" (unexported)
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: an unexported type in a convert package
    Given a convert package declares "type kind struct" (unexported)
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a non-struct type in a convert package
    Given a convert package declares "type Kind string" and "type Mapper interface"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a struct declared in a _test.go file
    Given "/domain/service/service_test.go" declares "type TestHarness struct"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a struct in generated code
    Given a file with the "Code generated … DO NOT EDIT." marker declares "type GenConfig struct"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: settings.exclude exempts a struct in both parts
    Given settings.exclude is "['LegacyDTO', 'Exempt']" and "LegacyDTO" is declared in /domain/service, "Exempt" in convert
    When the analyzer checks the files
    Then no diagnostic is reported for either

  Scenario: pinpoint exclusion via //nolint
    Given a violating struct declaration carries "//nolint:gidmodelplace"
    When golangci-lint runs the rule
    Then no diagnostic is reported for that declaration