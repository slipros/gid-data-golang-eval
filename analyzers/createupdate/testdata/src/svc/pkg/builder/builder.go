// Неприменимость: вне repo/service правило не действует —
// билдеры и фабрики могут возвращать создаваемое значение.
package builder

type Query struct{}

type Builder struct{}

func (b *Builder) CreateQuery() (Query, error) {
	return Query{}, nil
}
