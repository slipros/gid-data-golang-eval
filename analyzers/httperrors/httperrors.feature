# language: en

Feature: GID-162 — an http handler handles its own errors (gidhttperrors)
  As a developer
  I want each handler to turn its own error into a response on the spot
  So that the status and the payload of a failure are read next to the code
  that produced it, instead of in one universal error sink

  # One analyzer over the file's function declarations, LoadModeTypesInfo (both
  # http.ResponseWriter and error are matched by type, not by name).
  # Scope: the /server/http layer, matched through internal/pathseg.HasLayer —
  # anchored to the module root.
  # Two shapes are reported:
  #   1. a super-method — a function taking both http.ResponseWriter and error:
  #      the marker of a universal handler everyone funnels errors into;
  #   2. a handler taking http.ResponseWriter and returning error: the error
  #      has nowhere to go but such a sink.
  # A function is reported once: shape 1 wins if both match.
  # Generated code (ast.IsGenerated) is skipped.

  # --- Class 1: positive ---

  Scenario: positive — an error-handling method
    Given "func (h *Snapshot) handleError(w http.ResponseWriter, err error)" in /server/http/handler
    When the gidhttperrors analyzer checks the file
    Then the diagnostic "GID-162: \"handleError\" is a forbidden error-handling super-method. Fix: handle errors inside each http handler" is reported on the function name

  Scenario: positive — a package-level error writer
    Given "func writeError(w http.ResponseWriter, status int, err error)"
    When the gidhttperrors analyzer checks the file
    Then the diagnostic "GID-162: \"writeError\" is a forbidden error-handling super-method. …" is reported
    # Extra parameters between the two markers change nothing.

  Scenario: positive — a handler returning an error
    Given "func (h *Snapshot) Get(w http.ResponseWriter, r *http.Request) error"
    When the gidhttperrors analyzer checks the file
    Then the diagnostic "GID-162: http handler \"Get\" must not return error. Fix: handle the error in place" is reported

  # --- Class 2: negative ---

  Scenario: negative — a handler writing the response itself
    Given "func (h *Snapshot) Get(w http.ResponseWriter, r *http.Request)" that renders the error inline
    When the gidhttperrors analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — a helper without http.ResponseWriter
    Given "func wrapErr(err error) error"
    When the gidhttperrors analyzer checks the file
    Then no diagnostic is reported
    # Without the writer there is no handler and no sink.

  # --- Class 3: boundary ---

  Scenario: boundary — a function matching both shapes
    Given "func render(w http.ResponseWriter, err error) error"
    When the gidhttperrors analyzer checks the file
    Then a single diagnostic is reported, the super-method one
    # One diagnostic per function: the super-method shape is the finding, the
    # returned error is its consequence.

  Scenario: boundary — an http layer nested under another layer
    Given "func (h *Snapshot) handleError(w http.ResponseWriter, err error)" in svc/job/server/http/handler
    When the gidhttperrors analyzer checks the file
    Then no diagnostic is reported
    # The layer is anchored to the module root, so this package is out of scope.

  Scenario: boundary — a handler returning a non-error value
    Given "func (h *Snapshot) Get(w http.ResponseWriter, r *http.Request) bool"
    When the gidhttperrors analyzer checks the file
    Then no diagnostic is reported
    # Only a returned error is the smell the rule is after.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — a package outside the http layer
    Given the same super-method declared in /domain/service
    When the gidhttperrors analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — generated code
    Given a file carrying the "Code generated … DO NOT EDIT." header
    When the gidhttperrors analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-162)
#  [x] Layer chosen: go/analysis (package httperrors)
#  [x] Message is defined ("GID-162: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
