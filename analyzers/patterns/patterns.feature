# language: en

Feature: GID-001…008 — the simple AST patterns (gidtimenow, giduuidnil, giduuidversion, gidnewderef, gidyoda, gidquoteverb, giddeepequal)
  As a developer
  I want the small mechanical rules enforced by the binary itself
  So that no consumer has to copy a ruleguard rules file next to their config

  # Seven analyzers in one package, each shipping as its own linter so it can be
  # toggled independently in .golangci.yml. All of them run on
  # LoadModeTypesInfo and resolve callees through typeutil.Callee, so a local
  # function sharing a name with a stdlib one is never mistaken for it.
  # Generated code (ast.IsGenerated) is skipped by every one of them.

  # ═══ GID-001 gidtimenow — no direct time.Now() ═══

  Scenario: positive — a direct time.Now call
    Given "return time.Now()"
    When the gidtimenow analyzer checks the file
    Then the diagnostic "GID-001: time.Now() must not be called directly. Fix: use gdhelper.StdTime.Now() instead of time.Now()." is reported

  Scenario: negative — the helper clock
    Given "return gdhelper.StdTime.Now()"
    When the gidtimenow analyzer checks the file
    Then no diagnostic is reported
    # An injectable clock is what makes time testable.

  Scenario: boundary — a local function named Now
    Given a package-level "func Now() time.Time" and a call to it
    When the gidtimenow analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — other time package functions
    Given "time.Since(start)" and "time.Parse(layout, s)"
    When the gidtimenow analyzer checks the file
    Then no diagnostic is reported

  # ═══ GID-002 giduuidnil — compare a UUID with IsNil() ═══

  Scenario: positive — comparison with an empty literal
    Given "eq := id == uuid.UUID{}"
    When the giduuidnil analyzer checks the file
    Then the diagnostic "GID-002: do not compare a UUID with uuid.UUID{}. Fix: replace \"id == uuid.UUID{}\" with \"id.IsNil()\"." is reported with a suggested fix

  Scenario: boundary — the negated comparison
    Given "ne := id != uuid.UUID{}"
    When the giduuidnil analyzer checks the file
    Then the diagnostic suggesting "!id.IsNil()" is reported
    # The operand text is rendered back from the AST, so the fix reads like the
    # code it replaces.

  Scenario: negative — IsNil is already used
    Given "if id.IsNil() {"
    When the giduuidnil analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — an empty literal of another struct type
    Given "if p == point{} {"
    When the giduuidnil analyzer checks the file
    Then no diagnostic is reported
    # Only the gofrs uuid.UUID type is matched, at any major version.

  # ═══ GID-003 giduuidversion — generate UUIDs with uuid.Must(uuid.NewV7()) ═══

  Scenario: positive — an older generator version
    Given "_, _ = uuid.NewV1()"
    When the giduuidversion analyzer checks the file
    Then the diagnostic "GID-003: UUIDs must be generated uniformly. Fix: use uuid.Must(uuid.NewV7()) instead of uuid.NewV1()." is reported

  Scenario: positive — NewV7 with the error handled by hand
    Given "id, _ := uuid.NewV7()"
    When the giduuidversion analyzer checks the file
    Then the diagnostic "GID-003: UUIDs must be generated via uuid.Must. Fix: use uuid.Must(uuid.NewV7()) instead of handling the error." is reported
    # NewV7's error never fires; consuming it with Must keeps call sites short.

  Scenario: negative — the canonical form
    Given "return uuid.Must(uuid.NewV7())"
    When the giduuidversion analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — Must wrapping an older version
    Given "return uuid.Must(uuid.NewV4())"
    When the giduuidversion analyzer checks the file
    Then the diagnostic "GID-003: … instead of uuid.NewV4()." is reported
    # Must blesses the error handling, not the version.

  Scenario: boundary — a major version of the library
    Given the same calls on "github.com/gofrs/uuid/v5"
    When the giduuidversion analyzer checks the file
    Then the same diagnostics are reported
    # The library is matched through pathseg.SameLibrary.

  # ═══ GID-005 gidnewderef — avoid the new() builtin ═══

  Scenario: positive — new() on a struct type
    Given "return new(point)"
    When the gidnewderef analyzer checks the file
    Then the diagnostic "GID-005: avoid the new() builtin. Fix: use \"&T{}\" for structs or \"var x T\" instead of \"new(T)\"." is reported

  Scenario: negative — the composite literal form
    Given "return &point{}"
    When the gidnewderef analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — a local function named new
    Given a package-level function shadowing the builtin and a call to it
    When the gidnewderef analyzer checks the file
    Then no diagnostic is reported
    # The callee is checked against *types.Builtin.

  # ═══ GID-006 gidyoda — the literal goes on the right ═══

  Scenario: positive — a constant on the left
    Given "return 0 == x"
    When the gidyoda analyzer checks the file
    Then the diagnostic "GID-006: yoda condition — the literal must be on the right. Fix: write \"x == 0\" instead of \"0 == x\"." is reported

  Scenario: boundary — the != operator
    Given "return 5 != x"
    When the gidyoda analyzer checks the file
    Then the diagnostic "GID-006: yoda condition …" is reported

  Scenario: negative — the natural order
    Given "return x == 0"
    When the gidyoda analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — both sides constant
    Given "const ok = 1 == 1"
    When the gidyoda analyzer checks the file
    Then no diagnostic is reported
    # The shape needs a constant on the left and a non-constant on the right.

  # ═══ GID-007 gidquoteverb — use %q instead of hand-escaped quotes ═══

  Scenario: positive — an escaped %s in Sprintf
    Given "s := fmt.Sprintf(\"\\\"%s\\\"\", name)"
    When the gidquoteverb analyzer checks the file
    Then the diagnostic "GID-007: do not escape quotes around %s/%v by hand. Fix: use %q instead of \\\"%s\\\"." is reported

  Scenario: positive — the same in errors.Errorf
    Given "_ = errors.Errorf(\"\\\"%v\\\"\", name)"
    When the gidquoteverb analyzer checks the file
    Then the diagnostic "GID-007: do not escape quotes …" is reported
    # The format argument index is known per function: fmt.Sprintf/Errorf/Printf
    # and errors.Errorf take it first, fmt.Fprintf and errors.Wrapf second.

  Scenario: negative — the %q verb
    Given "s := fmt.Sprintf(\"%q\", name)"
    When the gidquoteverb analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — a non-constant format string
    Given "fmt.Sprintf(format, name)" where format is a variable
    When the gidquoteverb analyzer checks the file
    Then no diagnostic is reported
    # The format must be a constant to be inspected.

  # ═══ GID-008 giddeepequal — avoid reflect.DeepEqual ═══

  Scenario: positive — a DeepEqual call
    Given "return reflect.DeepEqual(a, b)"
    When the giddeepequal analyzer checks the file
    Then the diagnostic "GID-008: avoid reflect.DeepEqual. Fix: use require/cmp in tests or explicit field comparison in code." is reported

  Scenario: negative — an explicit comparison
    Given "return a.ID == b.ID && a.Name == b.Name"
    When the giddeepequal analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — other reflect functions
    Given "reflect.TypeOf(v)"
    When the giddeepequal analyzer checks the file
    Then no diagnostic is reported

  # ═══ Common: non-applicability ═══

  Scenario: non-applicability — generated code
    Given a file carrying the "Code generated … DO NOT EDIT." header and holding every pattern above
    When any of the seven analyzers checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] IDs and descriptions are recorded in the registry (RULES.md, GID-001, 002, 003, 005, 006, 007, 008)
#  [x] Layer chosen: go/analysis (package patterns), one analyzer per rule
#  [x] Messages are defined ("GID-00N: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest (one fixture package per rule)
#  [x] Rules enabled in .golangci.yml
