# language: en

Feature: GID-157 — an entity's code is a single contiguous block (gidentitygroup)
  As a developer
  I want the type, its constructor and its methods to sit together as one block
  So that the whole surface of an entity is read in one place, without hunting
  for its methods among free helpers or the declarations of other entities

  # One analyzer over the file's top-level declarations, no type info needed.
  # An entity is a struct type; the declaration file of every struct in the
  # package is collected first (structFiles), so a method or constructor placed
  # in a foreign file is caught across files.
  # Ownership of a declaration: a method — by its receiver type; a function
  # New<Entity> — by the entity, but only when a struct with that name is
  # declared in the package (a New* function without a matching struct is a
  # free function, not a constructor).
  # Order inside the block: type -> New<Entity> -> methods.
  # Contiguity: the block spans from the entity's first declaration to its last;
  # any declaration owned by nobody (a receiverless function, a non-struct type)
  # inside that span splits it. Either edge is fine — above the first type or
  # below the last method — the rule fixes no side.
  # const and var blocks are out of scope: GID-130 already fixes their place in
  # the file, and reporting them here would duplicate its diagnostic.
  # Generated code (ast.IsGenerated) is skipped.

  # --- Class 1: positive ---

  Scenario: positive — a free helper between the entity's methods
    Given a file with "type Connection struct", "NewConnection", "func (c *Connection) ExecContext", the free function "pingDB" and then "func (c *Connection) Ping"
    When the gidentitygroup analyzer checks the file
    Then the diagnostic "GID-157: \"pingDB\" splits the \"Connection\" entity block. Fix: move it above the first type or below the entity's last method" is reported on "pingDB"

  Scenario: positive — a non-struct type inside the block
    Given "type sessionState string" declared between two methods of "Session"
    When the gidentitygroup analyzer checks the file
    Then the diagnostic "GID-157: \"sessionState\" splits the \"Session\" entity block. …" is reported on "sessionState"

  Scenario: positive — a method above its type declaration
    Given "func (s *Snapshot) Early()" declared above "type Snapshot struct" and above "NewSnapshot"
    When the gidentitygroup analyzer checks the file
    Then the diagnostic "GID-157: method \"Early\" must be placed below the \"Snapshot\" type declaration" is reported
    And the diagnostic "GID-157: method \"Early\" must be placed below the NewSnapshot constructor" is reported

  Scenario: positive — the methods of two entities are interleaved
    Given "type Snapshot struct" with its methods, then "type Job struct" with "Run", then "func (s *Snapshot) Render()"
    When the gidentitygroup analyzer checks the file
    Then the diagnostic "GID-157: entity \"Snapshot\" code is interleaved with other entities. Fix: keep the entity block contiguous" is reported on "Render"

  Scenario: positive — a method in a file other than the type's
    Given "type Snapshot struct" declared in snapshot.go and "func (s *Snapshot) Foreign()" declared in foreign.go
    When the gidentitygroup analyzer checks the package
    Then the diagnostic "GID-157: \"Foreign\" belongs to entity \"Snapshot\". Fix: keep the entity's code in the file where it is declared" is reported

  # --- Class 2: negative ---

  Scenario: negative — the canonical order
    Given a file with "type Upload struct", "NewUpload", "func (u *Upload) ID", "func (u *Upload) Start", then "type Download struct" with its methods
    When the gidentitygroup analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — free helpers below the last method of the last entity
    Given the free functions "taskKey" and "drain" declared after every method of every entity in the file
    When the gidentitygroup analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — a free helper above the first type
    Given the free function "normalize" declared above "type Task struct"
    When the gidentitygroup analyzer checks the file
    Then no diagnostic is reported
    # Either edge is allowed — the rule constrains the inside of the block only.

  # --- Class 3: boundary ---

  Scenario: boundary — a free function exactly between two entity blocks
    Given "func taskKey" declared after the last method of "Task" and before "type Queue struct"
    When the gidentitygroup analyzer checks the file
    Then no diagnostic is reported
    # It is outside both spans: below the end of one block, above the start of
    # the next — it splits neither.

  Scenario: boundary — a const block between the entity's methods
    Given "const defaultTimeout = 30 * time.Second" declared between two methods of "Connection"
    When the gidentitygroup analyzer checks the file
    Then no diagnostic is reported
    # Deliberate: the place of const and var in a file is GID-130's rule.

  Scenario: boundary — a New* function without a matching struct
    Given "func NewID() string" in a package where no "type ID struct" is declared
    When the gidentitygroup analyzer checks the file
    Then it is treated as a free function, not as a constructor

  Scenario: boundary — a single entity with no free declarations at all
    Given a file holding only "type Queue struct" and its methods
    When the gidentitygroup analyzer checks the file
    Then no diagnostic is reported

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — a file without struct declarations
    Given a file holding only free functions and interface declarations
    When the gidentitygroup analyzer checks the file
    Then no diagnostic is reported
    # With no entity there is no block, hence no span anything could split.

  Scenario: non-applicability — generated code
    Given a file carrying the "Code generated … DO NOT EDIT." header
    When the gidentitygroup analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — a method on a non-struct type
    Given "type ContextKey string" and "func (k ContextKey) String()" declared in the package
    When the gidentitygroup analyzer checks the file
    Then no "belongs to entity" diagnostic is reported
    # Only structs are entities, so no declaration file is known for the
    # receiver and the file check has nothing to compare.

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-157)
#  [x] Layer chosen: go/analysis (package entitygroup)
#  [x] Message is defined ("GID-157: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
