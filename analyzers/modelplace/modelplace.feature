# language: en

Feature: GID-270 — data types live only in /domain/model
  As a developer
  I want data structures to be declared in /domain/model and nowhere else
  So that the domain vocabulary has one home and convert/service/usecase neither declare nor hand out types of their own
  Sources: owner requirement 2026-08-27 — «Не может convert возвращать какую-то
  бизнес сущность. Бизнес сущности объявляются только в пакете model»; «Все
  модели данных живут только в domain/model. Не могут они вот так просто
  объявляться в service, usecase, convert»

  Linter: gidmodelplace. LoadMode: TypesInfo (the declaring package of a
  receiver/result type is resolved through types.Named.Obj().Pkg()).

  Part A: a package whose final path segment is convert (the GID-247 boundary,
  pathseg.EndsWith) declares no types at all — only struct declarations are
  judged. The layer above convert does not matter: /dal/repository/convert,
  /server/grpc/convert and /domain/usecase/convert are all judged, and a
  convert package nested under /domain is judged by part A only.

  Part B: in /domain/service and /domain/usecase (matched by path segments —
  internal/domain/... and pkg/<module>/domain/... alike), an exported struct
  is a data model when ALL hold: no methods in this package (a struct with
  behavior is the layer's entity), not returned by any New/new-prefixed
  function of the package (a pointer result counts — what a constructor
  assembles is the layer's entity), name not ending with an options suffix
  (settings.suffixes, default Options/Option/Config/Params/Settings).

  Part C: in the same packages no function returns a struct declared in that
  same package — the result type is looked through a pointer, slice, array,
  map value and channel, and a function literal (including one nested in a
  body) is judged like any other function. Part C closes the bypass parts A
  and B leave open: declaring the data struct unexported and handing it out
  anyway. The exemptions are part B's, in every package (methods, a New/new
  constructor, an options suffix); a type already reported at its declaration
  is not reported a second time on the return.

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

  Scenario: a struct declared in a convert package outside /domain (part A)
    Given "/dal/repository/convert" declares "type Row struct"
    When the analyzer checks the file
    Then a "GID-270" diagnostic is reported — the layer above convert does not matter

  Scenario: a converter returns an unexported struct of its own package (part C)
    Given a convert package declares "type kind struct" and "func buildKind(code string) kind"
    When the analyzer checks the file
    Then a "GID-270" diagnostic is reported on the result type

  Scenario: a converter outside /domain returns its own struct (part C)
    Given "/dal/repository/convert" declares "func nextCursor(offset int) cursor"
    When the analyzer checks the file
    Then a "GID-270" diagnostic is reported

  Scenario: a method of /domain/service returns a struct of its package (part C)
    Given "/domain/service" declares "func (p *Processor) Snapshot() snapshot"
    When the analyzer checks the file
    Then a "GID-270" diagnostic is reported on the result type

  Scenario: the struct leaves the package inside a container (part C)
    Given "/domain/service" declares "func (p *Processor) Snapshots() []snapshot" and "/domain/usecase" declares "func summarize(rows []string) map[string]digest"
    When the analyzer checks the files
    Then a "GID-270" diagnostic is reported for each — a pointer, slice, array, map value and channel are looked through

  Scenario: a function literal nested in a body returns the package's struct (part C)
    Given a method body declares "sum := func() tally { … }"
    When the analyzer checks the file
    Then a "GID-270" diagnostic is reported for the literal

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

  Scenario: a type already reported at its declaration is not reported again (part C)
    Given a convert package declares "type Build struct" and "func BuildFromRow(id string) *Build"
    When the analyzer checks the file
    Then exactly one "GID-270" diagnostic is reported — on the declaration

  Scenario: the layer's entity is handed out by its own methods (part C)
    Given "/domain/service" declares "func (p *Processor) WithQueue(queue []string) *Processor"
    When the analyzer checks the file
    Then no diagnostic is reported — Processor has methods

  Scenario: a function returns what a constructor assembles (part C)
    Given "/domain/service" declares "func Rebuild(repo Repository) *Integration" next to "NewIntegration"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: an options-suffixed struct is returned (part C)
    Given "/domain/service" declares "func buildOptions() runOptions"
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

  Scenario: a function returns a struct of ANOTHER package
    Given "/domain/service" declares "func Timeout() model.User"
    When the analyzer checks the file
    Then no diagnostic is reported — the type is declared in /domain/model

  Scenario: a function of a package outside the rule's scope returns its own struct
    Given "/server/http" declares "func Decode(query string) Request"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a function of a _test.go file or of generated code returns the package's struct
    Given "service_test.go" declares "func harness() TestHarness" and a generated file declares "func genConfig() GenConfig"
    When the analyzer checks the files
    Then no diagnostic is reported for either

  Scenario: settings.exclude exempts a struct in all three parts
    Given settings.exclude is "['LegacyDTO', 'Exempt']", "LegacyDTO" is declared in /domain/service and returned by "makeLegacy", "Exempt" in convert and returned by "makeExempt"
    When the analyzer checks the files
    Then no diagnostic is reported for either

  Scenario: pinpoint exclusion via //nolint
    Given a violating struct declaration carries "//nolint:gidmodelplace"
    When golangci-lint runs the rule
    Then no diagnostic is reported for that declaration