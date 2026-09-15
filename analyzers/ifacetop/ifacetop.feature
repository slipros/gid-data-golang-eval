# language: en

Feature: GID-276 — interfaces open the file (gidifacetop)
  As a developer
  I want the dependency interfaces of a file declared before anything else it declares
  So that the reader sees what an entity depends on before the entity itself

  # One pass over the file's top-level declarations in source order. The first
  # declaration that is not an interface — a non-interface type spec or any
  # function (a method included) — is the boundary; every interface spec found
  # below it is reported on its name, citing the boundary and its line.
  # import, const and var are not boundaries: GID-130 already puts them above
  # types and functions, and a var below an interface is its diagnostic.
  # An interface is recognised by type information (underlying *types.Interface),
  # so an alias or a defined type over an interface counts, and so does a
  # constraint interface. A grouped type ( … ) block is judged spec by spec.
  # Scope: packages in /domain/service, /domain/usecase, /client, /dal/repository
  # and their subpackages (pathseg.HasLayer — anchored to the module root, so a
  # pkg/<module>/domain/service is in scope and an internal/app/client is not).
  # Not judged: generated code (ast.IsGenerated) and _test.go files (srcfile.IsTest).

  # --- Class 1: positive ---

  Scenario: positive — interfaces below the struct that holds them
    Given a file in /domain/service with "type Order struct" and then "type OrderRepository interface" and "type OrderWalletService interface"
    When the gidifacetop analyzer checks the file
    Then the diagnostic "GID-276: interface OrderRepository is declared below type Order (line 5); interfaces open the file, right after import, const and var. Fix: move type OrderRepository interface above type Order" is reported on "OrderRepository"
    And the same diagnostic is reported on "OrderWalletService"

  Scenario: positive — an interface after the entity's last method
    Given a file in /domain/service with "type Invoice struct", "NewInvoice", "func (i *Invoice) Invoice" and then "type InvoiceRepository interface"
    When the gidifacetop analyzer checks the file
    Then a "GID-276" diagnostic citing "type Invoice (line 5)" is reported on "InvoiceRepository"

  Scenario: positive — a free function is a boundary
    Given a file in /domain/service with "func pageSize" and then "type PageCounter interface"
    When the gidifacetop analyzer checks the file
    Then a "GID-276" diagnostic citing "func pageSize (line 3)" is reported on "PageCounter"

  Scenario: positive — a method is a boundary and is named with its receiver
    Given a file in /domain/service with "func (h *Hello) Ping" and then "type PingNotifier interface"
    When the gidifacetop analyzer checks the file
    Then a "GID-276" diagnostic citing "func (*Hello) Ping (line 3)" is reported on "PingNotifier"

  Scenario: positive — every layer in scope
    Given an interface declared below a struct in /domain/usecase, /client/billing, /dal/repository and pkg/billing/domain/service
    When the gidifacetop analyzer checks the packages
    Then a "GID-276" diagnostic is reported on each of those interfaces

  Scenario: positive — strict: a client's DTO is a boundary too
    Given a file in /client/billing with "type ChargeIn struct", "type Options struct" and then "type Core interface"
    When the gidifacetop analyzer checks the file
    Then a "GID-276" diagnostic citing "type ChargeIn (line 3)" is reported on "Core"

  # --- Class 2: negative ---

  Scenario: negative — the canonical layout
    Given a file in /domain/service with import, const, var, "type HelloRepository interface", "type HelloMetrics interface", "type Hello struct", "NewHello" and methods
    When the gidifacetop analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — interfaces above a client's DTO
    Given a file in /client/billing with "type Metrics interface" and then "type RefundIn struct"
    When the gidifacetop analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — a file of interfaces only, a constraint interface included
    Given a file in /domain/service with "type EmptyMarker interface{}" and "type NumberConstraint interface { ~int | ~int64 }"
    When the gidifacetop analyzer checks the file
    Then no diagnostic is reported

  # --- Class 3: boundary ---

  Scenario: boundary — a grouped type block is judged spec by spec
    Given a file in /domain/service with "type ( CartRepository interface; CartMetrics interface )" and then "type ( Cart struct; CartNotifier interface )"
    When the gidifacetop analyzer checks the file
    Then a "GID-276" diagnostic citing "type Cart (line 15)" is reported on "CartNotifier" only

  Scenario: boundary — an alias and a defined type over an interface are interfaces
    Given a file in /domain/service with "type Upload struct", "type UploadReader = io.Reader", "type UploadCloser io.Closer" and "type UploadSize int"
    When the gidifacetop analyzer checks the file
    Then a "GID-276" diagnostic is reported on "UploadReader" and on "UploadCloser"
    And nothing is reported on "UploadSize"

  Scenario: boundary — an inline interface field and a local type are not declarations of the file
    Given a file in /domain/service with "type Report struct { source interface{ Read() []byte } }" and a method declaring "type reportSink interface" inside its body
    When the gidifacetop analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — const and var below types are not this rule's concern
    Given a file in /domain/service whose const and var blocks follow its types and functions
    When the gidifacetop analyzer checks the file
    Then no GID-276 diagnostic is reported

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — layers outside the scope
    Given an interface declared below a struct in /domain/model, /server/grpc and internal/app/client
    When the gidifacetop analyzer checks the packages
    Then no diagnostic is reported

  Scenario: non-applicability — a _test.go file
    Given "hello_test.go" in /domain/service with "type fakeHelloRepository struct", its method and then "type helloHarness interface"
    When the gidifacetop analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — generated code
    Given a file in /domain/service marked "Code generated by mockgen. DO NOT EDIT." with a struct and then an interface
    When the gidifacetop analyzer checks the file
    Then no diagnostic is reported
