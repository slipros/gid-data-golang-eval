# language: en

Feature: GID-254 — a local variable holding a constant string expression must be a const
  As a developer
  I want a local string that is assembled ONLY from constants to be declared as a const
  So that the code says what it is — a compile-time constant, not a value built at run time —
  cannot be reassigned by mistake, and does not drift from the sibling methods of the same
  repository, which already declare their queries as `const sqlQuery = …`.

  # Layer: go/analysis (package constexpr, linter gidconstexpr), LoadModeTypesInfo.
  # Origin: resource-registry internal/dal/repository/integration.go (2026-08-04) —
  #   sqlQuery := "SELECT " + integrationColumns + " FROM integration WHERE id = @id"
  # sits next to methods that spell the very same thing as `const sqlQuery = …`. Every
  # operand is a constant, so the compiler folds the expression; only the ":=" suggests
  # otherwise. No standard linter covers this (goconst is about repeated literals,
  # GID-194/gidconstscope is about WHERE a const is declared, not about a var that should
  # be one).
  #
  # Detect: inside a function body, a declaration (`x := <expr>` / `var x = <expr>`) where
  #   - the initializer has a known constant VALUE (types.Info.Types[expr].Value != nil), AND
  #   - its type is string-kinded (a named string type counts — a typed constant is still
  #     a constant), AND
  #   - the initializer is NOT a lone *ast.BasicLit, AND
  #   - the variable is never assigned again and its address is never taken.
  # Reported on the variable name.
  #
  # Deliberate limits: a lone literal (`msg := "hi"`) is left alone — the rule targets
  # expressions that LOOK assembled while being constant, not every local string; numeric
  # constant expressions (`timeout := 5 * 60`) are out of scope for now (rare in our code,
  # higher noise risk — extending the kind check is the obvious next step); a declaration in
  # the init statement of if/for/switch is skipped (Go has no const declaration there).
  # Exceptions: //nolint:gidconstexpr. Generated code and _test.go are skipped.

  # --- Class 1: positive (the violation is caught) ---

  Scenario: positive — a query assembled from constants (the rule's origin)
    Given the function body "sqlQuery := sqlPrefix + columns + \" FROM \" + table + \" WHERE id = @id\"" where every operand is a const
    When the gidconstexpr analyzer checks the file
    Then the diagnostic "GID-254: \"sqlQuery\" is initialized with a constant string expression — every operand is a constant, so the value is folded at compile time and nothing here is dynamic. Fix: declare it as a const (const sqlQuery = \"SELECT \" + columns + \" FROM t\")" is reported on "sqlQuery"

  Scenario: positive — the same concatenation spread over several lines
    Given the function body "sqlQuery := \"SELECT \" + columns +\n\" FROM \" + table + \" WHERE id = @id AND organization_id = @organization_id\""
    When the gidconstexpr analyzer checks the file
    Then the diagnostic "GID-254: \"sqlQuery\" is initialized with a constant string expression …" is reported on "sqlQuery"
    # Line breaks are formatting, not dynamism.

  Scenario: positive — the `var x = <const expr>` spelling
    Given the function body "var sqlQuery = \"SELECT count(*) FROM \" + table"
    When the gidconstexpr analyzer checks the file
    Then the diagnostic "GID-254: \"sqlQuery\" is initialized with a constant string expression …" is reported on "sqlQuery"

  Scenario: positive — a bare reference to another constant, no concatenation
    Given the function body "cols := columns" where columns is a package-level const
    When the gidconstexpr analyzer checks the file
    Then the diagnostic "GID-254: \"cols\" is initialized with a constant string expression …" is reported on "cols"

  Scenario: positive — a named string type: a typed constant expression is still constant
    Given the function body "full := namePrefix + \"integration\"" where namePrefix is a const of type Name (string)
    When the gidconstexpr analyzer checks the file
    Then the diagnostic "GID-254: \"full\" is initialized with a constant string expression …" is reported on "full"

  Scenario: positive — several names declared at once, each judged on its own initializer
    Given the function body "head, tail := sqlPrefix+columns, \" FROM \"+table"
    When the gidconstexpr analyzer checks the file
    Then a diagnostic is reported on "head" and on "tail"

  # --- Class 2: negative (clean code passes) ---

  Scenario: negative — already declared as a const
    Given the function body "const sqlQuery = \"SELECT \" + columns + \" FROM \" + table"
    When the gidconstexpr analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — genuinely dynamic: an operand is a variable
    Given the function "func dynamic(tableName string) string { sqlQuery := \"SELECT \" + columns + \" FROM \" + tableName; return sqlQuery }"
    When the gidconstexpr analyzer checks the file
    Then no diagnostic is reported
    # With a variable operand the expression has no constant value at all.

  Scenario: negative — built at run time through a call
    Given the function body "sqlQuery := fmt.Sprintf(\"SELECT %s FROM %s WHERE id = %d\", columns, table, id)"
    When the gidconstexpr analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — the variable is reassigned later
    Given the function body "sqlQuery := \"SELECT \" + columns + \" FROM \" + table; if withFilter { sqlQuery = sqlQuery + \" WHERE id = @id\" }"
    When the gidconstexpr analyzer checks the file
    Then no diagnostic is reported
    # A const cannot be reassigned — the suggestion would not compile.

  Scenario: negative — the variable's address is taken
    Given the function body "sqlQuery := \"SELECT \" + columns; return &sqlQuery"
    When the gidconstexpr analyzer checks the file
    Then no diagnostic is reported
    # A const has no address.

  # --- Class 3: boundary (looks like a violation but is allowed) ---

  Scenario: boundary — a lone string literal
    Given the function body "msg := \"integration not found\""
    When the gidconstexpr analyzer checks the file
    Then no diagnostic is reported
    # Deliberately out of scope: the rule targets expressions that look assembled while being
    # constant, not every local string in the codebase.

  Scenario: boundary — the declaration is the init statement of an if
    Given the function body "if sqlQuery := \"SELECT \" + columns; sqlQuery != \"\" { return true }"
    When the gidconstexpr analyzer checks the file
    Then no diagnostic is reported
    # Go allows no const declaration in an init statement — there is no fix to suggest.

  Scenario: boundary — a numeric constant expression
    Given the function body "timeout := maxRetry * 60"
    When the gidconstexpr analyzer checks the file
    Then no diagnostic is reported
    # String kind only for now; numeric locals are rare in our code and noisier.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — a package-level var, not a local declaration
    Given the file-level declaration "var packageQuery = \"SELECT \" + columns + \" FROM \" + table"
    When the gidconstexpr analyzer checks the file
    Then no diagnostic is reported
    # Package-level placement is GID-130/GID-194 territory.

  Scenario: non-applicability — the initializer is a function call (no constant value)
    Given the function body "sqlQuery := build()"
    When the gidconstexpr analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — a _test.go file
    Given a test file assembling "expected := \"SELECT \" + columns + \" FROM \" + table"
    When the gidconstexpr analyzer checks the file
    Then no diagnostic is reported
    # Fixtures and expected values legitimately build strings this way.

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-254)
#  [x] Layer chosen: go/analysis (package constexpr: gidconstexpr)
#  [x] Message is defined ("GID-254: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
