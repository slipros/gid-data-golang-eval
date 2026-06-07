// Package modelmethod реализует правило GID-195: приватная функция в
// service/usecase, принимающая единственное значение model-типа и не
// зависящая от своего пакета, — это поведение модели: её место — публичный
// метод этого типа в model-слое.
//
// Тот же случай — приватный метод service/usecase-структуры, который не
// использует ресивер. Непереносимые кандидаты не задеваются: метод,
// использующий ресивер; функция, ссылающаяся на package-level символы
// своего пакета (включая типы пакета в результатах).
package modelmethod

import (
	"go/ast"
	"go/types"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/exclude"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-195"

// scopes — слои, где действует правило. Только корни слоя (pathseg.EndsWith):
// подпакеты convert/ и т.п. не задеваются.
var scopes = [][]string{
	{"domain", "service"},
	{"domain", "usecase"},
}

// Analyzer — правило GID-195 с настройками по умолчанию.
var Analyzer = NewAnalyzer(Settings{})

// Settings — настройки правила GID-195 из .golangci.yml.
type Settings struct {
	// Exclude — исключения: "Функция" или "Тип.Метод".
	Exclude []string `json:"exclude"`
}

// NewAnalyzer строит анализатор GID-195 из настроек линтера (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "gidmodelmethod",
		Doc: ruleID + ": a private service/usecase function over a single model value is " +
			"model behaviour; expose it as a public method of that type. Fix: move it onto the model",
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
		if ast.IsGenerated(file) || isTestFile(pass, file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			checkFunc(pass, fn, s)
		}
	}
	return nil, nil
}

func checkFunc(pass *analysis.Pass, fn *ast.FuncDecl, s Settings) {
	if fn.Name.IsExported() || fn.Name.Name == "init" || fn.Type.TypeParams != nil {
		return
	}
	if exclude.Match(s.Exclude, recvTypeName(fn), fn.Name.Name) {
		return
	}
	param, ok := singleModelParam(pass, fn)
	if !ok {
		return
	}
	// Метод, использующий ресивер, непереносим — он легитимно
	// принадлежит своей структуре.
	if fn.Recv != nil && usesReceiver(pass, fn) {
		return
	}
	// Зависимость от package-level символов своего пакета (включая типы
	// пакета в сигнатуре) — функцию нельзя перенести в model.
	if dependsOnPackage(pass, fn) {
		return
	}
	paramObj := param.Obj()
	paramPkg := paramObj.Pkg()
	display := paramPkg.Name() + "." + paramObj.Name()
	if fn.Recv != nil {
		pass.Reportf(fn.Name.Pos(),
			"%s: method %q ignores its receiver and works only with the %s value. "+
				"Fix: this is model behaviour, make it a public method of that type",
			ruleID, fn.Name.Name, display)
		return
	}
	pass.Reportf(fn.Name.Pos(),
		"%s: private function %q works only with the %s value. "+
			"Fix: this is model behaviour, make it a public method of that type",
		ruleID, fn.Name.Name, display)
}

// singleModelParam — единственный параметр функции вида T или *T,
// где T — именованный тип model-слоя (struct, enum и т.п., не интерфейс).
func singleModelParam(pass *analysis.Pass, fn *ast.FuncDecl) (*types.Named, bool) {
	params := fn.Type.Params
	if params == nil || len(params.List) != 1 {
		return nil, false
	}
	field := params.List[0]
	// func f(a, b *model.T) — два значения; variadic — слайс значений.
	if len(field.Names) > 1 {
		return nil, false
	}
	if _, ok := field.Type.(*ast.Ellipsis); ok {
		return nil, false
	}
	t := pass.TypesInfo.TypeOf(field.Type)
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	named, ok := types.Unalias(t).(*types.Named)
	if !ok {
		return nil, false
	}
	// Интерфейсу метод не добавить — он не «владеет» поведением.
	if _, ok := named.Underlying().(*types.Interface); ok {
		return nil, false
	}
	obj := named.Obj()
	pkg := obj.Pkg()
	if pkg == nil || !pathseg.Contains(pkg.Path(), "domain", "model") {
		return nil, false
	}
	return named, true
}

// usesReceiver сообщает, обращается ли тело метода к ресиверу.
func usesReceiver(pass *analysis.Pass, fn *ast.FuncDecl) bool {
	if len(fn.Recv.List) == 0 || len(fn.Recv.List[0].Names) == 0 {
		return false // безымянный ресивер
	}
	recv := fn.Recv.List[0].Names[0]
	if recv.Name == "_" {
		return false
	}
	obj := pass.TypesInfo.Defs[recv]
	if obj == nil || fn.Body == nil {
		return false
	}
	used := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok && pass.TypesInfo.Uses[id] == obj {
			used = true
		}
		return !used
	})
	return used
}

// dependsOnPackage сообщает, ссылается ли функция (сигнатура и тело)
// на package-level символы своего пакета — такая функция непереносима.
func dependsOnPackage(pass *analysis.Pass, fn *ast.FuncDecl) bool {
	self := pass.TypesInfo.Defs[fn.Name]
	depends := false
	check := func(n ast.Node) {
		ast.Inspect(n, func(node ast.Node) bool {
			id, ok := node.(*ast.Ident)
			if !ok {
				return true
			}
			obj := pass.TypesInfo.Uses[id]
			if obj == nil || obj == self || obj.Pkg() != pass.Pkg {
				return true
			}
			switch obj.(type) {
			case *types.PkgName, *types.Label:
				return true // импорты и метки — не зависимость
			}
			// Package-level символ (Parent == scope пакета) либо член
			// типа пакета — поле/метод (Parent == nil).
			if obj.Parent() == pass.Pkg.Scope() || obj.Parent() == nil {
				depends = true
			}
			return !depends
		})
	}
	check(fn.Type)
	if fn.Body != nil {
		check(fn.Body)
	}
	return depends
}

func inScope(pkgPath string) bool {
	for _, scope := range scopes {
		if pathseg.EndsWith(pkgPath, scope...) {
			return true
		}
	}
	return false
}

func recvTypeName(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
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

func isTestFile(pass *analysis.Pass, file *ast.File) bool {
	return strings.HasSuffix(filepath.Base(pass.Fset.Position(file.Pos()).Filename), "_test.go")
}
