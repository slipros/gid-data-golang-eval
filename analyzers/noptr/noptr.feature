# language: en

Feature: GID-120/GID-121 — no pointers where a zero value already says "empty" (gidnoptr)
  As a developer
  I want uuid and simple model fields kept as values
  So that reading a field never starts with a nil check for absence the type
  can express by itself

  # One analyzer carrying two rules, LoadModeTypesInfo.
  # GID-120 — *uuid.UUID is forbidden in any type position: a struct field, a
  # parameter, a result, a local declaration. The uuid type is matched through
  # internal/pathseg (SameLibrary/Contains), so a vendored or version-suffixed
  # gofrs/uuid counts. A StarExpr that is a dereference, not a type, is
  # skipped — the type info tells them apart.
  # GID-121 — in /domain/model and /event/dto (pathseg.HasLayer, anchored to
  # the module root) a struct field must not be a pointer to a simple type:
  # *time.Time, or a pointer to a basic numeric or string type, including
  # named types based on one. The zero value already answers the question:
  # t.IsZero(), len(s) == 0, == 0.
  # Exempt on purpose: *bool (false is a meaningful value of its own) and a
  # pointer to a nested struct.
  # Escape hatch: //nolint:gidnoptr where a pointer is genuinely needed.
  # Generated code (ast.IsGenerated) is skipped.

  # --- Class 1: positive ---

  Scenario: positive — *uuid.UUID as a model field
    Given the model field "ParentID *uuid.UUID"
    When the gidnoptr analyzer checks the file
    Then the diagnostic "GID-120: *uuid.UUID is forbidden. Fix: use uuid.UUID and check emptiness with IsNil()" is reported

  Scenario: positive — *uuid.UUID as a function parameter outside the model
    Given "func Lookup(id *uuid.UUID) bool" in /dal/repository
    When the gidnoptr analyzer checks the file
    Then the diagnostic "GID-120: *uuid.UUID is forbidden. …" is reported
    # GID-120 holds everywhere, not only in the model.

  Scenario: positive — *time.Time in the model
    Given the model field "CompletedAt *time.Time"
    When the gidnoptr analyzer checks the file
    Then the diagnostic "GID-121: *time.Time is unnecessary here. Fix: use time.Time and check absence with t.IsZero(); if a pointer is unavoidable, use //nolint:gidnoptr" is reported

  Scenario: positive — pointers to simple types in the model
    Given the model fields "Description *string", "Count *int" and "Ratio *float64"
    When the gidnoptr analyzer checks the file
    Then "GID-121: a pointer to a simple type is unnecessary here. Fix: use the value and check the zero value (len(s) == 0 for strings, == 0 for numbers); …" is reported for each

  Scenario: positive — the event dto layer is in scope too
    Given the fields "CompletedAt *time.Time" and "Count *int" in /event/dto
    When the gidnoptr analyzer checks the file
    Then the GID-121 diagnostics are reported

  # --- Class 2: negative ---

  Scenario: negative — value types in the model
    Given the fields "ID uuid.UUID", "CompletedAt time.Time" and "Count int"
    When the gidnoptr analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — a pointer to a simple type outside the model layers
    Given the field "Description *string" in /dal/repository
    When the gidnoptr analyzer checks the file
    Then no diagnostic is reported
    # GID-121 is scoped to model and event/dto; the DAL has GID-122 for its
    # nullable fields.

  # --- Class 3: boundary ---

  Scenario: boundary — a named type based on a simple one
    Given the model fields "Status *SnapshotStatus" and "Priority *Weight" where both are based on string and int
    When the gidnoptr analyzer checks the file
    Then the "GID-121: a pointer to a simple type …" diagnostic is reported for each
    # The underlying type decides, not the name.

  Scenario: boundary — *bool
    Given the model field "Enabled *bool"
    When the gidnoptr analyzer checks the file
    Then no diagnostic is reported
    # Deliberate: false is a value in its own right, so the pointer carries
    # information the zero value cannot.

  Scenario: boundary — a pointer to a nested struct
    Given the model field "Owner *User"
    When the gidnoptr analyzer checks the file
    Then no diagnostic is reported
    # A struct pointer is about ownership and size, not about absence.

  Scenario: boundary — a model nested under another layer
    Given the fields with pointers in svc/server/grpc/domain/model
    When the gidnoptr analyzer checks the file
    Then no GID-121 diagnostic is reported
    # The layer is anchored to the module root. GID-120 still applies there:
    # *uuid.UUID is banned regardless of the package.

  Scenario: boundary — a pointer dereference spelled *p
    Given the statement "v := *ptr" where ptr is a *uuid.UUID variable
    When the gidnoptr analyzer checks the file
    Then no diagnostic is reported on the dereference
    # Only a StarExpr in a type position is a type; the declaration of ptr
    # itself is where GID-120 fires.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — a uuid type from another library
    Given the model field "ParentID *otheruuid.UUID" from a package that is not gofrs/uuid
    When the gidnoptr analyzer checks the file
    Then no GID-120 diagnostic is reported
    # Importing it is GID-137's finding, not this rule's.

  Scenario: non-applicability — generated code
    Given a file carrying the "Code generated … DO NOT EDIT." header
    When the gidnoptr analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-120, GID-121)
#  [x] Layer chosen: go/analysis (package noptr)
#  [x] Messages are defined ("GID-120: …", "GID-121: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
