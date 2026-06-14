// Test files are out of scope for GID-196: inline chains are allowed.
package chains

import "strings"

func helperChain() string {
	return strings.NewReplacer("a", "b").Replace("aa")
}

var _ = helperChain
