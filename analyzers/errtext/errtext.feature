# language: en

Feature: GID-256 — the text of an error is not passed into an error constructor
  As the styleguide owner
  I want a cause to travel inside the error chain, not as a string inside another error's message
  So that errors.Is/errors.As still reach it downstream and the stack collected at the failure
  survives the trip. Flattening an error with .Error() keeps only its text: the chain is cut, the
  stack is dropped, and what arrives at the caller is a sentence nobody can branch on.

  # Layer: go/analysis (package errtext, linter giderrtext), LoadModeTypesInfo.
  # Config: settings.exclude ("Function" | "Type.Method").
  #
  # Detect: a call to an ERROR CONSTRUCTOR
  #   - github.com/pkg/errors: New, Errorf, Wrap, Wrapf, WithMessage, WithMessagef
  #   - fmt.Errorf
  # one of whose MESSAGE arguments is the text of an error:
  #   (a) <errExpr>.Error() written inline, or
  #   (b) an identifier whose only assignment in the function is such a call (msg := err.Error()) —
  #       the shape that hides the flattening one line above the constructor.
  # The FIRST argument of Wrap/Wrapf/WithMessage/WithMessagef is the error being wrapped, not a
  # message — it is never flagged.
  #
  # Why (incident 2026-08-06, ad-cabinet-connector):
  #   wrapped := errors.Wrapf(err, "confirm yandex audience segment %d", segmentID)
  #   return errors.WithMessage(ErrServerError, wrapped.Error())
  # The transport cause is flattened into the message of a package-level sentinel: errors.Is(err,
  # context.DeadlineExceeded) is false downstream, errors.As to *net.OpError is false, and the
  # stack Wrapf collected on the line above is discarded with the chain — what ships is a sentinel
  # whose only stack points at the package init (GID-177). The consumer repeated the shape from the
  # other side (domain/service/yandex_channel.go): msg := err.Error() before replacing err with a
  # domain sentinel, then errors.Wrap(err, msg) to carry the text across.
  #
  # The shape appears when BOTH a stable class and the cause have to go out, and a sentinel var
  # cannot carry a cause. The fix is a typed error that can:
  #   type ServerError struct { Err error }
  #   func (e *ServerError) Unwrap() error { return e.Err }
  #   return errors.Wrapf(&ServerError{Err: err}, "confirm yandex audience segment %d", segmentID)
  # — class by errors.As, cause in the chain, one stack collected at the failure.
  #
  # Out of scope: err.Error() anywhere that is not an error constructor — a log field, a gRPC
  # status message (status.Error(codes.Internal, err.Error())), a comparison, a struct field.
  # Generated code and _test.go are skipped.

  # --- Class 1: positive (the violation is caught) ---

  Scenario: positive — the cause is flattened into a sentinel's message
    Given the function "func (c *Client) Confirm(id int64) error { _, err := c.do(); if err != nil { wrapped := errors.Wrapf(err, \"confirm segment %d\", id); return errors.WithMessage(ErrServerError, wrapped.Error()) }; return nil }"
    When the giderrtext analyzer checks the file
    Then the diagnostic "GID-256: the text of an error is passed into an error constructor — .Error() flattens the error into a string, so errors.Is/errors.As can no longer reach the cause and the stack collected with it is dropped. Fix: put the cause in the chain (errors.Wrap(err, \"context\")); when a stable class is needed too, carry it in a typed error (type ServerError struct { Err error } with Unwrap), not in a sentinel's message" is reported on "wrapped.Error()"
    # The incident shape, both halves on one line: the Wrapf result is used only for its text.

  Scenario: positive — the flattening hides one line above, in a variable
    Given the function "func (s *Service) Deliver() error { err := s.client.Do(); if err != nil { msg := err.Error(); err = model.ErrRetryable; return errors.Wrap(err, msg) }; return nil }"
    When the giderrtext analyzer checks the file
    Then the diagnostic "GID-256: the text of an error is passed into an error constructor …" is reported on "msg"
    # The consumer half of the same incident: the text is carried across the sentinel replacement
    # in a local. Tracking the variable is what makes the rule see it.

  Scenario: positive — errors.New over a cause's text
    Given the function "func f(err error) error { return errors.New(err.Error()) }"
    When the giderrtext analyzer checks the file
    Then the diagnostic "GID-256: the text of an error is passed into an error constructor …" is reported

  Scenario: positive — fmt.Errorf with %s over a cause's text
    Given the function "func f(err error) error { return fmt.Errorf(\"do x: %s\", err.Error()) }"
    When the giderrtext analyzer checks the file
    Then the diagnostic "GID-256: the text of an error is passed into an error constructor …" is reported
    # %w with the error itself is the whole point of fmt.Errorf; %s of its text is the same
    # flattening in a different syntax.

  Scenario: positive — WithMessagef with the text among the format arguments
    Given the function "func f(err error, code int) error { return errors.WithMessagef(ErrServerError, \"status %d: %s\", code, err.Error()) }"
    When the giderrtext analyzer checks the file
    Then the diagnostic "GID-256: the text of an error is passed into an error constructor …" is reported

  # --- Class 2: negative (clean code passes) ---

  Scenario: negative — the cause is put in the chain
    Given the function "func (c *Client) Confirm(id int64) error { _, err := c.do(); if err != nil { return errors.Wrapf(err, \"confirm segment %d\", id) }; return nil }"
    When the giderrtext analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — a typed error carries the cause, Wrap collects the stack
    Given the function "func (c *Client) Confirm(id int64) error { _, err := c.do(); if err != nil { return errors.Wrapf(&ServerError{Err: err}, \"confirm segment %d\", id) }; return nil }"
    When the giderrtext analyzer checks the file
    Then no diagnostic is reported
    # The shape the rule points at: class by type, cause under Unwrap, one stack.

  Scenario: negative — err.Error() outside an error constructor (a gRPC status message)
    Given the function "func f(err error) *status.Status { return status.New(codes.Internal, err.Error()) }"
    When the giderrtext analyzer checks the file
    Then no diagnostic is reported
    # A status message IS a string by contract — nothing is being flattened into an error there.

  Scenario: negative — err.Error() in a log field
    Given the function "func f(err error) { log.Println(\"failed:\", err.Error()) }"
    When the giderrtext analyzer checks the file
    Then no diagnostic is reported

  # --- Class 3: boundary ---

  Scenario: boundary — the FIRST argument of Wrap is the error being wrapped, not a message
    Given the function "func f(err error) error { return errors.Wrap(err, \"ctx\") }"
    When the giderrtext analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — a variable assigned err.Error() but never handed to a constructor
    Given the function "func f(err error) string { msg := err.Error(); return msg }"
    When the giderrtext analyzer checks the file
    Then no diagnostic is reported
    # The rule is about what reaches an error constructor, not about calling .Error() at all.

  Scenario: boundary — a variable reassigned from something else is no longer error text
    Given the function "func f(err error, alt string) error { msg := err.Error(); msg = alt; return errors.New(msg) }"
    When the giderrtext analyzer checks the file
    Then no diagnostic is reported
    # Only a variable whose EVERY assignment is <err>.Error() counts — otherwise the value at the
    # constructor is unknown and the rule stays silent (no guessing).

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — a _test.go fixture rebuilds an expected error from a text
    Given the test helper "func wantErr(err error) error { return errors.New(err.Error()) }" in client_test.go
    When the giderrtext analyzer checks the file
    Then no diagnostic is reported
    # A test lives in the same package (GID-250) and legitimately constructs an error that only has
    # to LOOK like the production one — there is no chain for anyone to branch on downstream.

  Scenario: non-applicability — a constructor from another package
    Given the function "func f(err error) error { return gderror.New(err.Error()) }"
    When the giderrtext analyzer checks the file
    Then no diagnostic is reported
    # The rule names the constructors it knows (pkg/errors + fmt.Errorf); GID-146 keeps the code on
    # pkg/errors anyway.

  # --- Config: settings.exclude ---

  Scenario: config — an excluded function may flatten the error
    Given settings.exclude contains "Client.legacyConfirm" and that method flattening err into errors.New
    When the giderrtext analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-256)
#  [x] Layer chosen: go/analysis (package errtext: giderrtext)
#  [x] Message is defined ("GID-256: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
