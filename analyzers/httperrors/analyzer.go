// Package httperrors implements rule GID-162: an http handler always
// handles its own errors inside itself.
//
//   - "super-methods" that accept all errors and handle them universally
//     are forbidden (the marker: http.ResponseWriter + error parameters);
//   - a handler function (http.ResponseWriter, *http.Request) does not
//     return an error — the error is handled in place.
package httperrors

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-162"

// Analyzer — rule GID-162: an http handler handles its own errors inside itself.
var Analyzer = &analysis.Analyzer{
	Name: "gidhttperrors",
	Doc:  ruleID + ": an http handler handles its own errors inline, without super-methods. Fix: handle errors inside the handler",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
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
			checkFunc(pass, fn)
		}
	}
	return nil, nil
}

func checkFunc(pass *analysis.Pass, fn *ast.FuncDecl) {
	hasRW, hasErrParam := false, false
	for _, field := range fn.Type.Params.List {
		t := pass.TypesInfo.TypeOf(field.Type)
		switch {
		case isHTTPResponseWriter(t):
			hasRW = true
		case isErrorType(t):
			hasErrParam = true
		}
	}
	if hasRW && hasErrParam {
		pass.Reportf(fn.Name.Pos(),
			"%s: %q is a forbidden error-handling super-method. Fix: handle errors inside each http handler",
			ruleID, fn.Name.Name)
		return
	}
	if hasRW && fn.Type.Results != nil {
		for _, field := range fn.Type.Results.List {
			if isErrorType(pass.TypesInfo.TypeOf(field.Type)) {
				pass.Reportf(fn.Name.Pos(),
					"%s: http handler %q must not return error. Fix: handle the error in place",
					ruleID, fn.Name.Name)
				return
			}
		}
	}
}

func isHTTPResponseWriter(t types.Type) bool {
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	pkg := obj.Pkg()
	return pkg != nil && pkg.Path() == "net/http" && obj.Name() == "ResponseWriter"
}

func isErrorType(t types.Type) bool {
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	return obj.Pkg() == nil && obj.Name() == "error"
}
