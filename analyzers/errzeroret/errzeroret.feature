# language: en

Feature: GID-243 — on error, non-error results must be nil/zero
  As the styleguide owner
  I want a function that returns a non-nil error to also return nil/zero for every other result
  So that a caller checking only `if err != nil` never observes a partially populated, misleading value

  # Layer: go/analysis (package errzeroret, linter giderrzeroret), LoadModeTypesInfo.
  # No settings, no exceptions — the rule is absolute (owner's decision).
  #
  # Detect: a return with >=2 results whose LAST result is a DEFINITELY non-nil error:
  #   (a) the operand is a constructing call — status.Error/Errorf (grpc/status),
  #       errors.New/Wrap/Wrapf/Errorf/WithStack/WithMessage/WithMessagef (pkg/errors), fmt.Errorf; or
  #   (b) the return is lexically inside an `if <e> != nil { ... }` block guarding that same <e>.
  # → at least one non-error result must be zero: nil / false / a zero basic literal (0, 0.0, "")
  # / an empty composite literal T{}. A variable, a populated literal, or &T{} (a non-nil pointer —
  # the pointer zero VALUE is nil, not an address) are NOT zero.

  # --- Class 1: positive (the violation is caught) ---

  Scenario: positive (a) — a constructing error call alongside a populated (non-zero) result
    Given the function "func badConstructing() (*pb.Resp, error) { return &pb.Resp{}, status.Error(codes.Internal, \"x\") }"
    When the giderrzeroret analyzer checks the file
    Then the diagnostic "GID-243: on error, non-error results must be nil/zero (got a non-zero value alongside a non-nil error). Fix: return nil / T{} alongside the error" is reported on the return statement
    # &pb.Resp{} is a non-nil pointer (a populated address-of) — not zero, even though the struct
    # literal itself is empty: the zero VALUE of a pointer is nil, not the address of an empty struct.

  Scenario: positive (b) — return inside `if err != nil`, a variable result
    Given the function "func badGuarded() (int, error) { res, err := call(); if err != nil { return res, err }; return res, nil }"
    When the giderrzeroret analyzer checks the file
    Then the diagnostic "GID-243: on error, non-error results must be nil/zero (got a non-zero value alongside a non-nil error) …" is reported on "return res, err"
    # err is guarded by the enclosing `if err != nil` — a definitely non-nil error; res is a plain
    # variable, not a zero literal.

  # --- Class 2: negative (clean code passes) ---

  Scenario: negative — nil alongside a constructing error
    Given the function "func goodNilConstructing() (*pb.Resp, error) { return nil, status.Error(codes.Internal, \"x\") }"
    When the giderrzeroret analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — an empty composite literal inside an if-guard
    Given the function "func goodZeroGuarded() (model.Result, error) { res, err := callResult(); if err != nil { return model.Result{}, err }; return res, nil }"
    When the giderrzeroret analyzer checks the file
    Then no diagnostic is reported
    # model.Result{} is an empty (zero) value-type composite literal — unlike &T{}, this IS the zero value.

  Scenario: negative — a nil error alongside a variable result
    Given the function "func goodNilErr() (int, error) { res := 42; return res, nil }"
    When the giderrzeroret analyzer checks the file
    Then no diagnostic is reported
    # The error operand is the nil literal — never "definitely non-nil" — so the rule does not apply at all.

  # --- Class 3: boundary ---

  Scenario: boundary — an unconditional final forward is not flagged even though err could be non-nil
    Given the function "func forward(req int) (int, error) { resp, err := handler(req); return resp, err }"
    When the giderrzeroret analyzer checks the file
    Then no diagnostic is reported
    # err here is a plain variable: not a constructing call, and this return is not lexically inside an
    # `if err != nil` guard — it is the legitimate interceptor/pass-through shape and is deliberately exempt.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — a single-result return (no error in the signature shape checked here)
    Given the function "func single() error { return status.Error(codes.Internal, \"x\") }"
    When the giderrzeroret analyzer checks the file
    Then no diagnostic is reported
    # Fewer than 2 results — the rule only concerns multi-result returns.

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-243)
#  [x] Layer chosen: go/analysis (package errzeroret: giderrzeroret)
#  [x] Message is defined ("GID-243: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml
