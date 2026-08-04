// Eval of GID-255 boundary: a convert leaf under /client/** is function-only
// by design (a converter is a pure function over vocabulary types, GID-235) —
// having no method here is correct, not a missing client.
package convert

// Item is the shape produced by the converter.
type Item struct {
	ID string
}

// ItemFromAPI maps the API payload onto the client's model.
func ItemFromAPI(id string) Item {
	return Item{ID: id}
}
