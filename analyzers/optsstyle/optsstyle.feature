# language: en

Feature: GID-152 — opts by pointer in parameters, an unexported field in the entity (gidoptsstyle)
  As a developer
  I want options passed as *XxxOptions and stored as an unexported field
  So that the entity's public API stays its methods, and options are not
  promoted into it by an embedded struct

  # One analyzer over parameters and struct fields, LoadModeTypesInfo.
  # An Options type is a named type whose name ends with "Options" and which is
  # declared in the package under analysis. Options types from the stdlib, the
  # module cache or another first-party package are never reported: the rule
  # governs how an entity stores ITS OWN options, and a foreign type cannot be
  # fixed at the use site anyway.
  # Three findings:
  #   1. an Options parameter taken by value -> pass *XxxOptions;
  #   2. an embedded (anonymous) Options field, pointer or not -> it promotes
  #      the option fields into the public API;
  #   3. an exported named Options field -> opts must be unexported.
  # An unexported named field (opts Options / opts *Options) is the required
  # shape and is never reported.
  # Scope: the layers that own a constructor with opts — by default handler
  # leaf packages (settings.leaf, matched by trailing segments, so
  # handler/convert and handler/validate are out) and /domain/service,
  # /domain/usecase (settings.within, anchored to the module root). The config
  # and composition layers are deliberately out: they legitimately hold library
  # Options structs (httpserver.Options, …). Setting either list in
  # .golangci.yml replaces the whole default scope.
  # Generated code (ast.IsGenerated) is skipped.

  # --- Class 1: positive ---

  Scenario: positive — an Options parameter by value
    Given "func NewBad(opts Options) int" in /domain/service
    When the gidoptsstyle analyzer checks the file
    Then the diagnostic "GID-152: opts must be passed by pointer. Fix: use *Options" is reported on the parameter

  Scenario: positive — an embedded Options field
    Given "type Hello struct { Options }"
    When the gidoptsstyle analyzer checks the file
    Then the diagnostic "GID-152: embedding Options is forbidden: it promotes option fields into the public API. Fix: use an unexported named field `opts Options`" is reported

  Scenario: positive — an embedded pointer to Options
    Given "type Hello struct { *Options }"
    When the gidoptsstyle analyzer checks the file
    Then the diagnostic "GID-152: embedding Options is forbidden: …" is reported
    # A pointer changes nothing: the fields are promoted either way.

  Scenario: positive — an exported Options field
    Given "type Hello struct { Opts Options }"
    When the gidoptsstyle analyzer checks the file
    Then the diagnostic "GID-152: Options field \"Opts\" must be unexported. Fix: rename to `opts Options`" is reported

  # --- Class 2: negative ---

  Scenario: negative — the required shape
    Given "type Hello struct { opts *HelloOptions }" and "func NewHello(opts *HelloOptions) *Hello"
    When the gidoptsstyle analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — an Options type from another package
    Given "type Server struct { opts httpserver.Options }" where httpserver is an external library
    When the gidoptsstyle analyzer checks the file
    Then no diagnostic is reported
    # Only Options types declared in the same package are governed.

  # --- Class 3: boundary ---

  Scenario: boundary — a locally named Options type
    Given "type Hello struct { LocalOptions }" where LocalOptions is declared in the package
    When the gidoptsstyle analyzer checks the file
    Then the diagnostic "GID-152: embedding LocalOptions is forbidden: …" is reported
    # Any name ending with Options counts, not just the bare "Options".

  Scenario: boundary — a handler leaf package
    Given "type Handler struct { Options }" in /server/grpc/order/handler
    When the gidoptsstyle analyzer checks the file
    Then the diagnostic "GID-152: embedding Options is forbidden: …" is reported
    # A handler is matched by its trailing segment, at any depth.

  Scenario: boundary — a domain/service nested under another layer
    Given the same struct in svc/client/gateway/domain/service
    When the gidoptsstyle analyzer checks the file
    Then no diagnostic is reported
    # settings.within is anchored to the module root.

  Scenario: boundary — settings replace the default scope
    Given settings "leaf: [[config]]" and an embedded Options struct in /config
    When the gidoptsstyle analyzer checks the file
    Then the diagnostic "GID-152: embedding Options is forbidden: …" is reported
    # Setting either list replaces the whole default scope — the handler and
    # domain defaults are gone.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — the config layer with the default scope
    Given "type App struct { Options }" in /config
    When the gidoptsstyle analyzer checks the file
    Then no diagnostic is reported
    # The composition layer holds library Options structs by design.

  Scenario: non-applicability — a type whose name merely contains "Options"
    Given "type Hello struct { OptionsCache cache }"
    When the gidoptsstyle analyzer checks the file
    Then no diagnostic is reported
    # The suffix is what identifies an Options type.

  Scenario: non-applicability — generated code
    Given a file carrying the "Code generated … DO NOT EDIT." header
    When the gidoptsstyle analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-152)
#  [x] Layer chosen: go/analysis (package optsstyle)
#  [x] Message is defined ("GID-152: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest (default and custom scope)
#  [x] Rule enabled in .golangci.yml
