# language: en

Feature: GID-101 — no Get prefix in method names (gidnogetprefix)
  As a developer
  I want a value-returning method to be named after the value
  So that call sites read as job.ID() rather than job.GetID()

  # One analyzer over the file's top-level function declarations, no type info
  # needed.
  # Trigger: a method (a function with a receiver) whose name starts with the
  # word Get — either the bare name "Get", or "Get" followed by an uppercase
  # letter or a digit. A name where "get" merely begins another word
  # ("Getaway") is not a violation.
  # The fix is spelled out in the diagnostic: the name with the prefix trimmed.
  # Receiverless functions are out of scope — the rule ports the method naming
  # section of the styleguide.
  # Generated code (ast.IsGenerated) is skipped: in protobuf stubs and the like
  # the Get prefix is part of the contract.

  # --- Class 1: positive ---

  Scenario: positive — Get before an uppercase word
    Given "func (j *Job) GetID() string"
    When the gidnogetprefix analyzer checks the file
    Then the diagnostic "GID-101: method \"GetID\" uses the Get prefix. Fix: name getters without it: \"ID\"" is reported on the method name

  Scenario: positive — a longer getter name
    Given "func (j *Job) GetStatus() string"
    When the gidnogetprefix analyzer checks the file
    Then the diagnostic "GID-101: method \"GetStatus\" uses the Get prefix. Fix: name getters without it: \"Status\"" is reported

  # --- Class 2: negative ---

  Scenario: negative — the method is named after the value
    Given "func (j *Job) ID() string" and "func (j *Job) Status() string"
    When the gidnogetprefix analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — Get is part of another word
    Given "func (j *Job) Getaway() string"
    When the gidnogetprefix analyzer checks the file
    Then no diagnostic is reported

  # --- Class 3: boundary ---

  Scenario: boundary — the bare name Get
    Given "func (j *Job) Get() string"
    When the gidnogetprefix analyzer checks the file
    Then the diagnostic "GID-101: method \"Get\" uses the Get prefix. …" is reported

  Scenario: boundary — Get followed by a digit
    Given "func (j *Job) Get2FA() string"
    When the gidnogetprefix analyzer checks the file
    Then the diagnostic "GID-101: method \"Get2FA\" uses the Get prefix. …" is reported
    # A digit starts a word just as an uppercase letter does.

  Scenario: boundary — a receiverless function with the Get prefix
    Given "func GetConfig() Config"
    When the gidnogetprefix analyzer checks the file
    Then no diagnostic is reported

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — generated protobuf accessors
    Given a file carrying the "Code generated … DO NOT EDIT." header and holding "func (m *Job) GetId() string"
    When the gidnogetprefix analyzer checks the file
    Then no diagnostic is reported
    # There the Get prefix is the generated contract, not a naming choice.

  Scenario: non-applicability — a method named in lowercase
    Given "func (j *Job) getID() string"
    When the gidnogetprefix analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-101)
#  [x] Layer chosen: go/analysis (package nogetprefix)
#  [x] Message is defined ("GID-101: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
