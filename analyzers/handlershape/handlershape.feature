# language: en

Feature: GID-230 — gRPC handler shape (grpc-handler-shape)
  As a developer
  I want every transport handler to be a struct with a
  Handle(ctx context.Context, req *T) (*R, error) method depending on
  <Handler>Validator / <Handler>Service interfaces, and the gRPC service
  struct to expose handlers as exported *Handler fields
  So that the transport layer follows the server.md template uniformly

  # Scope (pathseg, not strings.Contains):
  #   - handler packages: import path contains segment "server" and ends with
  #     segment "handler" (handler/convert, handler/validate are out of scope);
  #   - service struct check: any other package under segment "server", on
  #     structs embedding a type named Unimplemented*Server.
  # Exported struct types only; names with the Options suffix are skipped.
  # Handle: enough that the first param is context.Context and the last result
  # is error — request/response types differ per RPC and are not constrained.

  Scenario: positive — exported struct without a Handle method
    Given a handler package declares "type Purge struct{}" without a Handle method
    When the analyzer checks the file
    Then a "GID-230" diagnostic is reported on type "Purge"

  Scenario: positive — Handle without ctx as the first parameter
    Given a handler with method "func (h *Update) Handle(req *Req) error"
    When the analyzer checks the file
    Then a "GID-230" diagnostic is reported on type "Update"

  Scenario: positive — Handle whose last result is not error
    Given a handler with method "func (h *Remove) Handle(ctx context.Context, req *Req) (*Resp, bool)"
    When the analyzer checks the file
    Then a "GID-230" diagnostic is reported on type "Remove"

  Scenario: positive — interface dependency not named <Handler>Validator/<Handler>Service
    Given handler "Export" has a field "validator DocumentsValidator"
    When the analyzer checks the file
    Then a "GID-230" diagnostic is reported on field "validator" expecting "ExportValidator" or "ExportService"

  Scenario: positive — gRPC service struct field without the Handler suffix
    Given a struct embedding "consentpb.UnimplementedConsentServiceServer" has a field "Purge *handler.Purge"
    When the analyzer checks the file
    Then a "GID-230" diagnostic is reported on field "Purge"

  Scenario: positive — gRPC service struct with an unexported handler field
    Given a struct embedding "consentpb.UnimplementedConsentServiceServer" has a field "exportHandler *handler.Export"
    When the analyzer checks the file
    Then a "GID-230" diagnostic is reported on field "exportHandler"

  Scenario: negative — canonical handler
    Given handler "Documents" with fields "validator DocumentsValidator; service DocumentsService" and method "func (h *Documents) Handle(ctx context.Context, req *Req) (*Resp, error)"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — type with the Options suffix is not a handler
    Given a handler package declares "type ListOptions struct{}" without a Handle method
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — unexported struct is not flagged
    Given a handler package declares "type helper struct{}" without a Handle method
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — canonical gRPC service struct
    Given a struct embeds "consentpb.UnimplementedConsentServiceServer" and only has exported "*Handler" fields
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — Handle with a value receiver
    Given a handler with method "func (h Stats) Handle(ctx context.Context, req *Req) (*Resp, error)"
    When the analyzer checks the file
    Then no diagnostic is reported

  # Decision recorded in the rule: only ctx-first and error-last are enforced;
  # the number of params after ctx and of results before error is free.
  Scenario: boundary — Handle with extra params and ctx-only request
    Given a handler with method "func (h Ping) Handle(ctx context.Context) error"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — non-interface fields of a handler are not checked
    Given handler "Documents" also has a field "timeout int"
    When the analyzer checks the file
    Then no diagnostic is reported on field "timeout"

  Scenario: boundary — struct without Unimplemented*Server embed in a service package
    Given the package "internal/server/grpc/consent" declares "type Config struct{ Addr string }"
    When the analyzer checks the file
    Then no diagnostic is reported

  # Dropped checks (false-positive risk, recorded as non-applicability):
  #   - anonymous interface literal fields and embedded interfaces in a handler
  #     struct are skipped — the naming check applies to named interfaces only;
  #   - handler name vs RPC method name equality is not checked — the proto
  #     descriptor is not available to the analyzer;
  #   - the "exactly two interfaces" cardinality is not enforced — only naming
  #     of the interface-typed fields that do exist.
  Scenario: non-applicability — embedded interface in a handler struct
    Given handler "Stream" embeds "io.Closer" and has a canonical Handle method
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — handler-like struct outside the server layer
    Given the package "svc/handler" (no "server" segment) declares "type Job struct{}" without Handle
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — Unimplemented*Server embed outside the server layer
    Given the package "internal/domain/service" declares a struct embedding "UnimplementedFooServer" with non-Handler fields
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — type listed in settings.exclude
    Given type "HealthCheck" is listed in settings.exclude
    And a handler package declares "type HealthCheck struct{}" without a Handle method
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — service struct listed in settings.exclude
    Given type "Job" is listed in settings.exclude
    And struct "Job" embeds "consentpb.UnimplementedConsentServiceServer" with non-Handler fields
    When the analyzer checks the file
    Then no diagnostic is reported

# --- Checklist for adding a new rule ---
#  [x] ID and description registered in the registry (RULES.md — outside this change)
#  [x] Layer chosen: go/analysis (types needed — context.Context, error, interfaces)
#  [x] Severity and message defined
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml (outside this change)
