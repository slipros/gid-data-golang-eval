# language: en

Feature: GID-165/166/167 — context keys and helpers live in the model (gidctxkeys)
  As a developer
  I want everything that goes into a context declared in /domain/model
  So that a business layer reads request data without importing the middleware
  that put it there, and keys cannot collide

  # One analyzer carrying three rules, LoadModeTypesInfo (context.WithValue and
  # the key types are resolved through the type info).
  # The package layer splits the behaviour: inside /domain/model
  # (pathseg.HasLayer) the helpers and keys are checked; everywhere else only
  # GID-165 applies.
  # GID-165 — context.WithValue outside /domain/model is forbidden.
  # GID-166 — the shape of the helpers in the model: a function whose body
  # calls context.WithValue must be exported and named ContextWith<Name>; one
  # that reads ctx.Value must be exported and named <Name>FromContext; and the
  # helper lives in the file where its <Name> entity is declared.
  # GID-167 — the key is the exported named string type ContextKey; its const
  # values are snake_case strings declared in the same file as the type. The
  # suggested snake_case spelling is computed and shown in the diagnostic.
  # Generated code (ast.IsGenerated) is skipped.

  # Scope: only a module laid out as a service — internal/modlayout walks up to
  # the package's go.mod and looks for a layer directory (domain, dal, server,
  # app, usecase, repository) at the module root or under internal/. A flat
  # library module is skipped.

  # --- Class 1: positive ---

  Scenario: positive — WithValue in an http middleware
    Given "ctx := context.WithValue(r.Context(), contextKey(\"user\"), \"id\")" in /server/http/middleware
    When the gidctxkeys analyzer checks the file
    Then the diagnostic "GID-165: context.WithValue outside /domain/model is forbidden. Fix: keep context keys and helpers in /domain/model so business layers do not depend on middleware" is reported

  Scenario: positive — a writer helper named the wrong way
    Given "func WithUserID(ctx context.Context, id string) context.Context" in /domain/model
    When the gidctxkeys analyzer checks the file
    Then the diagnostic "GID-166: function \"WithUserID\" stores data in ctx. Fix: make it public and name it ContextWith<Name>" is reported

  Scenario: positive — an unexported writer helper
    Given "func contextWithTrace(ctx context.Context, id string) context.Context" in /domain/model
    When the gidctxkeys analyzer checks the file
    Then the diagnostic "GID-166: function \"contextWithTrace\" stores data in ctx. …" is reported
    # The prefix is checked case-sensitively: the helper is public API.

  Scenario: positive — a reader helper named the wrong way
    Given "func GetUserID(ctx context.Context) (string, bool)" in /domain/model
    When the gidctxkeys analyzer checks the file
    Then the diagnostic "GID-166: function \"GetUserID\" reads data from ctx. Fix: make it public and name it <Name>FromContext" is reported

  Scenario: positive — a raw key in WithValue
    Given "return context.WithValue(ctx, \"raw\", s)" in /domain/model
    When the gidctxkeys analyzer checks the file
    Then the diagnostic "GID-167: context key must be the public type ContextKey (type ContextKey string), not a raw value. Fix: declare type ContextKey string and use its typed constants" is reported

  Scenario: positive — a private key type
    Given "return context.WithValue(ctx, secretKey(\"secret\"), s)" in /domain/model
    When the gidctxkeys analyzer checks the file
    Then the diagnostic "GID-167: context key must be the public type ContextKey (type ContextKey string), not \"secretKey\". …" is reported

  Scenario: positive — a key value that is not snake_case
    Given the constants "BadCamelKey ContextKey = \"UserID\"" and "BadDashKey ContextKey = \"user-id\""
    When the gidctxkeys analyzer checks the file
    Then "GID-167: ContextKey value must be a snake_case string, got \"UserID\". Fix: use \"user_id\"" and the same for "user-id" are reported

  Scenario: positive — a key constant away from its type
    Given "const LegacySessionKey ContextKey = \"legacy_session\"" declared in a file other than the one holding "type ContextKey string"
    When the gidctxkeys analyzer checks the file
    Then the diagnostic "GID-167: ContextKey values must be declared next to the ContextKey type declaration (same file)" is reported

  Scenario: positive — a helper away from its entity
    Given "func TokenFromContext(ctx context.Context) (Token, bool)" declared in a file other than the one holding "type Token struct"
    When the gidctxkeys analyzer checks the file
    Then the diagnostic "GID-166: helper \"TokenFromContext\" must live in the same file as the \"Token\" entity it stores into / reads from ctx" is reported

  # --- Class 2: negative ---

  Scenario: negative — the canonical pair in the model
    Given "func ContextWithSession(ctx context.Context, s Session) context.Context" and "func SessionFromContext(ctx context.Context) (Session, bool)" in session.go, next to "type Session struct"
    When the gidctxkeys analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — typed snake_case keys next to the type
    Given "type ContextKey string" and "const SessionKey ContextKey = \"session\"" in the same file
    When the gidctxkeys analyzer checks the file
    Then no diagnostic is reported

  # --- Class 3: boundary ---

  Scenario: boundary — a reader helper named FromContextLegacy
    Given "func JobFromContextLegacy(ctx context.Context) (Job, bool)"
    When the gidctxkeys analyzer checks the file
    Then the diagnostic "GID-166: function \"JobFromContextLegacy\" reads data from ctx. …" is reported
    # FromContext must end the name, not sit in the middle of it.

  Scenario: boundary — a model nested under another layer
    Given "context.WithValue(ctx, key, v)" in svc/client/domain/model
    When the gidctxkeys analyzer checks the file
    Then the GID-165 diagnostic is reported
    # The model layer is anchored to the module root: a nested domain/model is
    # not the model, so storing into ctx there is the forbidden case.

  Scenario: boundary — a helper whose entity is not a struct of the package
    Given "func UserIDFromContext(ctx context.Context) (string, bool)" where no "UserID" struct is declared
    When the gidctxkeys analyzer checks the file
    Then no file-placement diagnostic is reported
    # There is no entity file to be next to.

  Scenario: boundary — reading ctx.Value outside the model
    Given "v := ctx.Value(key)" in /server/http/middleware
    When the gidctxkeys analyzer checks the file
    Then no diagnostic is reported
    # GID-165 bans putting data in; reading what the model exposes is what the
    # <Name>FromContext helper does for the caller.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — a function taking ctx but touching neither WithValue nor Value
    Given "func (s *Snapshot) Do(ctx context.Context) error" in /domain/model
    When the gidctxkeys analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — generated code
    Given a file carrying the "Code generated … DO NOT EDIT." header
    When the gidctxkeys analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — a library module
    Given "context.WithValue(ctx, txKey, tx)" in the root package of libs/postgres
    When the gidctxkeys analyzer checks the file
    Then no diagnostic is reported
    # A library keeps its context helpers next to the API they serve
    # (postgres.ContextWithTx): there is no /domain/model to move them into.

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-165, GID-166, GID-167)
#  [x] Layer chosen: go/analysis (package ctxkeys)
#  [x] Messages are defined ("GID-165: …", "GID-166: …", "GID-167: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
