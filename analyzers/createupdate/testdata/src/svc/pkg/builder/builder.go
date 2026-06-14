// Not applicable: outside repo/service the rule does not apply —
// builders and factories may return the value being created.
package builder

type Query struct{}

type Builder struct{}

func (b *Builder) CreateQuery() (Query, error) {
	return Query{}, nil
}
