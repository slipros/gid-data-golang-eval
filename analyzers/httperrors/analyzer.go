// Package httperrors реализует правило GID-162: http handler всегда
// обрабатывает свои ошибки внутри себя.
//
//   - запрещены «супер-методы», принимающие все ошибки и универсально
//     их обрабатывающие (признак: параметры http.ResponseWriter + error);
//   - handler-функция (http.ResponseWriter, *http.Request) не возвращает
//     error — ошибка обрабатывается на месте.
package httperrors

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-162"

// Analyzer — правило GID-162: http handler обрабатывает свои ошибки внутри себя.
var Analyzer = &analysis.Analyzer{
	Name: "gidhttperrors",
	Doc:  ruleID + ": http handler обрабатывает свои ошибки внутри себя, без супер-методов",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	if !pathseg.Contains(pass.Pkg.Path(), "server", "http") {
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
			"%s: %q — супер-метод обработки ошибок запрещён, http handler обрабатывает ошибки внутри себя",
			ruleID, fn.Name.Name)
		return
	}
	if hasRW && fn.Type.Results != nil {
		for _, field := range fn.Type.Results.List {
			if isErrorType(pass.TypesInfo.TypeOf(field.Type)) {
				pass.Reportf(fn.Name.Pos(),
					"%s: http handler %q не возвращает error — ошибка обрабатывается на месте",
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
