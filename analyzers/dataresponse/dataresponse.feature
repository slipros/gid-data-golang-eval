# language: en

Feature: GID-163 — http handlers are built on data-response.go (giddataresponse)
  As a developer
  I want handlers written against github.com/raoptimus/data-response.go/v2
  So that responses and errors are serialised by one library instead of every
  handler writing to http.ResponseWriter its own way

  # One analyzer over the file's function declarations, LoadModeTypesInfo (both
  # http types are matched by type).
  # Scope: the /server/http layer, matched through internal/pathseg.HasLayer —
  # anchored to the module root.
  # Trigger: a plain golang handler — exactly the signature
  # (http.ResponseWriter, *http.Request) with no results. A function that takes
  # the pair but returns something is a different shape and belongs to GID-162.
  # Exceptions: //nolint:giddataresponse pointwise, settings.exclude centrally,
  # in the "Function" or "Type.Method" form (internal/exclude).
  # Generated code (ast.IsGenerated) is skipped.

  # --- Class 1: positive ---

  Scenario: positive — a handler method with the plain signature
    Given "func (h *Snapshot) Get(w http.ResponseWriter, r *http.Request)" in /server/http/handler
    When the giddataresponse analyzer checks the file
    Then the diagnostic "GID-163: \"Get\" is a plain golang handler, which is forbidden. Fix: use github.com/raoptimus/data-response.go/v2 (exceptions: nolint or settings.exclude)" is reported on the function name

  Scenario: positive — a package-level handler function
    Given "func Ready(w http.ResponseWriter, r *http.Request)"
    When the giddataresponse analyzer checks the file
    Then the diagnostic "GID-163: \"Ready\" is a plain golang handler, which is forbidden. …" is reported

  # --- Class 2: negative ---

  Scenario: negative — a handler built on the library
    Given "func (h *Snapshot) Get(ctx context.Context, req GetRequest) (dataresponse.Data, error)"
    When the giddataresponse analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — the handler is listed in settings.exclude
    Given settings "exclude: [Snapshot.Get]" and that method with the plain signature
    When the giddataresponse analyzer checks the file
    Then no diagnostic is reported

  # --- Class 3: boundary ---

  Scenario: boundary — the plain signature with a result
    Given "func (h *Snapshot) Get(w http.ResponseWriter, r *http.Request) error"
    When the giddataresponse analyzer checks the file
    Then no diagnostic is reported
    # A returned error is GID-162's finding; this rule matches the exact
    # resultless shape.

  Scenario: boundary — middleware taking the pair plus a handler
    Given "func Log(w http.ResponseWriter, r *http.Request, next http.Handler)"
    When the giddataresponse analyzer checks the file
    Then no diagnostic is reported
    # Exactly two parameters make a handler; a third one makes it something else.

  Scenario: boundary — an http layer nested under another layer
    Given the plain handler declared in svc/client/x/server/http/handler
    When the giddataresponse analyzer checks the file
    Then no diagnostic is reported
    # The layer is anchored to the module root.

  Scenario: boundary — exclude by function name only
    Given settings "exclude: [Ready]" and the package-level "func Ready(w http.ResponseWriter, r *http.Request)"
    When the giddataresponse analyzer checks the file
    Then no diagnostic is reported
    # internal/exclude accepts both "Function" and "Type.Method".

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — a package outside the http layer
    Given the plain handler declared in /pkg/middleware
    When the giddataresponse analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — generated code
    Given a file carrying the "Code generated … DO NOT EDIT." header
    When the giddataresponse analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-163)
#  [x] Layer chosen: go/analysis (package dataresponse)
#  [x] Message is defined ("GID-163: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
