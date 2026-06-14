// "Inapplicability" class: a library package that does not use flag.
package cleanlib

import "strings"

func Normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
