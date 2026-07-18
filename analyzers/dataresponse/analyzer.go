// Package dataresponse implements rule GID-163: an http handler is built
// on github.com/raoptimus/data-response.go/v2 — a plain golang handler
// func(http.ResponseWriter, *http.Request) is forbidden.
//
// Exceptions are possible:
//   - targeted: //nolint:giddataresponse
//   - centralized: settings.exclude — "Function" or "Type.Method".
package dataresponse

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/exclude"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const (
	ruleID      = "GID-163"
	responseLib = "github.com/raoptimus/data-response.go/v2"
)

// Analyzer — the variant without exclusions.
var Analyzer = NewAnalyzer(Settings{})

// Settings — linter settings from .golangci.yml.
type Settings struct {
	// Exclude — excluded handlers: "Function" or "Type.Method".
	Exclude []string `json:"exclude"`
}

// NewAnalyzer builds the GID-163 analyzer from the linter settings (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "giddataresponse",
		Doc:  ruleID + ": an http handler must use " + responseLib + ", not a plain golang handler",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, s)
		},
	}
}

func run(pass *analysis.Pass, s Settings) (any, error) {
	if !pathseg.HasLayer(pass.Pkg.Path(), "server", "http") {
		return nil, nil
	}
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			if !isRawHandler(pass, fn.Type) {
				continue
			}
			if exclude.Match(s.Exclude, recvTypeName(fn), fn.Name.Name) {
				continue
			}
			pass.Reportf(fn.Name.Pos(),
				"%s: %q is a plain golang handler, which is forbidden. Fix: use %s (exceptions: nolint or settings.exclude)",
				ruleID, fn.Name.Name, responseLib)
		}
	}
	return nil, nil
}

// isRawHandler: the signature (http.ResponseWriter, *http.Request) without results.
func isRawHandler(pass *analysis.Pass, fnType *ast.FuncType) bool {
	if fnType.Results != nil && len(fnType.Results.List) > 0 {
		return false
	}
	params := fnType.Params.List
	if len(params) != 2 || len(params[0].Names) > 1 || len(params[1].Names) > 1 {
		return false
	}
	return isHTTPType(pass.TypesInfo.TypeOf(params[0].Type), "ResponseWriter", false) &&
		isHTTPType(pass.TypesInfo.TypeOf(params[1].Type), "Request", true)
}

func isHTTPType(t types.Type, name string, wantPtr bool) bool {
	if wantPtr {
		ptr, ok := t.(*types.Pointer)
		if !ok {
			return false
		}
		t = ptr.Elem()
	}
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	pkg := obj.Pkg()
	return pkg != nil && pkg.Path() == "net/http" && obj.Name() == name
}

func recvTypeName(fn *ast.FuncDecl) string {
	if fn.Recv == nil {
		return ""
	}
	t := fn.Recv.List[0].Type
	if star, ok := t.(*ast.StarExpr); ok {
		t = star.X
	}
	if ident, ok := t.(*ast.Ident); ok {
		return ident.Name
	}
	return ""
}
