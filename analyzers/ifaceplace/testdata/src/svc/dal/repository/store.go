// Eval for GID-271: the consumers of the dependencies.go port file. Three
// distinct structs use its interfaces — 3 ≥ 2, the port file stays.
package repository

// Store uses JobReader and JobWriter.
type Store struct {
	reader JobReader
	writer JobWriter
}

// Archive uses JobReader and JobLister.
type Archive struct {
	reader JobReader
	lister JobLister
}

// Remote uses JobLister only.
type Remote struct {
	lister JobLister
}
