// Package errwrap реализует правила обработки ошибок по слоям:
//
//   - GID-176 (giderrwrap): ошибки извне оборачиваются errors.Wrap.
//     На границе приложения (/client/** и /dal/repository) ошибка из
//     внешнего вызова не пробрасывается как есть (return err) и не
//     обогащается без контекста (WithStack/WithMessage) — нужен Wrap:
//     он собирает стек И добавляет обязательный контекст. Внутри
//     приложения (/domain/**) для уже пришедшей нестатичной ошибки
//     Wrap запрещён (стек собран на границе) — контекст добавляется
//     WithMessage. Возврат статичной ошибки (package-level var, именованный
//     error-тип) на границе — не нарушение GID-176 (это зона GID-177).
//
//   - GID-177 (gidstaticerr): статичные ошибки оборачиваются WithStack.
//     Возврат статичной ошибки (package-level error-var ErrSome или
//     композит-литерал/адрес именованного error-типа BigError{}/&BigError{})
//     без обёртки лишён стека — нужен errors.WithStack (или errors.Wrap,
//     если требуется контекст). Обёрнутая ошибка (WithStack/Wrap) — ок.
//     Объявления самих var не задеваются (это не return).
//
// pkg/errors определяется по import-пути github.com/pkg/errors.
// Сгенерированный код (ast.IsGenerated) пропускается.
package errwrap

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"

	"github.com/slipros/gid-data-golang-eval/internal/exclude"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const (
	ruleIDWrap   = "GID-176"
	ruleIDStatic = "GID-177"
)

// boundaryScopes — граничные слои для GID-176 (часть 1): внешний вызов.
var boundaryScopes = [][]string{
	{"client"},
	{"dal", "repository"},
}

// WrapAnalyzer — GID-176 с настройками по умолчанию (без исключений).
var WrapAnalyzer = NewWrapAnalyzer(Settings{})

// StaticAnalyzer — GID-177 с настройками по умолчанию (без исключений).
var StaticAnalyzer = NewStaticAnalyzer(Settings{})

// Settings — настройки линтеров из .golangci.yml.
type Settings struct {
	// Exclude — имена конструкторов/ошибок-исключений, которые сами
	// собирают стек (например, gderror.NewUnhandledValueError):
	// "Функция" или "Пакет.Функция".
	Exclude []string `json:"exclude"`
}

// NewWrapAnalyzer строит анализатор GID-176.
func NewWrapAnalyzer(s Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "giderrwrap",
		Doc: ruleIDWrap + ": ошибки извне оборачиваются errors.Wrap; " +
			"внутри приложения нестатичную ошибку не Wrap, а WithMessage",
		Run: runWrap,
	}
}

// NewStaticAnalyzer строит анализатор GID-177.
func NewStaticAnalyzer(s Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "gidstaticerr",
		Doc:  ruleIDStatic + ": статичные ошибки при возврате оборачиваются errors.WithStack",
		Run: func(pass *analysis.Pass) (any, error) {
			return runStatic(pass, s)
		},
	}
}

// ===== GID-176 =====

func runWrap(pass *analysis.Pass) (any, error) {
	pkgPath := pass.Pkg.Path()
	boundary := inBoundary(pkgPath)
	domain := pathseg.Contains(pkgPath, "domain")
	if !boundary && !domain {
		return nil, nil
	}
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			if boundary && funcReturnsError(pass, fn) {
				checkBoundaryPassThrough(pass, fn)
			}
		}
		if domain {
			checkDomainWrap(pass, file)
		}
	}
	return nil, nil
}

