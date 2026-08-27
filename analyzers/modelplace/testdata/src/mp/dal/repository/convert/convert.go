// Eval for GID-270 (model-place): a convert package lives on ANY layer — the
// boundary is the final path segment, not /domain/**. A converter in
// /dal/repository is judged exactly like one under /domain.
package convert

// Row — positive (part A): the declaration is judged outside /domain too.
type Row struct { // want `GID-270: type "Row" is declared in a convert package`
	ID   string
	Body []byte
}

// cursor — an unexported struct: part A lets it through, part C judges the
// function handing it out.
type cursor struct {
	offset int
}

// nextCursor — positive (part C): a converter of the DAL layer hands out a
// struct declared next to it.
func nextCursor(offset int) cursor { // want `GID-270: function "nextCursor" returns "cursor" — a struct declared in this package, and a convert package holds no data types of its own\. Fix: declare the returned type in /domain/model \(or //nolint:gidmodelplace when explicitly intended\)`
	return cursor{offset: offset}
}

// RowsFrom — boundary: "Row" is already reported at its declaration.
func RowsFrom(ids []string) []Row {
	return nil
}
