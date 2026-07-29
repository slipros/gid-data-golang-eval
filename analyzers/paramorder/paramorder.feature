# language: en

Feature: GID-110/113/153/252 — parameter order: ctx, opts, logger, metrics, the rest (gidparamorder)
  As a developer
  I want the known parameters in one fixed order at the head of every signature
  So that reading a call tells the infrastructure apart from the entity's own
  dependencies without looking up the declaration

  # One analyzer carrying four rules over function and method declarations,
  # LoadModeTypesInfo. Parameters are flattened first, so "a, b string" counts
  # as two positions.
  # Classification by type: context.Context -> ctx; a named type with the
  # Options suffix (pointer or value) -> opts; a logger recognised by
  # internal/lgr (logrus *Entry/*Logger/FieldLogger, *slog.Logger) -> logger; a
  # named type with the Metric/Metrics suffix (the GID-174 shape) -> metrics;
  # everything else -> a dependency.
  # GID-110 — ctx is the first parameter.
  # GID-113 — opts comes right after ctx, or first when there is no ctx.
  # GID-153 — the logger comes after opts when both are present.
  # GID-252 — scoped to constructors (package-level New* functions): the known
  # parameters all precede the entity's own dependencies, and metrics follow
  # the logger. An ordinary function is out of scope — a helper may
  # legitimately take a logger last. GID-110/113/153 already cover ctx and
  # opts, so GID-252 reports only what they do not.
  # Generated code (ast.IsGenerated) is skipped.

  # --- Class 1: positive ---

  Scenario: positive — ctx is not first
    Given "func (h *Hello) BadCtx(id int, ctx context.Context) error"
    When the gidparamorder analyzer checks the file
    Then the diagnostic "GID-110: context.Context must be the first parameter. Fix: move ctx first" is reported

  Scenario: positive — opts last
    Given "func (h *Hello) BadOpts(ctx context.Context, id int, opts *HelloOptions) error"
    When the gidparamorder analyzer checks the file
    Then the diagnostic "GID-113: opts must come right after ctx, not last. Fix: move opts after ctx" is reported

  Scenario: positive — the logger before opts
    Given "func NewBad(logger *logrus.Entry, opts *HelloOptions) *Hello"
    When the gidparamorder analyzer checks the file
    Then "GID-113: opts must come right after ctx, not last. …" and "GID-153: logger must come after the entity opts. Fix: move logger after opts" are both reported

  Scenario: positive — the logger after the dependencies in a constructor
    Given "func NewInterceptor(opts InterceptorOptions, checker BanChecker, logger *slog.Logger) *Interceptor"
    When the gidparamorder analyzer checks the file
    Then the diagnostic "GID-252: the logger comes after the entity's own dependencies. Fix: order the parameters ctx, opts, logger, metrics, then the rest" is reported

  Scenario: positive — metrics after the dependencies in a constructor
    Given "func NewCollector(checker BanChecker, metrics *InterceptorMetrics) *Interceptor"
    When the gidparamorder analyzer checks the file
    Then the diagnostic "GID-252: metrics come after the entity's own dependencies. …" is reported

  Scenario: positive — metrics before the logger
    Given "func NewSwapped(metrics *InterceptorMetrics, logger *slog.Logger) *Interceptor"
    When the gidparamorder analyzer checks the file
    Then the diagnostic "GID-252: metrics come before the logger. …" is reported

  # --- Class 2: negative ---

  Scenario: negative — the canonical order
    Given "func NewHello(ctx context.Context, opts *HelloOptions, logger *logrus.Entry, metrics *HelloMetrics, repo Repository) *Hello"
    When the gidparamorder analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — a signature with none of the known parameters
    Given "func (h *Hello) Name(id int) string"
    When the gidparamorder analyzer checks the file
    Then no diagnostic is reported

  # --- Class 3: boundary ---

  Scenario: boundary — opts first when there is no ctx
    Given "func NewHello(opts *HelloOptions, repo Repository) *Hello"
    When the gidparamorder analyzer checks the file
    Then no diagnostic is reported
    # Without ctx the expected position of opts is the first one.

  Scenario: boundary — an ordinary function taking a logger last
    Given "func decorate(msg string, logger *logrus.Entry) string"
    When the gidparamorder analyzer checks the file
    Then no diagnostic is reported
    # GID-252 is scoped to constructors on purpose: a helper enriching what the
    # caller passes in may take the logger last.

  Scenario: boundary — a logger without opts
    Given "func NewHello(ctx context.Context, logger *logrus.Entry, repo Repository) *Hello"
    When the gidparamorder analyzer checks the file
    Then no diagnostic is reported
    # GID-153 orders the logger against opts; with no opts there is nothing to
    # order it against, and it already precedes the dependencies.

  Scenario: boundary — parameters sharing one type
    Given "func (h *Hello) Do(a, b string, ctx context.Context)"
    When the gidparamorder analyzer checks the file
    Then the "GID-110: …" diagnostic is reported
    # The list is flattened first, so a shared type still counts as two
    # positions before ctx.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — a New* method rather than a function
    Given "func (f *Factory) NewHello(repo Repository, logger *logrus.Entry) *Hello"
    When the gidparamorder analyzer checks the file
    Then no GID-252 diagnostic is reported
    # A constructor here is a package-level function.

  Scenario: non-applicability — a type whose name merely contains Options
    Given "func NewHello(optionsCache Cache, repo Repository) *Hello"
    When the gidparamorder analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — generated code
    Given a file carrying the "Code generated … DO NOT EDIT." header
    When the gidparamorder analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-110, GID-113, GID-153, GID-252)
#  [x] Layer chosen: go/analysis (package paramorder)
#  [x] Messages are defined ("GID-110: …", "GID-113: …", "GID-153: …", "GID-252: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest (methods and constructors)
#  [x] Rule enabled in .golangci.yml
