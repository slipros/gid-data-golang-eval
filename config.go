package gidrules

import (
	_ "embed"
	"slices"
)

// distConfig — the distributable config for gid.team services, the same file a
// service copies into its repository as .golangci.yml. It is embedded so that
// custom-gcl carries a working ruleset on its own: a binary and its config can
// no longer drift apart, and a repository without .golangci.yml is still linted
// by the gid rules (see internal/defaultconfig).
//
//go:embed gid-golangci.yml
var distConfig []byte

// DefaultConfig returns the golangci-lint config built into the binary.
func DefaultConfig() []byte {
	return slices.Clone(distConfig)
}
