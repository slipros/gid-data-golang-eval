// Тестовые файлы вне scope GID-196: цепочки inline допустимы.
package chains

import "strings"

func helperChain() string {
	return strings.NewReplacer("a", "b").Replace("aa")
}

var _ = helperChain
