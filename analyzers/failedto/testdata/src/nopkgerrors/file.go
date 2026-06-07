package nopkgerrors

import stderrors "errors"

// --- Класс 4: неприменимость — файл без github.com/pkg/errors ---
// Локальная функция Wrap с тем же именем не должна матчиться.

func Wrap(err error, message string) error { return err }

func useLocalWrap(err error) error {
	return Wrap(err, "failed to select") // не pkg/errors — не матчится
}

func useStd() error {
	return stderrors.New("failed to do") // std — зона GID-146, не матчится
}
