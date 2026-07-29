# language: en

Feature: GID-122 — nullable entity fields use sql.Null*, not pointers (gidsqlnull)
  As a developer
  I want a nullable DAL field described by a database/sql type
  So that NULL is a value the scanner understands, not a pointer every caller
  has to nil-check before use

  # One analyzer over struct declarations, LoadModeTypesInfo (the field type is
  # resolved, so a named type and an alias are handled).
  # Scope: the root of /dal/entity, matched through internal/pathseg.EndsWith —
  # a subpackage such as /dal/entity/filter is out of scope, its pointers mean
  # "not set", not "NULL".
  # Trigger: a pointer field. The diagnostic names the sql type that fits the
  # element: time.Time -> sql.NullTime, string -> sql.NullString,
  # int32 -> sql.NullInt32, int/int64 -> sql.NullInt64,
  # float -> sql.NullFloat64, bool -> sql.NullBool, anything else (including a
  # struct) -> the generic sql.Null[T].
  # Generated code (ast.IsGenerated) is skipped.

  # --- Class 1: positive ---

  Scenario: positive — a pointer to time.Time
    Given the entity field "CompletedAt *time.Time"
    When the gidsqlnull analyzer checks the file
    Then the diagnostic "GID-122: a nullable entity field must use sql.NullTime, not a pointer. Fix: replace the pointer with it" is reported on the field

  Scenario: positive — a pointer to a string
    Given the entity field "Description *string"
    When the gidsqlnull analyzer checks the file
    Then the diagnostic "GID-122: a nullable entity field must use sql.NullString, not a pointer. …" is reported

  Scenario: positive — the integer widths are told apart
    Given the entity fields "FileCount *int32" and "Size *int64"
    When the gidsqlnull analyzer checks the file
    Then "sql.NullInt32" is suggested for the first and "sql.NullInt64" for the second

  # --- Class 2: negative ---

  Scenario: negative — the sql types themselves
    Given the entity fields "CompletedAt sql.NullTime" and "Description sql.NullString"
    When the gidsqlnull analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — a non-nullable field
    Given the entity fields "ID string" and "CreatedAt time.Time"
    When the gidsqlnull analyzer checks the file
    Then no diagnostic is reported

  # --- Class 3: boundary ---

  Scenario: boundary — a pointer to a custom type
    Given the entity field "Custom *MyType" where MyType is a struct
    When the gidsqlnull analyzer checks the file
    Then the diagnostic "GID-122: a nullable entity field must use sql.Null[T], not a pointer. …" is reported
    # The generic sql.Null[T] covers what the named Null* types do not.

  Scenario: boundary — a subpackage of the entity layer
    Given the field "Status *string" in /dal/entity/filter
    When the gidsqlnull analyzer checks the file
    Then no diagnostic is reported
    # The scope is the entity root; a filter's pointer means "no filter set".

  Scenario: boundary — a slice field
    Given the entity field "Tags []string"
    When the gidsqlnull analyzer checks the file
    Then no diagnostic is reported
    # A slice is nil-able on its own; only a pointer field is the shape checked.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — a model struct
    Given the field "CompletedAt *time.Time" in /domain/model
    When the gidsqlnull analyzer checks the file
    Then no diagnostic is reported
    # database/sql types belong to the DAL; the model has its own vocabulary.

  Scenario: non-applicability — generated code
    Given a file carrying the "Code generated … DO NOT EDIT." header
    When the gidsqlnull analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-122)
#  [x] Layer chosen: go/analysis (package sqlnull)
#  [x] Message is defined ("GID-122: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