// checkBoundaryPassThrough — GID-176 часть 1: на границе нельзя
// пробрасывать нестатичную ошибку из вызова без Wrap.
func checkBoundaryPassThrough(pass *analysis.Pass, fn *ast.FuncDecl) {
	callErrs := localCallErrors(pass, fn)
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		ret, ok := n.(*ast.ReturnStmt)
		if !ok {
			return true
		}
		for _, res := range ret.Results {
			expr := res
			// errors.WithStack(err) / errors.WithMessage(err) — обёртка без контекста.
			if call, ok := expr.(*ast.CallExpr); ok {
				name := pkgErrorsCallName(pass, call)
				if name == "WithStack" || name == "WithMessage" {
					if len(call.Args) > 0 && isLocalCallErr(pass, call.Args[0], callErrs) {
						pass.Reportf(call.Pos(),
							"%s: ошибка с границы приложения оборачивается errors.Wrap — "+
								"собрать стек и контекст (%s контекста не добавляет)",
							ruleIDWrap, name)
					}
					continue
				}
				// errors.Wrap / иной вызов — ок (Wrap уже правильный).
				continue
			}
			if isLocalCallErr(pass, expr, callErrs) {
				pass.Reportf(expr.Pos(),
					"%s: оберните errors.Wrap — ошибка с границы приложения должна собрать стек и контекст",
					ruleIDWrap)
			}
		}
		return true
	})
}

// checkDomainWrap — GID-176 часть 2: в /domain/** Wrap нестатичной ошибки запрещён.
func checkDomainWrap(pass *analysis.Pass, file *ast.File) {
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if pkgErrorsCallName(pass, call) != "Wrap" {
			return true
		}
		if len(call.Args) == 0 {
			return true
		}
		// Статичная ошибка (model.ErrX, &BigError{}) — Wrap разрешён.
		if isStaticError(pass, call.Args[0]) {
			return true
		}
		// Нестатичная (локальная переменная из вызова и т.п.) — запрещён.
		if isErrorExpr(pass, call.Args[0]) {
			pass.Reportf(call.Pos(),
				"%s: стек уже собран на границе — используйте errors.WithMessage вместо errors.Wrap для пришедшей ошибки",
				ruleIDWrap)
		}
		return true
	})
}

// ===== GID-177 =====

func runStatic(pass *analysis.Pass, s Settings) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				ret, ok := n.(*ast.ReturnStmt)
				if !ok {
					return true
				}
				for _, res := range ret.Results {
					checkStaticReturn(pass, s, res)
				}
				return true
			})
		}
	}
	return nil, nil
}

func checkStaticReturn(pass *analysis.Pass, s Settings, expr ast.Expr) {
	// Уже обёрнуто (WithStack/Wrap/иной вызов pkg/errors) — ок.
	if call, ok := expr.(*ast.CallExpr); ok {
		name := pkgErrorsCallName(pass, call)
		if name == "WithStack" || name == "Wrap" || name == "Wrapf" {
			return
		}
		// Конструктор-исключение (сам собирает стек) — ок.
		if isExcludedCtor(pass, call, s.Exclude) {
			return
		}
		return
	}
	if isStaticError(pass, expr) {
		pass.Reportf(expr.Pos(),
			"%s: статичная ошибка возвращается без стека — оберните errors.WithStack (или errors.Wrap, если нужен контекст)",
			ruleIDStatic)
	}
}

// ===== общие хелперы =====

func inBoundary(pkgPath string) bool {
	for _, scope := range boundaryScopes {
		if pathseg.Contains(pkgPath, scope...) {
			return true
		}
	}
	return false
}

// pkgErrorsCallName возвращает имя функции github.com/pkg/errors,
// если call — её вызов; иначе "".
func pkgErrorsCallName(pass *analysis.Pass, call *ast.CallExpr) string {
	const pkgErrorsPath = "github.com/pkg/errors"
	fn := typeutil.Callee(pass.TypesInfo, call)
	f, ok := fn.(*types.Func)
	if !ok || f.Pkg() == nil {
		return ""
	}
	pkg := f.Pkg()
	if pkg.Path() != pkgErrorsPath {
		return ""
	}
	return f.Name()
}

func isExcludedCtor(pass *analysis.Pass, call *ast.CallExpr, list []string) bool {
	fn := typeutil.Callee(pass.TypesInfo, call)
	f, ok := fn.(*types.Func)
	if !ok {
		return false
	}
	pkgName := ""
	if pkg := f.Pkg(); pkg != nil {
		pkgName = pkg.Name()
	}
	return exclude.Match(list, pkgName, f.Name())
}

