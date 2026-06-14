# language: en

Feature: GID-180 — init() is deterministic (avoid init)
  As a developer
  I want func init() not to start goroutines and not to do I/O
  So that package initialization is deterministic, while background work and
  calls to the OS/network/DB are started explicitly from main/a constructor/app

  # Linter: gidinitclean. LoadMode: TypesInfo (the imported package path is needed).
  # Forbidden inside func init():
  #   1) a go statement — directly in the init body, including nested blocks and the
  #      bodies of closures declared in init itself;
  #   2) calls to functions of I/O packages (pkg.Func selector, the package path via
  #      TypesInfo). The default list: os, net, net/http, database/sql,
  #      io/ioutil, bufio. Configurable via settings.packages (replaces the default).
  # Reading env (os.Getenv, os.LookupEnv) is allowed — it is not I/O.
  # Generated code (ast.IsGenerated) is skipped.

  Scenario: starting a goroutine in init — violation (positive)
    Given the package declares "func init() { go func(){}() }"
    When the analyzer checks the file
    Then a "GID-180" diagnostic about a goroutine in init() is reported

  Scenario: the I/O call os.Open in init — violation (positive)
    Given the package declares "func init() { os.Open(\"/etc/hosts\") }"
    When the analyzer checks the file
    Then a "GID-180" diagnostic about the I/O call os.Open is reported

  Scenario: the I/O call sql.Open in init — violation (positive)
    Given the package declares "func init() { sql.Open(\"postgres\", \"\") }"
    When the analyzer checks the file
    Then a "GID-180" diagnostic about the I/O call database/sql.Open is reported

  Scenario: constructing a map in init — ok (negative)
    Given the package declares "func init() { cfg[\"a\"] = \"1\" }"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: reading env in init — ok (negative)
    Given the package declares "func init() { os.Getenv(\"HOST\"); os.LookupEnv(\"PORT\") }"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a go statement in an ordinary function — the rule does not apply (negative)
    Given the package declares "func StartWorker() { go func(){}() }"
    When the analyzer checks the file
    Then no diagnostic is reported

  Scenario: a closure declared and called in init does os.Open — violation (boundary)
    Given the package declares "func init() { fn := func(){ os.Open(\"/tmp/x\") }; fn() }"
    When the analyzer checks the file
    Then a "GID-180" diagnostic about the I/O call os.Open is reported
    # The body of a closure declared in init is traversed as part of init.

  Scenario: a helper outside init with os.Open, called from init — NOT matched (boundary, limitation)
    Given the package declares "func loadFile(){ os.Open(...) }" and "func init(){ loadFile() }"
    When the analyzer checks the file
    Then no diagnostic is reported
    # Limitation: the analysis is intra-procedural — the body of a separate function called
    # from init is not traversed as init. No call graph is built.

  Scenario: a package without init() — the rule does not apply (non-applicability)
    Given the package has no "func init()" at all, but has a go statement and os.Open in an ordinary function
    When the analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-180)
#  [x] Layer chosen: go/analysis (the package path via TypesInfo + AST are needed)
#  [x] Severity and message are defined
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml
