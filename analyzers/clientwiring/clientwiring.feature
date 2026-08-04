# language: en

Feature: GID-255 — a package in /client/** that declares no client is wiring
  As the styleguide owner
  I want a package in the client layer to actually contain a client — a type whose methods call
  the external API — and a package that only assembles a connection to live in the composition root
  So that the client layer means something: when it holds nothing but a constructor, the layer is
  not being used, it is being imitated — the directory and the import cost something and give
  nothing back, and the consumer ends up talking to the foreign service's transport types.

  # Layer: go/analysis (package clientwiring, linter gidclientwiring), LoadModeTypesInfo.
  # Origin: ad-cabinet-connector internal/client/resourceregistry (2026-08-04) — two functions,
  # NewConnection (assembles a gRPC connection out of options, logger and metrics; returns the grpc
  # library's type) and NewLoggingDecider (its logging policy). No client type, no method, no model;
  # the only consumer is internal/app/api. Assembling a transport object once at start-up out of
  # options and dependencies IS the composition root's job — the package moves into /app/** as is.
  #
  # Why "a constructor and no method" is the right symptom: a real client (client.md) owns a type
  # whose methods call the external API, owns its own models and converts them, so the consumer
  # never sees the transport. Without that type, the layer contributes nothing — and in the origin
  # case the domain ended up importing the foreign service's pb types directly.
  #
  # Detect (per package): the package lies under /client/** in a module laid out as a service,
  # declares at least one function, and NOT ONE method (a FuncDecl with a receiver). One diagnostic
  # per package, on the package clause.
  #
  # Not flagged: a package with methods (a real client, however thin); a package of pure
  # types/constants with no functions at all (client models); the leaf subpackages
  # convert/dto/mock/mocks, function-only by design. Generated code and _test.go are ignored when
  # deciding. A flat library module is skipped (internal/modlayout) — libs/grpc.git/client is a
  # library's own client package, not a service layer.
  # Exceptions: //nolint:gidclientwiring on the package clause.

  # --- Class 1: positive (the violation is caught) ---

  Scenario: positive — a connection factory living in the client layer (the rule's origin)
    Given the package "internal/client/resourceregistry" declaring "func NewConnection(opts *Options) (*connection, error)" and "func NewLoggingDecider(logBody bool) func(string) bool", with no method
    When the gidclientwiring analyzer checks the package
    Then the diagnostic "GID-255: package \"resourceregistry\" is in the client layer but declares no client — it has functions and not one method, so nothing here calls an external API on a type of its own. A package that only assembles a connection out of options, a logger and metrics is the composition root spelled in the wrong directory, and it leaves the consumer talking to the foreign service's DTOs. Fix: move it as is into /app/** (wiring), or give the layer a real client — a type whose methods call the API and return the client's own models (client.md)" is reported on the package clause
    # Note the second tell in the fixture: the constructor returns a type declared elsewhere — a
    # factory builds a foreign object, a client never does.

  # --- Class 2: negative (clean code passes) ---

  Scenario: negative — a real client: a type whose method calls the API and returns its own model
    Given the package "internal/client/payments" declaring "type Client struct", "func NewClient(addr string) *Client" and the method "func (c *Client) Payment(ctx context.Context, id string) (Payment, error)"
    When the gidclientwiring analyzer checks the package
    Then no diagnostic is reported
    # The constructor is fine precisely because it is not the only thing in the package.

  # --- Class 3: boundary (looks like a violation but is allowed) ---

  Scenario: boundary — a convert leaf under /client/** is function-only by design
    Given the package "internal/client/catalog/convert" declaring "func ItemFromAPI(id string) Item" and no method
    When the gidclientwiring analyzer checks the package
    Then no diagnostic is reported
    # A converter is a pure function over vocabulary types (GID-235); having no method is correct
    # there, not a missing client. Same for dto/mock/mocks leaves.

  Scenario: boundary — a package of pure types under /client/** (the client's models)
    Given the package "internal/client/models" declaring "type Account struct" and no function at all
    When the gidclientwiring analyzer checks the package
    Then no diagnostic is reported
    # Nothing here pretends to be a client, so there is nothing to move — the rule needs a function
    # to fire.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — the same factory in the composition root
    Given the package "internal/app/api" declaring "func NewRegistryConnection(opts *Options) (*connection, error)" and no method
    When the gidclientwiring analyzer checks the package
    Then no diagnostic is reported
    # /app/** is not the client layer: a function-only wiring package is the norm there — that is
    # exactly where the origin package is being sent.

  Scenario: non-applicability — a function-only package in another layer
    Given the package "internal/domain/service" declaring "func NewSnapshot(id string) Snapshot" and no method
    When the gidclientwiring analyzer checks the package
    Then no diagnostic is reported
    # Package functions in the service layer are GID-133's business, not this rule's.

  Scenario: non-applicability — a flat library module
    Given a module whose root holds no service layer directory (libs/grpc.git with its own client package)
    When the gidclientwiring analyzer checks the package
    Then no diagnostic is reported
    # internal/modlayout: a library's client package is not a service layer.

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-255)
#  [x] Layer chosen: go/analysis (package clientwiring: gidclientwiring)
#  [x] Message is defined ("GID-255: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
