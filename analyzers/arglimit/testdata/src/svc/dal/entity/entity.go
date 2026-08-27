// Eval of GID-272 non-applicability: a package outside /domain/** is not
// judged — the rule is about the domain layer, where the signature is ours to
// change. A DAL entity builder may take as many arguments as the columns it
// fills.
package entity

func buildEntity(a, b, c, d, e string) *Entity { return &Entity{} }

// Entity is a DAL entity.
type Entity struct {
	ID string
}
