// Package grpcinservice implements rule GID-160: a service calls gRPC
// through a repository, not directly. In /domain/service and /domain/usecase
// the following imports are forbidden:
//
//   - google.golang.org/grpc — direct use of connections;
//   - packages that themselves import google.golang.org/grpc —
//     this catches generated pb stubs and gRPC clients.
//
// This rule has exceptions — sometimes gRPC is called directly
// in a service:
//   - pointwise: //nolint:gidgrpcinservice
//   - centrally: settings.exclude — a list of import paths
//     allowed in the domain layer.
package grpcinservice

import (
	"go/ast"
	"go/types"
	"slices"
	"strconv"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const (
	ruleID  = "GID-160"
	grpcPkg = "google.golang.org/grpc"
)

var scopes = [][]string{
	{"domain", "service"},
	{"domain", "usecase"},
}

// Analyzer — variant without exclusions.
var Analyzer = NewAnalyzer(Settings{})

// Settings — linter settings from .golangci.yml.
type Settings struct {
	// Exclude — import paths allowed in the domain layer (rule exceptions).
	Exclude []string `json:"exclude"`
}

// NewAnalyzer builds the GID-160 analyzer from the linter settings (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "gidgrpcinservice",
		Doc:  ruleID + ": a service calls gRPC through a repository, not directly. Fix: move the gRPC call into a repository",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, s)
		},
	}
}

func run(pass *analysis.Pass, s Settings) (any, error) {
	if !inScope(pass.Pkg.Path()) {
		return nil, nil
	}
	grpcBacked := grpcBackedImports(pass.Pkg)
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, imp := range file.Imports {
			path, err := strconv.Unquote(imp.Path.Value)
			if err != nil || slices.Contains(s.Exclude, path) {
				continue
			}
			switch {
			case path == grpcPkg:
				pass.Reportf(imp.Pos(),
					"%s: direct import of %s in the domain layer is forbidden. Fix: call gRPC through a repository "+
						"(exceptions: nolint or settings.exclude)",
					ruleID, grpcPkg)
			case grpcBacked[path]:
				pass.Reportf(imp.Pos(),
					"%s: importing the gRPC package %q in the domain layer is forbidden. Fix: call gRPC through a repository "+
						"(exceptions: nolint or settings.exclude)",
					ruleID, path)
			}
		}
	}
	return nil, nil
}

// grpcBackedImports — the package's direct imports that themselves import
// google.golang.org/grpc (pb stubs, gRPC clients).
func grpcBackedImports(pkg *types.Package) map[string]bool {
	out := map[string]bool{}
	for _, imp := range pkg.Imports() {
		for _, sub := range imp.Imports() {
			if sub.Path() == grpcPkg {
				out[imp.Path()] = true
				break
			}
		}
	}
	return out
}

func inScope(pkgPath string) bool {
	for _, scope := range scopes {
		if pathseg.HasLayer(pkgPath, scope...) {
			return true
		}
	}
	return false
}
