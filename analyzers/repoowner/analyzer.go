// Package repoowner enforces one service owner per concrete repository type.
package repoowner

import (
	"go/types"
	"slices"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/astwalk"
	"github.com/slipros/gid-data-golang-eval/internal/dependencywiring"
)

// Analyzer checks GID-280 with default settings.
var Analyzer = NewAnalyzer(Settings{})

// Settings provides targeted exceptions by fully qualified repository type.
type Settings struct {
	Exclude []string `json:"exclude"`
}

// NewAnalyzer builds the repository ownership check.
func NewAnalyzer(s Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "gidrepoowner",
		Doc:      "GID-280: a repository belongs to one service. Fix: move its operations into one owning service and orchestrate through usecase",
		Requires: astwalk.Requires,
		Run:      func(pass *analysis.Pass) (any, error) { return run(pass, s) },
	}
}

func run(pass *analysis.Pass, s Settings) (any, error) {
	owners := make(map[string]string)
	for _, edge := range dependencywiring.Edges(pass) {
		if !dependencywiring.InLayer(edge.Dependency, "dal", "repository") {
			continue
		}
		repo := types.TypeString(edge.Dependency, nil)
		if slices.Contains(s.Exclude, repo) {
			continue
		}
		service := types.TypeString(edge.Service, nil)
		if owner, ok := owners[repo]; ok && owner != service {
			pass.Reportf(edge.Pos, "GID-280: repository %q is supplied to services %q and %q. Fix: give the repository one owning service and compose its operations in usecase", repo, owner, service)
			continue
		}
		owners[repo] = service
	}
	return nil, nil
}