func funcReturnsError(pass *analysis.Pass, fn *ast.FuncDecl) bool {
	obj, ok := pass.TypesInfo.Defs[fn.Name].(*types.Func)
	if !ok {
		return false
	}
	sig, ok := obj.Type().(*types.Signature)
	if !ok {
		return false
	}
	results := sig.Results()
	for v := range results.Variables() {
		if isErrorType(v.Type()) {
			return true
		}
	}
	return false
}

// localCallErrors собирает локальные переменные функции, значение
// которых получено из вызова и реализует error (err := f(); a, err := f()).
func localCallErrors(pass *analysis.Pass, fn *ast.FuncDecl) map[types.Object]struct{} {
	out := map[types.Object]struct{}{}
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		// Источник — ровно один вызов в правой части.
		if len(assign.Rhs) != 1 {
			return true
		}
		if _, ok := assign.Rhs[0].(*ast.CallExpr); !ok {
			return true
		}
		for _, lhs := range assign.Lhs {
			id, ok := lhs.(*ast.Ident)
			if !ok || id.Name == "_" {
				continue
			}
			obj := objectOf(pass, id)
			if obj == nil || !isErrorType(obj.Type()) {
				continue
			}
			out[obj] = struct{}{}
		}
		return true
	})
	return out
}

func isLocalCallErr(pass *analysis.Pass, expr ast.Expr, callErrs map[types.Object]struct{}) bool {
	id, ok := expr.(*ast.Ident)
	if !ok {
		return false
	}
	obj := objectOf(pass, id)
	if obj == nil {
		return false
	}
	_, ok = callErrs[obj]
	return ok
}

// isErrorExpr сообщает, что выражение имеет тип error и не является
// статичной ошибкой (package-level var / именованный error-литерал).
func isErrorExpr(pass *analysis.Pass, expr ast.Expr) bool {
	tv, ok := pass.TypesInfo.Types[expr]
	if !ok {
		return false
	}
	return isErrorType(tv.Type) && !isStaticError(pass, expr)
}

// isStaticError: package-level var типа error (ErrSome) либо
// композит-литерал / адрес именованного error-типа (BigError{}, &BigError{}).
func isStaticError(pass *analysis.Pass, expr ast.Expr) bool {
	switch e := expr.(type) {
	case *ast.Ident, *ast.SelectorExpr:
		obj := objectOf(pass, exprIdent(e))
		v, ok := obj.(*types.Var)
		if !ok || !isPackageLevel(v) {
			return false
		}
		return isErrorType(v.Type())
	case *ast.UnaryExpr:
		if e.Op == token.AND {
			if cl, ok := e.X.(*ast.CompositeLit); ok {
				return isNamedErrorType(pass, cl)
			}
		}
	case *ast.CompositeLit:
		return isNamedErrorType(pass, e)
	}
	return false
}

func isNamedErrorType(pass *analysis.Pass, cl *ast.CompositeLit) bool {
	tv, ok := pass.TypesInfo.Types[cl]
	if !ok {
		return false
	}
	t := tv.Type
	if _, ok := t.(*types.Named); !ok {
		return false
	}
	// Тип должен реализовывать error сам по себе или по указателю.
	if isErrorType(t) {
		return true
	}
	return isErrorType(types.NewPointer(t))
}

// exprIdent извлекает целевой идентификатор из Ident или SelectorExpr.
func exprIdent(e ast.Expr) *ast.Ident {
	switch x := e.(type) {
	case *ast.Ident:
		return x
	case *ast.SelectorExpr:
		return x.Sel
	}
	return nil
}

func objectOf(pass *analysis.Pass, id *ast.Ident) types.Object {
	if id == nil {
		return nil
	}
	if obj := pass.TypesInfo.Uses[id]; obj != nil {
		return obj
	}
	return pass.TypesInfo.Defs[id]
}

func isPackageLevel(v *types.Var) bool {
	pkg := v.Pkg()
	pkgScope := pkg.Scope()
	return v.Parent() != nil && v.Parent() == pkgScope
}

func isErrorType(t types.Type) bool {
	if t == nil {
		return false
	}
	errObj := types.Universe.Lookup("error")
	errType := errObj.Type()
	errIface, ok := errType.Underlying().(*types.Interface)
	if !ok {
		return false
	}
	return types.Implements(t, errIface)
}
