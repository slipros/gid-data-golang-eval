// Eval for GID-196 with the min-calls threshold: 3 — short inline chains are allowed.
package threshold

import "strings"

func two() string {
	return strings.NewReplacer("a", "b").Replace("aa") // 2 links < threshold — OK
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
	return b.sel("id").from("snapshots").build() // want `GID-196: a chain of 3 calls must put one call per line, including the first\. Fix: break each \.Method\(\) onto its own line\.`
}
