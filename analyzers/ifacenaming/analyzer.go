// Package ifacenaming implements rule GID-173 (iface-entity-prefix):
//
//   - GID-173 (gidifacenaming): dependency interfaces are named with an
//     entity prefix (`HelloRepository`, `HelloConnection`). A bare role name
//     (`Repository`, `Connection`, …) is forbidden: it does not reveal which
//     entity the interface is a dependency of.
//
// Scope: packages in the layers /domain/service, /domain/usecase, /dal/repository,
// /server/**, /event/** (pathseg.HasLayer — anchored to the module root).
// Declarations of interface types
// whose name EXACTLY matches a bare role from the dictionary are checked.
// Generated code is skipped.
package ifacenaming

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-173"

// Analyzer — variant with the default role dictionary.
var Analyzer = NewAnalyzer(Settings{})

// Settings — linter settings from .golangci.yml.
type Settings struct {
	// Names — the dictionary of bare roles (replaces the default list).
	Names []string `json:"names"`
}

// NewAnalyzer builds the GID-173 analyzer from the linter settings (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	names := resolveNames(s)
	roles := make(map[string]struct{}, len(names))
	for _, n := range names {
		roles[n] = struct{}{}
	}
	return &analysis.Analyzer{
		Name: "gidifacenaming",
		Doc:  ruleID + ": dependency interfaces are named with an entity prefix (e.g. HelloRepository). Fix: prefix the interface with the entity name",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, roles)
		},
	}
}

func resolveNames(s Settings) []string {
	if len(s.Names) == 0 {
		return []string{
			"Repository",
			"Service",
			"Client",
			"Connection",
			"Producer",
			"Consumer",
			"Validator",
			"Storage",
			"Cache",
		}
	}
	return s.Names
}

func run(pass *analysis.Pass, roles map[string]struct{}) (any, error) {
	if !inScope(pass.Pkg.Path()) {
		return nil, nil
	}
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if _, ok := ts.Type.(*ast.InterfaceType); !ok {
					continue // not an interface — the rule does not apply
				}
				if _, ok := roles[ts.Name.Name]; !ok {
					continue // the name does not exactly match a bare role
				}
				pass.Reportf(ts.Name.Pos(),
					"%s: interface %q must be named with an entity prefix. Fix: e.g. HelloRepository",
					ruleID, ts.Name.Name)
			}
		}
	}
	return nil, nil
}

// inScope reports whether the package belongs to a layer where the rule
// applies. The layer is anchored to the module root (pathseg.HasLayer): a
// segment nested below another layer, e.g. .../dal/entity/event/…, is NOT
// that layer.
func inScope(path string) bool {
	return pathseg.HasLayer(path, "domain", "service") ||
		pathseg.HasLayer(path, "domain", "usecase") ||
		pathseg.HasLayer(path, "dal", "repository") ||
		pathseg.HasLayer(path, "server") ||
		pathseg.HasLayer(path, "event")
}
