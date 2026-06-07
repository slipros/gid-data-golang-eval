// Eval GID-212: squirrel вне build-пакета (/dal/repository) запрещён;
// произвольные сигнатуры функций вне build не флагаются (неприменимость).
package repository

import (
	"github.com/Masterminds/squirrel" // want `GID-212: squirrel is allowed only in repository build packages \(/dal/repository/build\)\. Fix: move squirrel usage into /dal/repository/build`
)

// --- Класс неприменимости: проверка сигнатуры не действует вне build ---

// Экспортируемая функция с произвольной сигнатурой в /dal/repository — не флагается.
func DoStuff(id string) (int, error) { return 0, nil }

// Функция без результатов вне build — не флагается.
func Reset() {}

// squirrel импортирован выше — пойман как нарушение импорта.
func use() squirrel.SelectBuilder { return squirrel.Select("id") }
