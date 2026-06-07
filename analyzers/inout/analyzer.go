// Package inout реализует правило GID-111: входные данные (model/entity
// структуры) передаются по указателю, выходные возвращаются по значению.
// Действует на слоях repo, service, usecase и handler.
//
// Исключения:
//   - точечно: //nolint:gidinout
//   - централизованно: settings.exclude в .golangci.yml —
//     записи вида "Метод" или "Тип.Метод".
package inout

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/exclude"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-111"

var scopes = [][]string{
	{"dal", "repository"},
	{"domain", "service"},
	{"domain", "usecase"},
	{"handler"},
}

// layerTypeTrees — деревья пакетов, чьи структуры считаются данными слоёв.
var layerTypeTrees = [][]string{
	{"domain", "model"},
	{"dal", "entity"},
}

// Analyzer — вариант с настройками по умолчанию (без исключений).
var Analyzer = NewAnalyzer(Settings{})

// Settings — настройки линтера из .golangci.yml.
type Settings struct {
	// Exclude — методы-исключения: "Метод" или "Тип.Метод".
	Exclude []string `json:"exclude"`
}

// NewAnalyzer строит анализатор GID-111 из настроек линтера (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "gidinout",
		Doc:  ruleID + ": входные model/entity-структуры по указателю, выходные по значению",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, s)
		},
	}
}

func run(pass *analysis.Pass, s Settings) (any, error) {
	if !inScope(pass.Pkg.Path()) {
		return nil, nil
	}
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || !fn.Name.IsExported() {
				continue
			}
			if exclude.Match(s.Exclude, recvTypeName(fn), fn.Name.Name) {
				continue
			}
			checkSignature(pass, fn)
		}
	}
	return nil, nil
}

func checkSignature(pass *analysis.Pass, fn *ast.FuncDecl) {
	if fn.Type.Params != nil {
		for _, field := range fn.Type.Params.List {
			if name, ok := layerStructName(pass.TypesInfo.TypeOf(field.Type)); ok {
				pass.Reportf(field.Pos(),
					"%s: входные данные передаются по указателю — *%s", ruleID, name)
			}
		}
	}
	if fn.Type.Results != nil {
		for _, field := range fn.Type.Results.List {
			t := pass.TypesInfo.TypeOf(field.Type)
			ptr, ok := t.(*types.Pointer)
			if !ok {
				continue
			}
			if name, ok := layerStructName(ptr.Elem()); ok {
				pass.Reportf(field.Pos(),
					"%s: выходные данные возвращаются по значению — %s", ruleID, name)
			}
		}
	}
}

// layerStructName возвращает имя типа, если это структура из
// /domain/model или /dal/entity (по значению, без указателя).
func layerStructName(t types.Type) (string, bool) {
	named, ok := t.(*types.Named)
	if !ok {
		return "", false
	}
	if _, ok := named.Underlying().(*types.Struct); !ok {
		return "", false
	}
	obj := named.Obj()
	if obj.Pkg() == nil {
		return "", false
	}
	pkg := obj.Pkg()
	for _, tree := range layerTypeTrees {
		if pathseg.Contains(pkg.Path(), tree...) {
			return pkg.Name() + "." + obj.Name(), true
		}
	}
	return "", false
}

func inScope(pkgPath string) bool {
	for _, scope := range scopes {
		if pathseg.Contains(pkgPath, scope...) {
			return true
		}
	}
	return false
}

func recvTypeName(fn *ast.FuncDecl) string {
	t := fn.Recv.List[0].Type
	if star, ok := t.(*ast.StarExpr); ok {
		t = star.X
	}
	if ident, ok := t.(*ast.Ident); ok {
		return ident.Name
	}
	return ""
}
