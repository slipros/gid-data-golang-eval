// Eval GID-212: squirrel в /domain/service запрещён (только build-пакеты).
package service

import (
	"github.com/Masterminds/squirrel" // want `GID-212: squirrel используется только в build-пакетах репозитория \(/dal/repository/build\)`
)

// --- Класс неприменимости: проверка сигнатуры не действует вне build ---

// Произвольная сигнатура в сервисе — не флагается.
func Process(id string) (bool, error) { return true, nil }

func use() squirrel.SelectBuilder { return squirrel.Select("id") }
