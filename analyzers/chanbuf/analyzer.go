// Package chanbuf реализует правило GID-179 (Uber: channel size is one or none):
// размер буфера канала в make(chan T, N) допустим только 0 или 1.
// Больший буфер (N > 1) с константным значением запрещён — он почти всегда
// маскирует проблему синхронизации и должен быть обоснован явно.
//
// Что матчится:
//   - make(chan T, 2), make(chan T, 100) — литерал > 1;
//   - make(chan T, maxWorkers), где maxWorkers — именованная const = 10
//     (значение вычисляется через TypesInfo, constant.Int).
//
// Что НЕ матчится:
//   - make(chan T), make(chan T, 0), make(chan T, 1) — буфер 0 или 1;
//   - make(chan T, n), где n — переменная/вызов (размер обоснован рантаймом —
//     решение за review);
//   - make([]T, N), make(map[K]V, N) — не каналы.
//
// Точечное отключение: //nolint:gidchanbuf (работает через golangci-lint,
// в коде анализатора ничего не требуется).
package chanbuf

import (
	"go/ast"
	"go/constant"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-179"

// Analyzer — правило GID-179: размер буфера канала только 0 или 1.
var Analyzer = &analysis.Analyzer{
	Name: "gidchanbuf",
	Doc:  ruleID + ": channel buffer size must be 0 or 1. Fix: use an unbuffered channel or buffer 1, or justify a larger buffer with //nolint:gidchanbuf.",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if !isMakeBuiltin(pass, call) {
				return true
			}
			// make(chan T, N): первый аргумент — тип канала, второй — размер.
			if len(call.Args) < 2 {
				return true
			}
			if _, ok := call.Args[0].(*ast.ChanType); !ok {
				return true // make([]T, N) / make(map[K]V, N) — не канал.
			}
			sizeExpr := call.Args[1]
			tv, ok := pass.TypesInfo.Types[sizeExpr]
			if !ok || tv.Value == nil {
				return true // размер не константа (переменная/вызов) — пропускаем.
			}
			size, ok := constant.Int64Val(constant.ToInt(tv.Value))
			if !ok {
				return true
			}
			if size <= 1 {
				return true // 0 и 1 — допустимы.
			}
			pass.Reportf(sizeExpr.Pos(),
				"%s: channel buffer %d is not allowed (only 0 or 1). "+
					"Fix: use an unbuffered channel or buffer 1, or justify a larger buffer with //nolint:gidchanbuf.",
				ruleID, size)
			return true
		})
	}
	return nil, nil
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
