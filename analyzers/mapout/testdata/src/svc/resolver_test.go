package svc

// --- A _test.go file IS judged (GID-250): a fixture helper can return its map
// exactly as easily as production code, and no production interface has a
// fill-in-the-map method for a double to mirror — this rule removes those. ---

func seedCabinets(ids []string, into map[string]Cabinet) {
	for _, id := range ids {
		into[id] = Cabinet{ID: id} // want `GID-257: a map received as a parameter is filled in`
	}
}
