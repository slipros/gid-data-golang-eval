// Package buildsig реализует правило GID-212 (build-signature): контракт
// build-функций репозитория.
//
// Источник: repo.md.
//
// Проверки:
//
//  1. Сигнатура результата. В пакетах /dal/repository/build/** экспортируемые
//     функции (FuncDecl без получателя) обязаны возвращать ЛИБО
//     (string, []any, error) — одиночный запрос (sql, args, err), ЛИБО
//     (*<...>.Batch, error) — batch-операция (матч по имени именованного типа
//     Batch, пакет любой). Любая другая сигнатура результата → диагностика.
//     Неэкспортируемые функции-хелперы build-пакета не проверяются.
//
//  2. Бан импорта squirrel. Импорт github.com/Masterminds/squirrel разрешён
//     только в пакетах /dal/repository/build/**. В любом другом пакете импорт
//     squirrel → диагностика.
//
// Сигнатуры распознаются структурно через go/types (LoadModeTypesInfo).
// Сгенерированный код пропускается.
package buildsig

import (
	"go/ast"
	"go/types"
	"strconv"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-212"

// Analyzer — правило GID-212: контракт build-функций репозитория.
var Analyzer = &analysis.Analyzer{
	Name: "gidbuildsig",
	Doc: ruleID + ": build functions return (string, []any, error) or (*batch.Batch, error); " +
		"squirrel only in /dal/repository/build",
	Run: run,
}

func run(pass *analysis.Pass) (any, error) {
	inBuild := pathseg.Contains(pass.Pkg.Path(), "dal", "repository", "build")

	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}

		// Проверка 2: импорт squirrel разрешён только в build-пакетах.
		if !inBuild {
			checkSquirrelImports(pass, file)
		}

		// Проверка 1: сигнатура результата экспортируемых build-функций.
		if inBuild {
			checkBuildSignatures(pass, file)
		}
	}
	return nil, nil
}

// checkSquirrelImports флагует импорт squirrel вне build-пакета.
func checkSquirrelImports(pass *analysis.Pass, file *ast.File) {
	const (
		squirrelPkg = "github.com/Masterminds/squirrel"
		msgSquirrel = ruleID + ": squirrel is allowed only in repository build packages (/dal/repository/build). Fix: move squirrel usage into /dal/repository/build"
	)
	for _, imp := range file.Imports {
		path, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			continue
		}
		if path == squirrelPkg {
			pass.Reportf(imp.Pos(), msgSquirrel)
		}
	}
}

// checkBuildSignatures проверяет результат экспортируемых функций без получателя.
func checkBuildSignatures(pass *analysis.Pass, file *ast.File) {
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		// Методы (есть получатель) и неэкспортируемые хелперы не проверяем.
		if fn.Recv != nil || !fn.Name.IsExported() {
			continue
		}
		obj, ok := pass.TypesInfo.Defs[fn.Name].(*types.Func)
		if !ok {
			continue
		}
		sig, ok := obj.Type().(*types.Signature)
		if !ok {
			continue
		}
		if isSingleQuerySig(sig) || isBatchSig(sig) {
			continue
		}
		const msgSignature = ruleID +
			": a build function must return (sql string, args []any, err error) or (*batch.Batch, error). Fix: adjust the signature"
		pass.Reportf(fn.Name.Pos(), msgSignature)
	}
}

// isSingleQuerySig — результат (string, []any, error).
func isSingleQuerySig(sig *types.Signature) bool {
	res := sig.Results()
	if res.Len() != 3 {
		return false
	}
	sqlRes, argsRes, errRes := res.At(0), res.At(1), res.At(2)
	if !isString(sqlRes.Type()) {
		return false
	}
	if !isSliceOfAny(argsRes.Type()) {
		return false
	}
	return isError(errRes.Type())
}

// isBatchSig — результат (*<...>.Batch, error): указатель на именованный тип
// с именем Batch (пакет любой).
func isBatchSig(sig *types.Signature) bool {
	const batchType = "Batch"
	res := sig.Results()
	if res.Len() != 2 {
		return false
	}
	batchRes, errRes := res.At(0), res.At(1)
	if !isPtrToNamed(batchRes.Type(), batchType) {
		return false
	}
	return isError(errRes.Type())
}

func isString(t types.Type) bool {
	b, ok := t.Underlying().(*types.Basic)
	return ok && b.Kind() == types.String
}

// isSliceOfAny — []any (срез с пустым интерфейсом в качестве элемента).
func isSliceOfAny(t types.Type) bool {
	sl, ok := t.Underlying().(*types.Slice)
	if !ok {
		return false
	}
	elem := sl.Elem()
	iface, ok := elem.Underlying().(*types.Interface)
	return ok && iface.NumMethods() == 0
}

func isError(t types.Type) bool {
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	return obj != nil && obj.Pkg() == nil && obj.Name() == "error"
}

// isPtrToNamed — указатель на именованный тип с заданным именем.
func isPtrToNamed(t types.Type, name string) bool {
	ptr, ok := t.(*types.Pointer)
	if !ok {
		return false
	}
	named, ok := ptr.Elem().(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	return obj != nil && obj.Name() == name
}
