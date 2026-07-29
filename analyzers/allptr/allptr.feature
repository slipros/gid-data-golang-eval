# language: en

Feature: GID-004 — iterate over a slice of structs through gdhelper.AllPtr (gidallptr)
  As a developer
  I want a range over a slice of structs to go through gdhelper.AllPtr
  So that each iteration takes a pointer instead of copying the whole element

  # One analyzer over RangeStmt nodes, LoadModeTypesInfo (the ranged expression
  # is resolved by type, not by name).
  # Trigger: the type of the ranged expression is a slice whose element type is
  # a struct underneath. Slices of pointers copy nothing and never count.
  # Correct code — "for _, v := range gdhelper.AllPtr(items)" — is not flagged
  # for free: AllPtr returns an iterator (range-over-func), not a slice, so it
  # fails the slice check on its own.
  # The helper lives in gitlab.gid.team/gid-data/tech/golang/libs/helper.git and
  # is named in the diagnostic. Generated code (ast.IsGenerated) is skipped.

  # --- Class 1: positive ---

  Scenario: positive — a range over a slice of structs
    Given "files []File" where File is a struct and the loop "for _, f := range files"
    When the gidallptr analyzer checks the file
    Then the diagnostic "GID-004: ranging over a slice of structs copies each element. Fix: range over gdhelper.AllPtr(items) (gitlab.gid.team/gid-data/tech/golang/libs/helper.git) to iterate pointers." is reported on the ranged expression

  Scenario: positive — a named slice type
    Given "type Files []File" and the loop "for _, f := range files"
    When the gidallptr analyzer checks the file
    Then the diagnostic "GID-004: …" is reported
    # The check goes through the underlying type, so a named slice counts too.

  # --- Class 2: negative ---

  Scenario: negative — a range over gdhelper.AllPtr
    Given the loop "for _, f := range gdhelper.AllPtr(files)"
    When the gidallptr analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — a slice of pointers
    Given "files []*File" and the loop "for _, f := range files"
    When the gidallptr analyzer checks the file
    Then no diagnostic is reported
    # []*T iterates pointers already — there is nothing to copy.

  # --- Class 3: boundary ---

  Scenario: boundary — only the index is taken
    Given "files []File" and the loop "for i := range files"
    When the gidallptr analyzer checks the file
    Then the diagnostic "GID-004: …" is reported
    # Deliberate: the ranged expression is what the rule looks at, and the loop
    # body is free to add ", f" at any time.

  Scenario: boundary — a slice of a scalar type
    Given "ids []string" and the loop "for _, id := range ids"
    When the gidallptr analyzer checks the file
    Then no diagnostic is reported
    # Copying a string header is not the cost the rule is about.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — a range over a map, a channel or an integer
    Given the loops "for k := range m", "for v := range ch" and "for i := range 10"
    When the gidallptr analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — generated code
    Given a file carrying the "Code generated … DO NOT EDIT." header
    When the gidallptr analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-004)
#  [x] Layer chosen: go/analysis (package allptr)
#  [x] Message is defined ("GID-004: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
