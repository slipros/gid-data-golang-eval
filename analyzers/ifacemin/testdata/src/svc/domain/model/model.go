// Inapplicability of GID-197: the model layer is out of scope — model
// interfaces may describe a contract for external consumers.
package model

type Filterable interface {
	Apply(query string) string
	Reset()
}
