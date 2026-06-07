// Eval GID-186: format-строка printf-функций — литерал или const,
// не переменная.
package fmtconst

import (
	"fmt"
	"io"
	"log"

	"github.com/pkg/errors"
)

// fmtStr — константа: format из неё разрешён.
const fmtStr = "значение %d"

var errExternal = errors.New("boom")

// printf — локальная функция с printf-подобной сигнатурой; не из целевых
// пакетов, не должна матчиться (граничный класс).
func printf(format string, args ...any) {}

// --- Класс 1: позитивный (переменная в позиции format) ---

func positiveSprintf(s string, x int) string {
	return fmt.Sprintf(s, x) // want `GID-186: the format string is a variable\. Fix: declare a const, otherwise vet cannot check the arguments`
}

func positiveFprintf(w io.Writer, s string, x int) {
	fmt.Fprintf(w, s, x) // want `GID-186: the format string is a variable\. Fix: declare a const, otherwise vet cannot check the arguments`
}

func positiveWrapf(s string, x int) error {
	return errors.Wrapf(errExternal, s, x) // want `GID-186: the format string is a variable\. Fix: declare a const, otherwise vet cannot check the arguments`
}

func positivePrintf(s string, x int) {
	fmt.Printf(s, x) // want `GID-186: the format string is a variable\. Fix: declare a const, otherwise vet cannot check the arguments`
}

func positiveErrorf(s string, x int) error {
	return fmt.Errorf(s, x) // want `GID-186: the format string is a variable\. Fix: declare a const, otherwise vet cannot check the arguments`
}

func positivePkgErrorsErrorf(s string, x int) error {
	return errors.Errorf(s, x) // want `GID-186: the format string is a variable\. Fix: declare a const, otherwise vet cannot check the arguments`
}

func positiveWithMessagef(s string, x int) error {
	return errors.WithMessagef(errExternal, s, x) // want `GID-186: the format string is a variable\. Fix: declare a const, otherwise vet cannot check the arguments`
}

func positiveLogPrintf(s string, x int) {
	log.Printf(s, x) // want `GID-186: the format string is a variable\. Fix: declare a const, otherwise vet cannot check the arguments`
}

func positiveLogFatalf(s string, x int) {
	log.Fatalf(s, x) // want `GID-186: the format string is a variable\. Fix: declare a const, otherwise vet cannot check the arguments`
}

// --- Класс 2: негативный (литерал / const / конкатенация констант) ---

func negativeLiteral(x int) string {
	return fmt.Sprintf("значение %d", x)
}

func negativeConst(x int) string {
	return fmt.Sprintf(fmtStr, x)
}

func negativeConcatConst(x int) string {
	return fmt.Sprintf("a"+"b %d", x)
}

func negativeFprintfLiteral(w io.Writer, x int) {
	fmt.Fprintf(w, "значение %d", x)
}

// --- Класс 3: граничный ---

func boundarySprint(s string) string {
	// fmt.Sprint — не printf (нет позиции format) — не матчится.
	return fmt.Sprint(s)
}

func boundaryLocalPrintf(s string, x int) {
	// своя функция printf(format, ...) — не из целевых пакетов — не матчится.
	printf(s, x)
}
