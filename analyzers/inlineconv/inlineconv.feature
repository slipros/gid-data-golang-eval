# language: en

Feature: GID-215 — model ↔ entity conversion lives only in convert packages (no-inline-entity-literal)
  As a developer
  I want inline filling of entity types in the domain layer to be forbidden
  So that all model ↔ entity conversion lives in the convert package (<Dst><Type>From<Src>)

  # One analyzer inlineconv → linter gidinlineconv, LoadModeTypesInfo.
  # Source: service.md "Conversion is always performed via the convert package".
  # Scope: domain-layer packages (pathseg.Contains(pkgPath, "domain")), EXCEPT packages
  # with a convert segment. The literal type is resolved via TypesInfo: a named
  # type (a struct or a named slice) from a package of the entity layer
  # (pathseg.Contains(the type's package, "dal", "entity"), including filter/enum).
  # An empty literal (entity.Snapshot{}) is a zero value, allowed.
  # Only the outermost entity literal is flagged — we do not descend inside.
  # _test.go and generated code (ast.IsGenerated) are skipped.

  # --- Class 1: positive (the violation is caught) ---

  Scenario: positive — entity.CreateSnapshot{Field: ...} in /domain/service
    Given "return entity.CreateSnapshot{Name: name}" in the /domain/service package
    When the gidinlineconv analyzer checks the file
    Then the diagnostic "GID-215: inline-filling the entity type entity.CreateSnapshot in the domain layer is forbidden. Fix: put conversion in a convert package (<Dst><Type>From<Src>)" is reported

  Scenario: positive — &entity.Snapshot{Field: ...}
    Given "return &entity.Snapshot{ID: id}" in the /domain/service package
    When the gidinlineconv analyzer checks the file
    Then the diagnostic "GID-215: inline-filling the entity type entity.Snapshot in the domain layer is forbidden" is reported

  Scenario: positive — a named entity slice with elements
    Given "return entity.Snapshots{{ID: \"a\"}, {ID: \"b\"}}" in the /domain/service package
    When the gidinlineconv analyzer checks the file
    Then the diagnostic "GID-215: inline-filling the entity type entity.Snapshots in the domain layer is forbidden" is reported

  Scenario: positive — a filter struct from /dal/entity/filter with fields
    Given "return filter.Snapshots{Name: name, Limit: 10}" in the /domain/service package
    When the gidinlineconv analyzer checks the file
    Then the diagnostic "GID-215: inline-filling the entity type filter.Snapshots in the domain layer is forbidden" is reported

  # --- Class 2: negative (clean code passes) ---

  Scenario: negative — an empty entity literal (zero value)
    Given "return entity.Snapshot{}" in the /domain/service package
    When the gidinlineconv analyzer checks the file
    Then no diagnostic is reported
    # An empty literal is a zero value, not an inline conversion.

  Scenario: negative — a model-type literal with fields
    Given "return model.Snapshot{ID: id, Name: name}" in the /domain/service package
    When the gidinlineconv analyzer checks the file
    Then no diagnostic is reported
    # model in domain is the norm; the rule is only about entity types.

  Scenario: negative — an entity literal inside the service's convert package
    Given "return entity.CreateSnapshot{Name: in.Name}" in the /domain/service/convert package
    When the gidinlineconv analyzer checks the file
    Then no diagnostic is reported
    # The convert package is the place for conversion, inline entity is allowed.

  # --- Class 3: boundary ---

  Scenario: boundary — a nested literal inside a flagged outer one
    Given "entity.Snapshots{entity.Snapshot{ID: \"a\"}, entity.Snapshot{ID: \"b\"}}" in the /domain/service package
    When the gidinlineconv analyzer checks the file
    Then exactly one diagnostic is reported — on the outer literal entity.Snapshots
    # The analyzer does not descend inside a flagged literal.

  Scenario: boundary — map[string]entity.X{ key: entity.X{...} }
    Given "map[string]entity.Snapshot{id: entity.Snapshot{ID: id}}" in the /domain/service package
    When the gidinlineconv analyzer checks the file
    Then a diagnostic is reported on the entity.Snapshot value, but not on the map literal
    # The map literal itself is not a named entity type (not flagged); the
    # entity.Snapshot value is flagged, including an element without an explicit type ({ID: id}).

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — an entity literal in /dal/repository
    Given "return entity.Snapshot{ID: id}" in the /dal/repository package
    When the gidinlineconv analyzer checks the file
    Then no diagnostic is reported
    # /dal/repository is not part of the domain layer — outside the rule's scope.

  Scenario: non-applicability — _test.go
    Given "return entity.Snapshot{ID: id}" in the file service_test.go
    When the gidinlineconv analyzer checks the file
    Then no diagnostic is reported
    # Test files are skipped.

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-215)
#  [x] Layer chosen: go/analysis (package inlineconv: gidinlineconv), LoadModeTypesInfo
#  [x] Message is defined ("GID-215: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [ ] Rule enabled in .golangci.yml
