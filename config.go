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

// rulesOnlyConfig — the same ruleset minus the stock linters and the formatter,
// for a repository that already runs a golangci-lint of its own: there the ~40
// stock linters of distConfig are enabled a second time, and the pair costs
// roughly twice what it should (lk-api, 1681 files: 17.0 s against 1.12 s, for
// identical gid diagnostics). Selected with --gid-rules-only.
//
//go:embed gid-golangci-rules.yml
var rulesOnlyConfig []byte

// DefaultConfig returns the golangci-lint config built into the binary.
func DefaultConfig() []byte {
	return slices.Clone(distConfig)
}

// RulesOnlyConfig returns the built-in config that enables the gid rules alone.
func RulesOnlyConfig() []byte {
	return slices.Clone(rulesOnlyConfig)
}
