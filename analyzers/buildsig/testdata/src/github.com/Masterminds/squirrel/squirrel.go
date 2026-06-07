// Stub библиотеки github.com/Masterminds/squirrel для eval.
package squirrel

// SelectBuilder — заглушка билдера SELECT-запросов.
type SelectBuilder struct{}

// ToSql возвращает sql, args и ошибку.
func (b SelectBuilder) ToSql() (string, []any, error) { return "", nil, nil }

// Select создаёт SelectBuilder.
func Select(columns ...string) SelectBuilder { return SelectBuilder{} }
