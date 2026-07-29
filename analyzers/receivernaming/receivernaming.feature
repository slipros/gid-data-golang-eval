# language: en

Feature: GID-103 — a receiver is the type's first letter (gidreceiver)
  As a developer
  I want receivers named uniformly by the type's lowercase first letter
  So that reading any method tells the receiver from a parameter at a glance,
  without svc/this/self appearing at random

  # One analyzer over the file's method declarations, LoadModeTypesInfo (the
  # receiver type is resolved through the type info, so a pointer receiver and
  # a named slice type are both handled).
  # Expected name: the lowercase first letter of the type name; for a type
  # whose underlying type is a slice — that letter doubled
  # (type Snapshots []Snapshot -> ss).
  # An unnamed receiver and "_" are left alone: there is no name to check.
  # No exceptions by layer — validate and handler packages included.
  # Generated code (ast.IsGenerated) is skipped.

  # --- Class 1: positive ---

  Scenario: positive — a letter that is not the type's
    Given "func (h *Snapshot) Get() string"
    When the gidreceiver analyzer checks the file
    Then the diagnostic "GID-103: receiver of type Snapshot is named \"s\". Fix: use the lowercase first letter of the type (two for slice types), got \"h\"" is reported on the receiver

  Scenario: positive — an abbreviation instead of a letter
    Given "func (svc *Snapshot) Bad() string"
    When the gidreceiver analyzer checks the file
    Then the diagnostic "GID-103: receiver of type Snapshot is named \"s\". … got \"svc\"" is reported

  Scenario: positive — this as a receiver
    Given "func (this *Snapshot) Worse() string"
    When the gidreceiver analyzer checks the file
    Then the diagnostic "GID-103: receiver of type Snapshot is named \"s\". … got \"this\"" is reported

  Scenario: positive — a validate package is not exempt
    Given "func (v *CreateSnapshot) Validate() error"
    When the gidreceiver analyzer checks the file
    Then the diagnostic "GID-103: receiver of type CreateSnapshot is named \"c\". … got \"v\"" is reported
    # The rule applies uniformly; "v for validator" is not an exception.

  # --- Class 2: negative ---

  Scenario: negative — the first letter, pointer receiver
    Given "func (s *Snapshot) Name() string"
    When the gidreceiver analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — a slice type with the doubled letter
    Given "type Snapshots []Snapshot" and "func (ss Snapshots) IDs() []string"
    When the gidreceiver analyzer checks the file
    Then no diagnostic is reported

  # --- Class 3: boundary ---

  Scenario: boundary — a slice type with a single letter
    Given "func (s Snapshots) IDs() []string"
    When the gidreceiver analyzer checks the file
    Then the diagnostic "GID-103: receiver of type Snapshots is named \"ss\". … got \"s\"" is reported

  Scenario: boundary — an unnamed receiver
    Given "func (*Snapshot) Noop()" and "func (_ *Snapshot) Skip()"
    When the gidreceiver analyzer checks the file
    Then no diagnostic is reported
    # Nothing is named, so there is nothing to rename.

  Scenario: boundary — a value receiver
    Given "func (s Snapshot) Name() string"
    When the gidreceiver analyzer checks the file
    Then no diagnostic is reported
    # Pointer or value makes no difference to the name.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — a receiverless function
    Given "func normalize(s string) string"
    When the gidreceiver analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — generated code
    Given a file carrying the "Code generated … DO NOT EDIT." header
    When the gidreceiver analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-103)
#  [x] Layer chosen: go/analysis (package receivernaming)
#  [x] Message is defined ("GID-103: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
