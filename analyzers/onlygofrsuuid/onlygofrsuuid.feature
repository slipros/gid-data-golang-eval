# language: en

Feature: GID-137 — UUID only through github.com/gofrs/uuid (gidonlygofrsuuid)
  As a developer
  I want a single UUID library across all services
  So that identifiers are generated, parsed and compared the same way
  everywhere, without two uuid.UUID types meeting in one call chain

  # One analyzer over the file's import declarations, no type info needed.
  # Trigger: an import of a known alternative uuid library — google/uuid,
  # satori/go.uuid, pborman/uuid, hashicorp/go-uuid, twinj/uuid.
  # Matching goes through internal/pathseg.SameLibrary, so any major version
  # suffix (…/uuid/v2) counts as the same library.
  # The allowed library, github.com/gofrs/uuid, is named in the diagnostic.
  # Generated code (ast.IsGenerated) is skipped: a stub importing whatever the
  # generator picked is not a choice made in review.

  # --- Class 1: positive ---

  Scenario: positive — google/uuid
    Given "googleuuid \"github.com/google/uuid\"" imported in a service
    When the gidonlygofrsuuid analyzer checks the file
    Then the diagnostic "GID-137: importing \"github.com/google/uuid\" is forbidden (github.com/google/uuid). Fix: use github.com/gofrs/uuid for UUID" is reported on the import

  Scenario: positive — satori/go.uuid
    Given "satori \"github.com/satori/go.uuid\"" imported in a repository
    When the gidonlygofrsuuid analyzer checks the file
    Then the diagnostic "GID-137: importing \"github.com/satori/go.uuid\" is forbidden (…). Fix: use github.com/gofrs/uuid for UUID" is reported

  # --- Class 2: negative ---

  Scenario: negative — the allowed library
    Given "github.com/gofrs/uuid" imported
    When the gidonlygofrsuuid analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — no uuid import at all
    Given a file importing only context and the model
    When the gidonlygofrsuuid analyzer checks the file
    Then no diagnostic is reported

  # --- Class 3: boundary ---

  Scenario: boundary — a major version of a denied library
    Given "github.com/google/uuid/v2" imported
    When the gidonlygofrsuuid analyzer checks the file
    Then the diagnostic "GID-137: …" is reported
    # SameLibrary strips the major version suffix — a new major is the same
    # library.

  Scenario: boundary — a first-party package with uuid in its path
    Given "example.com/uuidutil" imported
    When the gidonlygofrsuuid analyzer checks the file
    Then no diagnostic is reported
    # The list holds whole library paths; "uuid" appearing in a path is not a
    # match.

  Scenario: boundary — the denied library imported under an alias
    Given "uuid \"github.com/google/uuid\"" imported under the alias uuid
    When the gidonlygofrsuuid analyzer checks the file
    Then the diagnostic "GID-137: …" is reported
    # The import path is read, never the local name.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — generated code
    Given a file carrying the "Code generated … DO NOT EDIT." header and importing google/uuid
    When the gidonlygofrsuuid analyzer checks the file
    Then no diagnostic is reported

  Scenario: non-applicability — a uuid library outside the denied list
    Given an unknown "example.org/some/uuidlib" imported
    When the gidonlygofrsuuid analyzer checks the file
    Then no diagnostic is reported
    # The list is explicit; an unknown library is added to deniedPkgs rather
    # than guessed at by name.

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-137)
#  [x] Layer chosen: go/analysis (package onlygofrsuuid)
#  [x] Message is defined ("GID-137: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
