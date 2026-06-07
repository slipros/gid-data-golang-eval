// Класс «неприменимость»: библиотечный пакет, не использующий flag.
package cleanlib

import "strings"

func Normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
