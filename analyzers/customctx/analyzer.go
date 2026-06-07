// Package customctx реализует правило GID-188: запрет кастомных
// context-типов. По решению Google ("custom contexts — no exceptions")
// в позиции ctx-параметра и в embedding интерфейсов допустим только
// stdlib-тип context.Context. Данные передаются через context.WithValue
// (хелперы живут в /domain/model — GID-165/166).
package customctx

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-188"

// Analyzer — правило GID-188: запрет кастомных context-типов — только context.Context.
var Analyzer = &analysis.Analyzer{
	Name: "gidcustomctx",
	Doc:  ruleID + ": custom context types are forbidden, use context.Context. Fix: pass context.Context and store data via context.WithValue.",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	ctxIface := lookupContextInterface(pass)

	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				checkTypeDecls(pass, d, ctxIface)
			case *ast.FuncDecl:
				checkFuncParams(pass, d.Type)
			}
		}

		// Параметры функциональных литералов и типов функций тоже проверяем.
		ast.Inspect(file, func(n ast.Node) bool {
			if lit, ok := n.(*ast.FuncLit); ok {
				checkFuncParams(pass, lit.Type)
			}
			return true
		})
	}
	return nil, nil
}

// lookupContextInterface возвращает базовый интерфейс stdlib-типа
// context.Context, если пакет context импортируется (прямо или
// транзитивно). Иначе nil — проверки 1 и 2 неприменимы.
func lookupContextInterface(pass *analysis.Pass) *types.Interface {
	for _, imp := range allImports(pass.Pkg) {
		if imp.Path() != "context" {
			continue
		}
		scope := imp.Scope()
		obj := scope.Lookup("Context")
		if obj == nil {
			return nil
		}
		named, ok := obj.Type().(*types.Named)
		if !ok {
			return nil
		}
		iface, ok := named.Underlying().(*types.Interface)
		if !ok {
			return nil
		}
		return iface
	}
	return nil
}

func allImports(pkg *types.Package) []*types.Package {
	seen := map[string]bool{}
	var out []*types.Package
	var walk func(p *types.Package)
	walk = func(p *types.Package) {
		for _, imp := range p.Imports() {
			if seen[imp.Path()] {
				continue
			}
			seen[imp.Path()] = true
			out = append(out, imp)
			walk(imp)
		}
	}
	walk(pkg)
	return out
}

// isStdlibContext сообщает, является ли тип именно stdlib context.Context.
func isStdlibContext(t types.Type) bool {
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	pkg := obj.Pkg()
	return pkg != nil && pkg.Path() == "context" && obj.Name() == "Context"
}

// checkTypeDecls проверяет объявления типов в проверяемом пакете:
//   - именованный тип (struct/interface/любой), method set которого
//     покрывает context.Context (кейс 1);
//   - interface-тип, встраивающий context.Context (кейс 2).
func checkTypeDecls(pass *analysis.Pass, gen *ast.GenDecl, ctxIface *types.Interface) {
	for _, spec := range gen.Specs {
		ts, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}

		// Кейс 2: interface, встраивающий context.Context.
		if iface, ok := ts.Type.(*ast.InterfaceType); ok && embedsStdlibContext(pass, iface) {
			pass.Reportf(ts.Pos(),
				"%s: custom context type %s is forbidden. "+
					"Fix: pass context.Context and store data via context.WithValue "+
					"(helpers live in /domain/model, GID-165/166).",
				ruleID, ts.Name.Name)
			continue
		}

		// Кейс 1: тип, реализующий context.Context по method set.
		if ctxIface == nil {
			continue
		}
		obj := pass.TypesInfo.Defs[ts.Name]
		if obj == nil {
			continue
		}
		named, ok := obj.Type().(*types.Named)
		if !ok {
			continue
		}
		// Сам stdlib context.Context не считаем (он не из нашего пакета,
		// но на всякий случай не самоссылаемся).
		if isStdlibContext(named) {
			continue
		}
		if types.Implements(named, ctxIface) || types.Implements(types.NewPointer(named), ctxIface) {
			pass.Reportf(ts.Pos(),
				"%s: custom context type %s is forbidden. "+
					"Fix: pass context.Context and store data via context.WithValue "+
					"(helpers live in /domain/model, GID-165/166).",
				ruleID, ts.Name.Name)
		}
	}
}

// embedsStdlibContext проверяет, встраивает ли interface-декларация
// stdlib context.Context (embedded-поле без имени).
func embedsStdlibContext(pass *analysis.Pass, iface *ast.InterfaceType) bool {
	if iface.Methods == nil {
		return false
	}
	for _, field := range iface.Methods.List {
		if len(field.Names) != 0 {
			continue // обычный метод, не embedding
		}
		if isStdlibContext(pass.TypesInfo.TypeOf(field.Type)) {
			return true
		}
	}
	return false
}

// checkFuncParams проверяет кейс 3: параметр с именем ctx, чей тип —
// именованный не-stdlib context-тип.
func checkFuncParams(pass *analysis.Pass, ft *ast.FuncType) {
	if ft == nil || ft.Params == nil {
		return
	}
	for _, field := range ft.Params.List {
		hasCtxName := false
		for _, name := range field.Names {
			if name.Name == "ctx" {
				hasCtxName = true
				break
			}
		}
		if !hasCtxName {
			continue
		}

		t := pass.TypesInfo.TypeOf(field.Type)
		if t == nil || isStdlibContext(t) {
			continue
		}
		// Интересует только именованный тип (не stdlib). Анонимные
		// типы/встроенные — не наш случай.
		if _, ok := t.(*types.Named); !ok {
			continue
		}
		pass.Reportf(field.Type.Pos(),
			"%s: parameter ctx has type %s. Fix: use context.Context.",
			ruleID, t.String())
	}
}
