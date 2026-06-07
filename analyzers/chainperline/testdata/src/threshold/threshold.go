// Eval для GID-196 с порогом min-calls: 3 — короткие цепочки inline допустимы.
package threshold

import "strings"

func two() string {
	return strings.NewReplacer("a", "b").Replace("aa") // 2 звена < порога — нормы
}

func three() string {
	return strings.NewReplacer("a", "b").Replace("aa") + builderChain()
}

type builder struct{}

func (b builder) sel(s string) builder  { return b }
func (b builder) from(s string) builder { return b }
func (b builder) build() string         { return "" }

func builderChain() string {
	b := builder{}
	return b.sel("id").from("snapshots").build() // want `GID-196: цепочка из 3 вызовов оформляется по одному вызову на строке, включая первый`
}
