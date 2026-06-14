# language: en

Feature: GID-233 — no direct cast between enum types (enumcast)
  As a developer
  I want enums to cross layer boundaries only via map conversion
  So that an unknown enum value fails loudly instead of silently passing the boundary

  # Analyzer enumcast, linter gidenumcast, LoadMode TypesInfo.
  # Detection (type-based): a conversion expression DstEnum(x) where
  #   - DstEnum is a string-based enum: a named type with underlying string
  #     that has at least one typed const in its package (GID-123);
  #   - the static type of x is ANOTHER string-based enum declared in a
  #     DIFFERENT package.
  # Scope: the whole codebase — a cross-package enum→enum cast IS the layer
  #   boundary smell, no path scoping needed. _test.go files and generated
  #   code (ast.IsGenerated) are skipped.
  # Complements GID-143 (gidenumconvert): GID-143 checks that the map
  #   converter handles a missing key; GID-233 forbids bypassing the map
  #   converter altogether.
  # Decisions on FP risk:
  #   - The GID-143 map converter itself (map[SrcEnum]DstEnum literal and
  #     indexing) involves no conversion expression — it stays clean.
  #   - Casts between enums of the SAME package are allowed: inside the
  #     enum's own package family they are not a layer boundary.
  #   - A cast of a TYPED const of another package's enum
  #     (model.Status(entityenum.StatusActive)) IS flagged: the static type
  #     of the operand is a foreign enum, and the deterministic core does
  #     not distinguish const from variable operands. Suppress a deliberate
  #     case with //nolint:gidenumcast.
  # Per-case suppression: //nolint:gidenumcast.

  # --- Class 1: positive (violation is caught) ---

  Scenario: positive — cross-package enum→enum direct cast
    Given a conversion "modelenum.Status(s)" where s has type entityenum.Status, both string-based enums with consts
    When the gidenumcast analyzer checks the file
    Then the diagnostic "GID-233: direct cast between enum types crosses a layer boundary unchecked. Fix: convert via map with comma-ok + gderror.NewUnhandledValueError (see GID-143)" is reported on the cast

  Scenario: positive — reverse direction (model → entity) is also caught
    Given a conversion "entityenum.Status(s)" where s has type modelenum.Status
    When the gidenumcast analyzer checks the file
    Then the diagnostic "GID-233 …" is reported on the cast

  # --- Class 2: negative (clean code passes) ---

  Scenario: negative — map conversion per GID-143 stays clean
    Given a converter "v, ok := statusFromEntity[s]; if !ok { return \"\", gderror.NewUnhandledValueError(s) }" with map[entityenum.Status]modelenum.Status
    When the gidenumcast analyzer checks the file
    Then no diagnostic is reported

  # --- Class 3: boundary (close to the line, but allowed) ---

  Scenario: boundary — cast from plain string is allowed
    Given a conversion "modelenum.Status(s)" where s is a plain string
    When the gidenumcast analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — cast to plain string is allowed
    Given a conversion "string(s)" where s has type entityenum.Status
    When the gidenumcast analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — cast of an untyped constant or literal is allowed
    Given a conversion "modelenum.Status(\"active\")"
    When the gidenumcast analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — same-package enum→enum cast is allowed
    Given a conversion "Kind(s)" where s has type Status and both enums are declared in the same package
    When the gidenumcast analyzer checks the file
    Then no diagnostic is reported

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — named string types without consts are not enums
    Given a conversion "modelenum.Label(r)" where neither Label nor the type of r has typed consts in its package
    When the gidenumcast analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — destination without consts is not an enum
    Given a conversion "entityenum.Raw(s)" where Raw has no typed consts and s has type entityenum.Status
    When the gidenumcast analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — _test.go files are skipped
    Given a cross-package enum→enum cast inside a _test.go file
    When the gidenumcast analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — generated code is skipped
    Given a cross-package enum→enum cast inside a file with a "Code generated … DO NOT EDIT." header
    When the gidenumcast analyzer checks the file
    Then no diagnostic is reported

# --- Checklist for adding a new rule ---
#  [x] ID and description registered in RULES.md
#  [x] Layer chosen: go/analysis (type information is required)
#  [x] Severity and message defined
#  [x] Cases covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml (wired by the parent)
