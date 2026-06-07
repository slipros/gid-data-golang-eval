// Package ruleguard содержит простые правила (слой 1) в DSL go-ruleguard.
// Файл подключается в .golangci.yml через gocritic -> ruleguard, а также
// компилируется как обычный Go-код, чтобы eval (rules_test.go) и goimports
// работали с ним стандартно.
//
// Имя функции = имя группы правил: его можно точечно отключить в
// .golangci.yml через settings.ruleguard.disable.
package ruleguard

import "github.com/quasilyte/go-ruleguard/dsl"

// noTimeNow — GID-001: время берётся только через gdhelper.StdTime.Now(),
// прямой вызов time.Now() запрещён.
func noTimeNow(m dsl.Matcher) {
	m.
		Match(`time.Now()`).
		Report(`GID-001: используйте gdhelper.StdTime.Now() вместо time.Now()`)
}

// uuidOnlyV7 — GID-003: UUID генерируются единообразно через
// uuid.Must(uuid.NewV7()); генераторы других версий запрещены.
func uuidOnlyV7(m dsl.Matcher) {
	m.
		Match(
			`uuid.NewV1()`,
			`uuid.NewV3($*_)`,
			`uuid.NewV4()`,
			`uuid.NewV5($*_)`,
			`uuid.NewV6()`,
		).
		Report(`GID-003: UUID генерируются единообразно — uuid.Must(uuid.NewV7())`)
}

// noUUIDEmptyCompare — GID-002: пустой UUID проверяется через IsNil(),
// сравнение с uuid.UUID{} запрещено.
func noUUIDEmptyCompare(m dsl.Matcher) {
	m.Import(`github.com/gofrs/uuid`)

	m.
		Match(`$x == uuid.UUID{}`).
		Where(m["x"].Type.Is(`uuid.UUID`)).
		Report(`GID-002: используйте $x.IsNil() вместо сравнения с uuid.UUID{}`).
		Suggest(`$x.IsNil()`)

	m.
		Match(`$x != uuid.UUID{}`).
		Where(m["x"].Type.Is(`uuid.UUID`)).
		Report(`GID-002: используйте !$x.IsNil() вместо сравнения с uuid.UUID{}`).
		Suggest(`!$x.IsNil()`)
}

// newDeref — GID-005: запрет new($t) для аллокации. Стайлгайд предпочитает
// &T{} для структур и var x T для нулевых значений. Типовой фильтр (только
// структуры) в ruleguard ненадёжен, поэтому матчим все new($t).
func newDeref(m dsl.Matcher) {
	m.
		Match(`new($t)`).
		Report(`GID-005: используйте &T{} для структур или var x T вместо new($t)`)
}

// yodaConditions — GID-006: запрет «йода-условий» — литерал слева в сравнении.
// Переменная слева, литерал справа: $x == "foo", а не "foo" == $x.
// Фильтр исключает случай const == const (обе стороны константы).
func yodaConditions(m dsl.Matcher) {
	m.
		Match(`$lit == $x`, `$lit != $x`).
		Where(m["lit"].Const && !m["x"].Const).
		Report(`GID-006: переменная слева, литерал справа в сравнении`)
}

// quoteVerb — GID-007: запрет ручного экранирования кавычек \"%s\" / \"%v\"
// в format-строках printf-функций — используйте глагол %q.
func quoteVerb(m dsl.Matcher) {
	m.Import(`github.com/pkg/errors`)

	m.
		Match(
			`fmt.Sprintf($f, $*_)`,
			`fmt.Errorf($f, $*_)`,
			`fmt.Printf($f, $*_)`,
			`fmt.Fprintf($w, $f, $*_)`,
			`errors.Errorf($f, $*_)`,
			`errors.Wrapf($err, $f, $*_)`,
		).
		Where(m["f"].Const && m["f"].Text.Matches("\\\\\"%[sv]\\\\\"")).
		Report(`GID-007: используйте %q вместо ручного экранирования \"%s\"/\"%v\"`)
}

// noDeepEqual — GID-008: запрет reflect.DeepEqual — в тестах cmp/require,
// в коде явное сравнение.
func noDeepEqual(m dsl.Matcher) {
	m.
		Match(`reflect.DeepEqual($x, $y)`).
		Report(`GID-008: в тестах — cmp/require, в коде — явное сравнение вместо reflect.DeepEqual`)
}
