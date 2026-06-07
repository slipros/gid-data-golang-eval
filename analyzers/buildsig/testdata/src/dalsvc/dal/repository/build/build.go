// Eval GID-212: контракт build-функций в /dal/repository/build.
package build

import (
	"pgx/batch"

	"github.com/Masterminds/squirrel"
)

// --- Негативный класс: корректные сигнатуры проходят ---

// Одиночный запрос: (sql string, args []any, err error) — ок.
func SelectJobs(status string) (string, []any, error) {
	return "SELECT 1", []any{status}, nil
}

// Batch-операция: (*batch.Batch, error) — ок.
func InsertJobsBatch(ids []string) (*batch.Batch, error) {
	return &batch.Batch{}, nil
}

// squirrel импортирован и используется в build-пакете — ок (проверка 2 здесь не действует).
func buildSquirrel() (string, []any, error) {
	return squirrel.Select("id").ToSql()
}

// --- Позитивный класс: нарушение контракта сигнатуры ловится ---

// Возвращает (string, error) — не соответствует ни одному контракту.
func BuildBad(status string) (string, error) { // want `GID-212: a build function must return \(sql string, args \[\]any, err error\) or \(\*batch\.Batch, error\)\. Fix: adjust the signature`
	return "", nil
}

// Возвращает *squirrel.SelectBuilder — билдер не разрешён как результат.
func BuildBuilder() *squirrel.SelectBuilder { // want `GID-212: a build function must return \(sql string, args \[\]any, err error\) or \(\*batch\.Batch, error\)\. Fix: adjust the signature`
	b := squirrel.Select("id")
	return &b
}

// --- Граничный класс ---

// Функция без результатов — нарушение (пустой список результатов).
func BuildVoid() { // want `GID-212: a build function must return \(sql string, args \[\]any, err error\) or \(\*batch\.Batch, error\)\. Fix: adjust the signature`
}

// Неэкспортируемый хелпер с другой сигнатурой — не флагается.
func helper(n int) int { return n }
