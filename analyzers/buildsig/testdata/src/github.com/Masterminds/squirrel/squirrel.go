// Stub of the github.com/Masterminds/squirrel library for eval.
package squirrel

// SelectBuilder — a stub of the SELECT query builder.
type SelectBuilder struct{}

// ToSql returns sql, args and an error.
func (b SelectBuilder) ToSql() (string, []any, error) { return "", nil, nil }

// Select creates a SelectBuilder.
func Select(columns ...string) SelectBuilder { return SelectBuilder{} }
