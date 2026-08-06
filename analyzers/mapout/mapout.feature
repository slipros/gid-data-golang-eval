# language: en

Feature: GID-257 — a map received as a parameter is not filled in
  As the styleguide owner
  I want what a function produces to be visible in its signature
  So that a caller reading `func resolve(ctx, ids) map[K]V` knows what comes back, instead of
  handing a map down and hoping. A map passed IN as data is fine; a map passed in to be FILLED
  hides the result from the type system and makes every outcome look identical from the outside —
  a chunk that resolved nothing and a chunk that failed leave the caller in the same state.

  # Layer: go/analysis (package mapout, linter gidmapout), LoadModeTypesInfo.
  # Config: settings.exclude ("Function" | "Type.Method").
  #
  # Detect: a FuncDecl with a parameter of map type whose body WRITES to that parameter:
  #   param[k] = v · param[k]++ · param[k] += n · delete(param, k) · clear(param)
  # Reading it (a lookup, a range, len) is not a write.
  #
  # Not reported: a map RECEIVER (a method on a named map type mutates its own value — that is
  # what such a type exists for), a map in a struct field, a local map built inside the function.
  #
  # Why (incident 2026-08-06, advertising-api internal/domain/service/ad_cabinet_resolver.go):
  #   func (a *AdCabinetResolver) resolveChunk(ctx context.Context, chunk []uuid.UUID,
  #       result map[uuid.UUID]model.AdCabinet) { … result[id] = model.AdCabinet{…} }
  # The whole output of the method went into a map handed down by the caller; the signature
  # promised nothing, and a registry outage was indistinguishable from an empty chunk. Returning
  # the map puts the result back where the reader looks for it — and leaves room for the error
  # result the failing branch needs (GID-258).
  #
  # Generated code is skipped. A _test.go file IS judged: a fixture can return its map exactly as
  # easily as production code (GID-250), and a double has no production interface with a
  # fill-in-the-map method to mirror — this rule removes those.

  # --- Class 1: positive (the violation is caught) ---

  Scenario: positive — the incident shape: the result goes into a caller's map
    Given the method "func (r *Resolver) resolveChunk(chunk []string, result map[string]Cabinet) { for _, id := range chunk { result[id] = Cabinet{ID: id} } }"
    When the gidmapout analyzer checks the file
    Then the diagnostic "GID-257: a map received as a parameter is filled in — the result is missing from the signature, so the caller cannot tell an empty result from a failure. Fix: return it instead (func resolve(ctx, ids) map[K]V) and let the caller merge; a map parameter is for data going IN, not for a result coming OUT" is reported on "result[id]"

  Scenario: positive — a counter accumulated into a caller's map
    Given the function "func countInto(items []string, counts map[string]int) { for _, item := range items { counts[item]++ } }"
    When the gidmapout analyzer checks the file
    Then the diagnostic "GID-257: a map received as a parameter is filled in …" is reported

  Scenario: positive — a compound assignment is a write too
    Given the function "func sumInto(items []string, totals map[string]int) { for _, item := range items { totals[item] += 2 } }"
    When the gidmapout analyzer checks the file
    Then the diagnostic "GID-257: a map received as a parameter is filled in …" is reported

  Scenario: positive — delete and clear mutate the map without an assignment
    Given the functions "func dropFrom(ids []string, cache map[string]Cabinet) { for _, id := range ids { delete(cache, id) } }" and "func resetAll(cache map[string]Cabinet) { clear(cache) }"
    When the gidmapout analyzer checks the file
    Then the diagnostic "GID-257: a map received as a parameter is filled in …" is reported on each

  Scenario: positive — a _test.go helper filling a map handed to it
    Given the test helper "func seedCabinets(ids []string, into map[string]Cabinet) { for _, id := range ids { into[id] = Cabinet{ID: id} } }" in resolver_test.go
    When the gidmapout analyzer checks the file
    Then the diagnostic "GID-257: a map received as a parameter is filled in …" is reported
    # The test side is judged: returning the map is just as easy in a fixture.

  # --- Class 2: negative (clean code passes) ---

  Scenario: negative — the function returns its result
    Given the method "func (r *Resolver) resolveChunkClean(chunk []string) map[string]Cabinet { result := make(map[string]Cabinet, len(chunk)); for _, id := range chunk { result[id] = Cabinet{ID: id} }; return result }"
    When the gidmapout analyzer checks the file
    Then no diagnostic is reported
    # The shape the rule asks for: the map is built locally and handed back.

  Scenario: negative — a map parameter that is only READ
    Given the function "func titlesOf(ids []string, known map[string]Cabinet) []string { … if cabinet, ok := known[id]; ok { … } … }"
    When the gidmapout analyzer checks the file
    Then no diagnostic is reported
    # Data going in, not a result coming out.

  # --- Class 3: boundary ---

  Scenario: boundary — a map held in a struct field is not a parameter
    Given the method "func (c *Cache) put(id string, cabinet Cabinet) { c.items[id] = cabinet }"
    When the gidmapout analyzer checks the file
    Then no diagnostic is reported

  Scenario: boundary — a map RECEIVER (a method on a named map type)
    Given "type Registry map[string]Cabinet" and the method "func (r Registry) Put(id string, cabinet Cabinet) { r[id] = cabinet }"
    When the gidmapout analyzer checks the file
    Then no diagnostic is reported
    # Mutating its own value is exactly what a named map type exists for.

  Scenario: boundary — a LOCAL map built inside the function
    Given the function "func buildLocal(ids []string) map[string]Cabinet { local := make(map[string]Cabinet, len(ids)); for _, id := range ids { local[id] = Cabinet{ID: id} }; return local }"
    When the gidmapout analyzer checks the file
    Then no diagnostic is reported

  # --- Class 4: non-applicability ---

  Scenario: non-applicability — no map parameter at all
    Given the function "func plain(ids []string) int { return len(ids) }"
    When the gidmapout analyzer checks the file
    Then no diagnostic is reported

  # --- Config: settings.exclude ---

  Scenario: config — an excluded method may fill a map parameter
    Given settings.exclude contains "Resolver.legacyResolveChunk" and that method writing into its map parameter
    When the gidmapout analyzer checks the file
    Then no diagnostic is reported

# --- Checklist when adding a new rule ---
#  [x] ID and description are recorded in the registry (RULES.md, GID-257)
#  [x] Layer chosen: go/analysis (package mapout: gidmapout)
#  [x] Message is defined ("GID-257: …")
#  [x] Case classes covered: positive, negative, boundary, non-applicability
#  [x] testdata with // want for analysistest
#  [x] Rule enabled in .golangci.yml
