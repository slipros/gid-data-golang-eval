// Package dataresponse реализует правило GID-163: http handler строится
// на github.com/raoptimus/data-response.go/v2 — чистый golang handler
// func(http.ResponseWriter, *http.Request) запрещён.
//
// Исключения возможны:
//   - точечно: //nolint:giddataresponse
//   - централизованно: settings.exclude — "Функция" или "Тип.Метод".
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

// Analyzer — вариант без исключений.
var Analyzer = NewAnalyzer(Settings{})

// Settings — настройки линтера из .golangci.yml.
type Settings struct {
	// Exclude — хендлеры-исключения: "Функция" или "Тип.Метод".
	Exclude []string `json:"exclude"`
}

// NewAnalyzer строит анализатор GID-163 из настроек линтера (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "giddataresponse",
		Doc:  ruleID + ": http handler использует " + responseLib + ", не чистый golang handler",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, s)
		},
	}
}

func run(pass *analysis.Pass, s Settings) (any, error) {
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
			if !isRawHandler(pass, fn.Type) {
				continue
			}
			if exclude.Match(s.Exclude, recvTypeName(fn), fn.Name.Name) {
				continue
			}
			pass.Reportf(fn.Name.Pos(),
				"%s: %q — чистый golang handler запрещён, используйте %s (исключения: nolint или settings.exclude)",
				ruleID, fn.Name.Name, responseLib)
		}
	}
	return nil, nil
}

// isRawHandler: сигнатура (http.ResponseWriter, *http.Request) без результатов.
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
