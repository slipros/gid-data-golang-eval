# language: en

Feature: GID-266 — a gRPC client call in a BFF carries a MappedFields option (gidmappedfields)
  As a developer of a BFF
  I want every call to another service to carry the mapping of its field names
  onto the names of the request the frontend sent
  So that a validation error of the callee arrives in the vocabulary the client
  speaks, instead of naming fields of a contract the frontend never saw

  # A BFF validates the request it received and forwards it to the service that
  # owns the data. That service validates it again, in its own vocabulary, and
  # answers InvalidArgument with a ValidationError whose fields carry ITS names
  # (genproto udmpapis/type/error). Passed through, that error reaches the
  # browser naming fields nobody sent — the form has nothing to highlight.
  #
  # The library closes it with a per-call option:
  #   resp, err := c.client.CreateSegment(ctx, req,
  #       gdgrpcerror.WithMappedFields(gdmapper.MappedFields{
  #           gdmapper.NewMappedFieldStringEqualWithWholePart("segment_id", "segmentId"),
  #       }))
  # MappedFieldsUnaryClientInterceptor picks the option out of the call options
  # and renames the fields of a ValidationError before it travels on. Without
  # the option it forwards the error untouched — and nothing about the call
  # looks broken until somebody submits a form and reads the answer.
  #
  # Scope — a BFF module only: laid out as a service (a composition root of its
  # own) and owning no data layer (no /dal, no /repository). That is the module
  # whose business logic IS calling other services over gRPC — lk-api, the same
  # shape GID-160 was taught to leave alone. Judged layers: /domain/service and
  # /domain/usecase, mirroring GID-160.
  #
  # A call is recognised by its signature: a variadic trailing parameter of type
  # grpc.CallOption — the marker GID-176 already uses, and one that covers both
  # a generated client and the consumer-side interface mirroring it (GID-134).
  #
  # The option is recognised by the name of its type (a named type whose name
  # contains "MappedFields"), so a value held in a variable counts as well as a
  # fresh WithMappedFields call, and a project helper carrying the same marker
  # in its name is recognised too.
  #
  # A call sending no request data is not judged: an RPC taking nothing beside
  # the context, a nil request, an empty literal. The callee has no field to
  # reject, so no ValidationError can name one — and the rule would be asking
  # for a mapping that has to be empty, the very thing its second diagnostic
  # reports.
  #
  # Escape hatches: //nolint:gidmappedfields, settings.exclude ("Method" |
  # "Client.Method" — the RPC being CALLED).

  # --- Class 1: positive ---

  Scenario: positive — the call carries no options at all
    Given "o.client.CreateOrder(ctx, in)" in /domain/service of a BFF
    When the gidmappedfields analyzer checks the file
    Then the diagnostic "GID-266: the gRPC call CreateOrder carries no MappedFields option" is reported on the call

  Scenario: positive — the call carries other options but no mapping
    Given "o.client.Order(ctx, &req, grpc.WaitForReady(true))"
    When the gidmappedfields analyzer checks the file
    Then the diagnostic "GID-266: the gRPC call Order carries no MappedFields option" is reported on the call

  Scenario: positive — the mapping is nil
    Given "o.client.CreateOrder(ctx, in, gdgrpcerror.WithMappedFields(nil))"
    When the gidmappedfields analyzer checks the file
    Then the diagnostic "GID-266: the MappedFields option of the gRPC call CreateOrder is empty" is reported on the option
    # The interceptor returns early on len(MappedFields) == 0: an empty mapping
    # is no mapping, spelled in a way that looks handled.

  Scenario: positive — the mapping is an empty literal
    Given "o.client.CreateOrder(ctx, in, gdgrpcerror.WithMappedFields(gdmapper.MappedFields{}))"
    When the gidmappedfields analyzer checks the file
    Then the diagnostic "GID-266: the MappedFields option of the gRPC call CreateOrder is empty" is reported on the option

  Scenario: positive — the option is written as a literal with no mapping field
    Given "o.client.CreateOrder(ctx, in, gdgrpcerror.MappedFieldsInterceptorCallOption{})"
    When the gidmappedfields analyzer checks the file
    Then the diagnostic "GID-266: the MappedFields option of the gRPC call CreateOrder is empty" is reported on the option

  Scenario: positive — the request is built in the call and carries a field
    Given "o.client.CreateOrder(ctx, &orderpb.CreateOrderRequest{Name: name})"
    When the gidmappedfields analyzer checks the file
    Then the diagnostic "GID-266: the gRPC call CreateOrder carries no MappedFields option" is reported on the call
    # The field is there to be rejected by the name of the foreign contract.

  # --- Class 2: negative ---

  Scenario: negative — the mapping is passed
    Given "o.client.CreateOrder(ctx, in, gdgrpcerror.WithMappedFields(orderFields))"
    When the gidmappedfields analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — the option is prepared in a variable
    Given "option := gdgrpcerror.WithMappedFields(orderFields)" and "o.client.CreateOrder(ctx, in, option)"
    When the gidmappedfields analyzer checks the file
    Then no diagnostic is reported
    # The option is recognised by the name of its type, not by the call shape.

  # --- Class 3: boundary ---

  Scenario: boundary — the mapping is not the first option
    Given "o.client.CreateOrder(ctx, in, grpc.WaitForReady(true), gdgrpcerror.WithMappedFields(orderFields))"
    When the gidmappedfields analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — the caller spreads a prepared slice of options
    Given "o.client.CreateOrder(ctx, in, opts...)"
    When the gidmappedfields analyzer checks the file
    Then no diagnostic is reported
    # The option may well be inside the slice; judging the call would be a guess.

  Scenario: boundary — a client interface that does not forward the call options
    Given "o.catalog.Items(ctx)" where Items has no "opts ...grpc.CallOption"
    When the gidmappedfields analyzer checks the file
    Then no diagnostic is reported
    # There is nowhere to pass the option, so the rule has nothing to ask for.
    # Trimming the options out of an interface hides the call from this rule —
    # a review matter, not a linter one.

  Scenario: boundary — the RPC takes nothing beside the context
    Given "o.client.Ping(ctx)" where Ping has no request parameter
    When the gidmappedfields analyzer checks the file
    Then no diagnostic is reported
    # Nothing was sent, so nothing can be rejected by name.

  Scenario: boundary — the request is nil
    Given "o.client.DeleteOrder(ctx, nil)"
    When the gidmappedfields analyzer checks the file
    Then no diagnostic is reported
    # Measured on lk-api: Statuses(ctx, nil) and FieldActivities(ctx, nil).

  Scenario: boundary — the request is an empty literal
    Given "o.client.CreateOrder(ctx, &orderpb.CreateOrderRequest{})"
    When the gidmappedfields analyzer checks the file
    Then no diagnostic is reported
    # The same case spelled out: AvailableDatesRange(ctx,
    # &rpc.AvailableDatesRangeRequest{}) in lk-api.

  Scenario: boundary — the request is built in a variable
    Given "req := orderpb.CreateOrderRequest{Name: name}" and "o.client.CreateOrder(ctx, &req)"
    When the gidmappedfields analyzer checks the file
    Then the diagnostic "GID-266: the gRPC call CreateOrder carries no MappedFields option" is reported on the call
    # What the variable holds is not followed: the call stays judged. This is the
    # shape most of lk-api is written in.

  Scenario: boundary — an ordinary variadic call
    Given "join(in.Name, \"order\")" with a "...string" parameter
    When the gidmappedfields analyzer checks the file
    Then no diagnostic is reported

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — the module owns a data layer
    Given the same unmapped call in /domain/service of a module holding internal/dal
    When the gidmappedfields analyzer checks the file
    Then no diagnostic is reported
    # A service with a data layer reaches another service through a repository
    # (GID-160); it is not a BFF.

  Scenario: non-applicability — a library module
    Given the same unmapped call in a module with no composition root and no data layer
    When the gidmappedfields analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — the transport layer of a BFF
    Given the same unmapped call in /server/http
    When the gidmappedfields analyzer checks the file
    Then no diagnostic is reported
    # The judged layers are /domain/service and /domain/usecase, as in GID-160.

  Scenario: non-applicability — a _test.go file
    Given a test calling the double of a client interface without any option
    When the gidmappedfields analyzer checks the file
    Then no diagnostic is reported
    # The double repeats the signatures of the interface it fakes (GID-250 keeps
    # it in the same package), and the test passes the arguments its assertion
    # needs.

  Scenario: non-applicability — the RPC is named in settings.exclude
    Given settings.exclude holding "DeleteOrder" and "OrderServiceClient.Order"
    When the gidmappedfields analyzer checks the file
    Then no diagnostic is reported for those two calls
    And the sibling "CreateOrder" call is still reported
