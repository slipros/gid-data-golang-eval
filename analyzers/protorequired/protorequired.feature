Feature: GID-232 — no NewRequired on proto3 enum fields (proto-enum-required)
  As a developer
  I want validator.NewRequired() to be forbidden on proto3 enum fields
  So that the enum zero value (*_UNSPECIFIED = 0) is actually rejected:
  the validator library treats it as non-empty (String() returns
  "..._UNSPECIFIED"), so NewRequired silently passes — NewInRange with the
  allowed values must be used instead

  # Scope: only packages with a "validate" path segment.
  # The validated struct is resolved from the enclosing Validate(ctx, req)
  # method or from a constructor of a type that has such a method.
  # A proto3 enum is a named int32 with String() string and
  # EnumDescriptor()/Descriptor() methods (generated protobuf code).

  Scenario: positive — NewRequired on a proto3 enum field
    Given a validate package with rules 'validator.RuleSet{"Executor": {validator.NewRequired()}}'
    And the validated request has field "Executor" of a proto3 enum type
    When the analyzer checks the file
    Then a "GID-232" diagnostic is reported on the NewRequired call

  Scenario: positive — chained NewRequired().When(...) on a proto3 enum field
    Given rules '"Executor": {validator.NewRequired().When(cond)}'
    When the analyzer checks the file
    Then a "GID-232" diagnostic is reported on the NewRequired call

  Scenario: positive — NewRequired inside NewNested for a nested message enum
    Given rules '"Input": {validator.NewNested(validator.RuleSet{"Source": {validator.NewRequired()}})}'
    And field "Source" of the nested message is a proto3 enum
    When the analyzer checks the file
    Then a "GID-232" diagnostic is reported on the inner NewRequired call

  Scenario: positive — NewRequired inside NewEach(NewNested(...)) for a repeated message
    Given rules '"Stages": {validator.NewEach(validator.NewNested(validator.RuleSet{"Executor": {validator.NewRequired()}}))}'
    And field "Executor" of the repeated message element is a proto3 enum
    When the analyzer checks the file
    Then a "GID-232" diagnostic is reported on the inner NewRequired call

  Scenario: negative — NewInRange on a proto3 enum field is the mandated fix
    Given rules '"Status": {validator.NewInRange([]any{pb.Status_ACTIVE, pb.Status_CLOSED})}'
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative (boundary) — NewRequired on a string field
    Given rules '"Name": {validator.NewRequired()}' where "Name" is a string
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative (boundary) — NewRequired on a pointer enum (proto3 optional)
    Given rules '"OptionalExecutor": {validator.NewRequired()}' where the field type is "*pb.StageExecutor"
    When the analyzer checks the file
    Then no diagnostic is reported, because a nil pointer is genuinely empty

  Scenario: negative (boundary) — named int32 that is not a proto enum
    Given rules '"Priority": {validator.NewRequired()}' where "Priority" is a plain "type Priority int32" without String/EnumDescriptor
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — the validated struct cannot be resolved
    Given a RuleSet literal in a helper "func sharedRules() validator.RuleSet" with no Validate(ctx, req) context
    When the analyzer checks the file
    Then no diagnostic is reported (FP-safe skip)

  Scenario: non-applicability — the key is not a field of the request struct
    Given rules '"Unknown": {validator.NewRequired()}' where the request has no field "Unknown"
    When the analyzer checks the file
    Then no diagnostic is reported (FP-safe skip)

  Scenario: non-applicability — not a validate-layer package
    Given the same rules in a package "/domain/service"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — field listed in settings.exclude
    Given "CreateStageRequest.Executor" is listed in settings.exclude
    And rules '"Executor": {validator.NewRequired()}' in a validate package
    When the analyzer checks the file
    Then no diagnostic is reported

# --- Checklist for adding a new rule ---
#  [x] ID and description registered in RULES.md (wired by the parent task)
#  [x] Layer chosen: go/analysis (types needed — RuleSet identity, field types, enum methods)
#  [x] Severity and message defined
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml (wired by the parent task)
