# language: en

Feature: GID-143 — handling of a missing enum-conversion key (enumconvert)
  As a developer
  I want a map-based enum conversion to handle a missing key
  via gderror.NewUnhandledValueError
  So that an unknown enum value does not silently turn into a zero value

  # Analyzer enumconvert, linter gidenumconvert, LoadMode TypesInfo.
  # Scope: only convert packages (the last path segment is convert),
  #   matched via internal/pathseg.EndsWith.
  # Detection (by types): a map index expression m[key] where the key type is a
  #   named type with underlying string (enum per GID-123), and the value type is
  #   also a named type (enum→enum / enum→model type).
  # The missing-key handler is recognized by the constructor name
  #   NewUnhandledValueError ALONE — the import path is not pinned, so a helper
  #   library on any module path counts (gitlab.gid.team/.../helper.git/errors,
  #   git.k8s.nomilk.space/go-library/ehelper, …).
  # Generated code (ast.IsGenerated) is skipped.

  # --- Class 1: positive (the violation is caught) ---

  Scenario: positive — indexing without comma-ok (single assignment)
    Given a convert package with "v := statusMap[s]" where the key is an enum and the value is a named type
    When the gidenumconvert analyzer checks the file
    Then the diagnostic "GID-143: enum conversion via map without comma-ok. Fix: a missing key must return gderror.NewUnhandledValueError" is reported on "statusMap[s]"

  Scenario: positive — indexing without comma-ok (used as an expression)
    Given a convert package with "return statusMap[s]"
    When the gidenumconvert analyzer checks the file
    Then the diagnostic "GID-143: enum conversion via map without comma-ok …" is reported

  Scenario: positive — comma-ok is present but the function has no NewUnhandledValueError
    Given a convert package with "v, ok := statusMap[s]" and no call to gderror.NewUnhandledValueError in the function body
    When the gidenumconvert analyzer checks the file
    Then the diagnostic "GID-143: a missing enum-conversion key must be handled with gderror.NewUnhandledValueError" is reported on "statusMap[s]"

  # --- Class 2: negative (clean code passes) ---

  Scenario: negative — comma-ok + handling of the missing key
    Given a convert package with "v, ok := statusMap[s]; if !ok { return \"\", gderror.NewUnhandledValueError(s) }"
    When the gidenumconvert analyzer checks the file
    Then no diagnostic is reported

  Scenario: negative — the handler comes from a helper on a non-gitlab module path
    Given a convert package importing "git.k8s.nomilk.space/go-library/ehelper" as gderror
    And "v, ok := formatMap[f]; if !ok { return \"\", gderror.NewUnhandledValueError(f) }"
    When the gidenumconvert analyzer checks the file
    Then no diagnostic is reported
    # Matched by symbol name, not by import path.

  # --- Class 3: boundary (looks like a violation but is allowed) ---

  Scenario: boundary — the map key is a plain string (not an enum)
    Given a convert package with "titleMap[s]" where the key is string, not a named enum
    When the gidenumconvert analyzer checks the file
    Then no diagnostic is reported
    # A plain string/int key is not an enum conversion.

  Scenario: boundary — the map value is a basic type (not named)
    Given a convert package with "weightMap[s]" where the value is int
    When the gidenumconvert analyzer checks the file
    Then no diagnostic is reported
    # The value is not a named type — this is not enum→enum/model.

  Scenario: boundary — the same construct outside a convert package
    Given a package in "/domain/service" (not convert) with "return statusMap[s]"
    When the gidenumconvert analyzer checks the file
    Then no diagnostic is reported
    # The scope is convert packages only.

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — a convert package without enum map indexing
    Given a convert package with ordinary field converters without maps
    When the gidenumconvert analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-143)
#  [x] Layer chosen: go/analysis (package enumconvert: gidenumconvert)
#  [x] Messages are defined ("GID-143: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml
