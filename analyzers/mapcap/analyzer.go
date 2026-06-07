// Package mapcap реализует правило GID-183 (Uber perf: map capacity hints):
// если map создаётся через make(map[K]V) БЕЗ аргумента-ёмкости, а затем в той же
// функции заполняется в range-цикле по коллекции с известной длиной (слайс, мапа,
// строка), нужно указать хинт ёмкости: make(map[K]V, len(src)). Стандартный
// prealloc такой случай не покрывает — он умеет только слайсы.
//
// Паттерн в пределах одной функции:
//  1. m := make(map[K]V)        // или var m = make(map[K]V) — без capacity;
//  2. for ... := range src {    // src — слайс/мапа/строка
//     m[...] = ...          // безусловное присваивание по индексу в m
//     }
//
// Эвристика консервативная — матчим только заведомо безопасные случаи:
//   - между make и циклом m НЕ должна использоваться никак (ни заполнение вне
//     цикла, ни передача в вызов): любое упоминание m отменяет диагностику,
//     т.к. её длина к моменту цикла уже неизвестна анализатору;
//   - range по каналу НЕ матчим — у канала нет len, заранее размер неизвестен;
//   - присваивание m[...] = ... внутри if (условное заполнение) в теле цикла
//     НЕ матчим — реальное число вставок меньше len(src), хинт может навредить.
//
// make с уже указанной ёмкостью (make(map[K]V, n)) — корректен, не матчим.
// Сгенерированный код (ast.IsGenerated) пропускается.
// LoadMode — TypesInfo (нужны типы, чтобы отличить слайс/мапу/строку от канала).
package mapcap

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-183"

// Analyzer — правило GID-183: make(map) без capacity при заполнении из range — укажите хинт len(src).
var Analyzer = &analysis.Analyzer{
	Name: "gidmapcap",
	Doc:  ruleID + ": make(map) без capacity при заполнении из range — укажите хинт len(src)",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				return true
			}
			checkBlock(pass, fn.Body.List)
			return true
		})
	}
	return nil, nil
}

// checkBlock анализирует последовательность операторов одного блока: находит
// make(map) без ёмкости и ищет ниже в том же блоке range-цикл, заполняющий map.
// Рекурсивно спускается во вложенные блоки (тела if/for/...), чтобы паттерн
// внутри вложенного блока тоже ловился.
func checkBlock(pass *analysis.Pass, stmts []ast.Stmt) {
	for i, stmt := range stmts {
		// Спуск во вложенные блоки — паттерн локален для своего блока операторов.
		inspectNestedBlocks(pass, stmt)

		name, makeCall := mapMakeWithoutCap(pass, stmt)
		if name == "" {
			continue
		}
		obj := pass.TypesInfo.ObjectOf(declIdent(stmt, name))
		if obj == nil {
			continue
		}
		analyzeAfterMake(pass, stmts[i+1:], obj, makeCall)
	}
}

// analyzeAfterMake просматривает операторы после make. Возвращает (через Report)
// диагностику, если ближайшее использование m — это range-цикл с безусловным
// заполнением m по слайсу/мапе/строке. Любое иное использование m до такого
// цикла отменяет диагностику.
func analyzeAfterMake(pass *analysis.Pass, rest []ast.Stmt, obj types.Object, makeCall *ast.CallExpr) {
	for _, stmt := range rest {
		rng, isRange := stmt.(*ast.RangeStmt)
		if isRange && fillsMapInRange(pass, rng, obj) {
			if !rangeOverKnownLen(pass, rng.X) {
				return // range по каналу/неизвестному — размер неизвестен.
			}
			if usesObjOutsideAssign(pass, rng.Body, obj) {
				return // условное заполнение или иное использование m в теле — не матчим.
			}
			pass.Reportf(makeCall.Pos(),
				"%s: make без capacity при заполнении из range — укажите хинт: make(map[K]V, len(src))",
				ruleID)
			return
		}
		// Любое использование m до подходящего цикла отменяет диагностику.
		if usesObj(pass, stmt, obj) {
			return
		}
	}
}

// mapMakeWithoutCap определяет, что stmt — это объявление map через make без
// аргумента-ёмкости. Возвращает имя переменной и сам вызов make.
// Поддерживает  m := make(map[K]V)  и  var m = make(map[K]V).
func mapMakeWithoutCap(pass *analysis.Pass, stmt ast.Stmt) (string, *ast.CallExpr) {
	var lhs ast.Expr
	var rhs ast.Expr

	switch s := stmt.(type) {
	case *ast.AssignStmt:
		if len(s.Lhs) != 1 || len(s.Rhs) != 1 {
			return "", nil
		}
		lhs, rhs = s.Lhs[0], s.Rhs[0]
	case *ast.DeclStmt:
		gen, ok := s.Decl.(*ast.GenDecl)
		if !ok || gen.Tok.String() != "var" || len(gen.Specs) != 1 {
			return "", nil
		}
		vs, ok := gen.Specs[0].(*ast.ValueSpec)
		if !ok || len(vs.Names) != 1 || len(vs.Values) != 1 {
			return "", nil
		}
		lhs, rhs = vs.Names[0], vs.Values[0]
	default:
		return "", nil
	}

	ident, ok := lhs.(*ast.Ident)
	if !ok {
		return "", nil
	}
	call, ok := rhs.(*ast.CallExpr)
	if !ok || !isMakeBuiltin(pass, call) {
		return "", nil
	}
	if len(call.Args) == 0 {
		return "", nil
	}
	if _, ok := call.Args[0].(*ast.MapType); !ok {
		return "", nil // make([]T, ...) / make(chan T) — не мапа.
	}
	if len(call.Args) >= 2 {
		return "", nil // ёмкость уже указана — корректно.
	}
	return ident.Name, call
}

// declIdent возвращает *ast.Ident объявляемой переменной из stmt по имени.
func declIdent(stmt ast.Stmt, name string) *ast.Ident {
	switch s := stmt.(type) {
	case *ast.AssignStmt:
		if id, ok := s.Lhs[0].(*ast.Ident); ok && id.Name == name {
			return id
		}
	case *ast.DeclStmt:
		if gen, ok := s.Decl.(*ast.GenDecl); ok {
			if vs, ok := gen.Specs[0].(*ast.ValueSpec); ok {
				return vs.Names[0]
			}
		}
	}
	return nil
}

// fillsMapInRange сообщает, что в теле range-цикла есть присваивание m[...] = ...,
// где m — наш объект (на верхнем уровне тела цикла, безусловно).
func fillsMapInRange(pass *analysis.Pass, rng *ast.RangeStmt, obj types.Object) bool {
	for _, stmt := range rng.Body.List {
		if isIndexAssignTo(pass, stmt, obj) {
			return true
		}
	}
	return false
}

// isIndexAssignTo: stmt — это присваивание вида m[key] = value, где m — obj.
func isIndexAssignTo(pass *analysis.Pass, stmt ast.Stmt, obj types.Object) bool {
	assign, ok := stmt.(*ast.AssignStmt)
	if !ok {
		return false
	}
	for _, lhs := range assign.Lhs {
		idx, ok := lhs.(*ast.IndexExpr)
		if !ok {
			continue
		}
		if id, ok := idx.X.(*ast.Ident); ok && pass.TypesInfo.ObjectOf(id) == obj {
			return true
		}
	}
	return false
}

// rangeOverKnownLen: источник range имеет известную длину (слайс, массив, мапа,
// строка). Канал не имеет len — false.
func rangeOverKnownLen(pass *analysis.Pass, x ast.Expr) bool {
	t := pass.TypesInfo.TypeOf(x)
	if t == nil {
		return false
	}
	switch u := t.Underlying().(type) {
	case *types.Slice, *types.Array, *types.Map:
		return true
	case *types.Basic:
		return u.Info()&types.IsString != 0
	case *types.Pointer:
		// *[N]T — указатель на массив, range допустим и длина известна.
		elem := u.Elem()
		_, isArr := elem.Underlying().(*types.Array)
		return isArr
	default:
		return false
	}
}

// usesObjOutsideAssign: в теле цикла obj используется где-либо, кроме безусловных
// присваиваний m[...] = ... на верхнем уровне тела. Любое такое использование
// (условное заполнение внутри if, чтение m, передача m в вызов) делает реальный
// размер неизвестным — диагностику отменяем.
func usesObjOutsideAssign(pass *analysis.Pass, body *ast.BlockStmt, obj types.Object) bool {
	for _, stmt := range body.List {
		if isIndexAssignTo(pass, stmt, obj) {
			continue // безусловное заполнение на верхнем уровне — ожидаемо.
		}
		if usesObj(pass, stmt, obj) {
			return true
		}
	}
	return false
}

// usesObj: в произвольном узле встречается ссылка на obj.
func usesObj(pass *analysis.Pass, node ast.Node, obj types.Object) bool {
	found := false
	ast.Inspect(node, func(n ast.Node) bool {
		if found {
			return false
		}
		if id, ok := n.(*ast.Ident); ok && pass.TypesInfo.ObjectOf(id) == obj {
			found = true
			return false
		}
		return true
	})
	return found
}

// inspectNestedBlocks рекурсивно запускает checkBlock на телах вложенных
// составных операторов, чтобы паттерн внутри них тоже анализировался.
func inspectNestedBlocks(pass *analysis.Pass, stmt ast.Stmt) {
	switch s := stmt.(type) {
	case *ast.BlockStmt:
		checkBlock(pass, s.List)
	case *ast.IfStmt:
		if s.Body != nil {
			checkBlock(pass, s.Body.List)
		}
		if s.Else != nil {
			inspectNestedBlocks(pass, s.Else)
		}
	case *ast.ForStmt:
		if s.Body != nil {
			checkBlock(pass, s.Body.List)
		}
	case *ast.RangeStmt:
		if s.Body != nil {
			checkBlock(pass, s.Body.List)
		}
	case *ast.SwitchStmt:
		if s.Body != nil {
			checkBlock(pass, s.Body.List)
		}
	case *ast.TypeSwitchStmt:
		if s.Body != nil {
			checkBlock(pass, s.Body.List)
		}
	case *ast.CaseClause:
		checkBlock(pass, s.Body)
	case *ast.SelectStmt:
		if s.Body != nil {
			checkBlock(pass, s.Body.List)
		}
	case *ast.CommClause:
		checkBlock(pass, s.Body)
	case *ast.LabeledStmt:
		inspectNestedBlocks(pass, s.Stmt)
	}
}

// isMakeBuiltin: вызов call — это встроенный make, а не локальная функция make.
func isMakeBuiltin(pass *analysis.Pass, call *ast.CallExpr) bool {
	ident, ok := call.Fun.(*ast.Ident)
	if !ok || ident.Name != "make" {
		return false
	}
	builtin, ok := pass.TypesInfo.Uses[ident].(*types.Builtin)
	return ok && builtin.Name() == "make"
}
