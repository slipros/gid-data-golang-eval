// Eval для GID-194: граничные случаи в обычном пакете вне слоёв.
package plain

import "fmt"

// --- Граница: iota-группа целиком используется одной функцией ---

const ( // want `GID-194: группа констант используется только в "stateName" — объявите её внутри этой функции`
	stateIdle = iota
	stateBusy
)

func stateName(s int) string {
	switch s {
	case stateIdle:
		return "idle"
	case stateBusy:
		return "busy"
	}
	return fmt.Sprintf("unknown:%d", s)
}

// --- Граница: iota-группа используется разными функциями — норма ---

const (
	colorRed = iota
	colorBlue
)

func isRed(c int) bool { return c == colorRed }

func isBlue(c int) bool { return c == colorBlue }

// --- Граница: iota-группа с экспортируемой константой — локализацию
// не предлагаем, диагностика только об экспорте ---

const (
	ModePrimary = iota // want `GID-194: экспортируемая константа "ModePrimary" объявлена вне model/entity — общие константы живут в /domain/model или /dal/entity, локальные объявляются там, где используются`
	modeSecondary
)

func modeLabel() int { return modeSecondary }

// --- Граница: использование в package-level var — константа непереносима ---

const defaultLabel = "default"

var currentLabel = defaultLabel

// --- Граница: использование в сигнатуре (длина массива) — непереносима ---

const bufSize = 8

func fill(buf [bufSize]byte) byte { return buf[0] }

// --- Граница: неиспользуемая константа — зона unused, не GID-194 ---

const orphan = "unused"

// --- Граница: используется только сгенерированным файлом — непереносима ---

const genLabel = "gen"

// --- Граница: используется только тестом — непереносима ---

const testLabel = "test"

func use() (string, bool, bool, int, byte) {
	return stateName(0), isRed(1), isBlue(1), modeLabel(), fill([bufSize]byte{})
}

var _ = use
