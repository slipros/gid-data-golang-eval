// Package grpcinservice реализует правило GID-160: service вызывает gRPC
// через repository, а не напрямую. В /domain/service и /domain/usecase
// запрещены импорты:
//
//   - google.golang.org/grpc — прямое использование соединений;
//   - пакетов, которые сами импортируют google.golang.org/grpc —
//     это ловит сгенерированные pb-стабы и gRPC-клиенты.
//
// Для этого правила бывают исключения — иногда gRPC вызывается прямо
// в service:
//   - точечно: //nolint:gidgrpcinservice
//   - централизованно: settings.exclude — список import-путей,
//     разрешённых в domain-слое.
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

// Analyzer — вариант без исключений.
var Analyzer = NewAnalyzer(Settings{})

// Settings — настройки линтера из .golangci.yml.
type Settings struct {
	// Exclude — import-пути, разрешённые в domain-слое (исключения правила).
	Exclude []string `json:"exclude"`
}

// NewAnalyzer строит анализатор GID-160 из настроек линтера (.golangci.yml).
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

// grpcBackedImports — прямые импорты пакета, которые сами импортируют
// google.golang.org/grpc (pb-стабы, gRPC-клиенты).
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
		if pathseg.Contains(pkgPath, scope...) {
			return true
		}
	}
	return false
}
