// Eval of GID-186: the format string of printf functions is a literal or
// a const, not a variable.
package fmtconst

import (
	"fmt"
	"io"
	"log"

	"github.com/pkg/errors"
)

// fmtStr — a constant: using it as the format is allowed.
const fmtStr = "значение %d"

var errExternal = errors.New("boom")

// printf — a local function with a printf-like signature; not from the target
// packages, must not be matched (boundary class).
func printf(format string, args ...any) {}

// --- Class 1: positive (a variable in the format position) ---

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

// --- Class 2: negative (a literal / const / concatenation of constants) ---

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

// --- Class 3: boundary ---

func boundarySprint(s string) string {
	// fmt.Sprint — not printf (no format position) — not matched.
	return fmt.Sprint(s)
}

func boundaryLocalPrintf(s string, x int) {
	// a local printf(format, ...) function — not from the target packages — not matched.
	printf(s, x)
}
