// Eval для GID-007 (quote-verb).
package quoteverb

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
)

// --- Позитивные кейсы: нарушение ловится ---

func badSprintf(name string) string {
	return fmt.Sprintf("name=\"%s\"", name) // want `GID-007: используйте %q вместо ручного экранирования`
}

func badErrorf(v int) error {
	return fmt.Errorf("value=\"%v\"", v) // want `GID-007: используйте %q вместо ручного экранирования`
}

func badPkgWrapf(err error, name string) error {
	return errors.Wrapf(err, "name=\"%s\"", name) // want `GID-007: используйте %q вместо ручного экранирования`
}

// Граничный кейс: Fprintf с экранированием тоже ловится.
func badFprintf(name string) {
	fmt.Fprintf(os.Stdout, "[\"%s\"]", name) // want `GID-007: используйте %q вместо ручного экранирования`
}

// --- Негативные кейсы: уже используется %q ---

func goodQuote(name string) string {
	return fmt.Sprintf("name=%q", name)
}

// --- Граничный кейс: %s без кавычек не матчится ---

func boundaryPlainVerb(name string) string {
	return fmt.Sprintf("name=%s", name)
}

// --- Неприменимость: не printf-функция ---

func notApplicable(name string) string {
	return "name=\"" + name + "\""
}
