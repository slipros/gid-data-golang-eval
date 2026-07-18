// Package buildsig implements rule GID-212 (build-signature): the contract
// of repository build functions.
//
// Source: repo.md.
//
// Checks:
//
//  1. Result signature. In /dal/repository/build/** packages, exported
//     functions (FuncDecl without a receiver) must return EITHER
//     (string, []any, error) — a single query (sql, args, err), OR
//     (*<...>.Batch, error) — a batch operation (matched by the name of the
//     named type Batch, any package). Any other result signature → diagnostic.
//     Unexported helper functions of the build package are not checked.
//
//  2. Ban on the squirrel import. Importing github.com/Masterminds/squirrel is
//     allowed only in /dal/repository/build/** packages. In any other package
//     a squirrel import → diagnostic.
//
// Signatures are recognized structurally via go/types (LoadModeTypesInfo).
// Generated code is skipped.
package buildsig

import (
	"go/ast"
	"go/types"
	"strconv"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-212"

// Analyzer — rule GID-212: the contract of repository build functions.
var Analyzer = &analysis.Analyzer{
	Name: "gidbuildsig",
	Doc: ruleID + ": build functions return (string, []any, error) or (*batch.Batch, error); " +
		"squirrel only in /dal/repository/build",
	Run: run,
}

func run(pass *analysis.Pass) (any, error) {
	inBuild := pathseg.HasLayer(pass.Pkg.Path(), "dal", "repository", "build")

	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}

		// Check 2: the squirrel import is allowed only in build packages.
		if !inBuild {
			checkSquirrelImports(pass, file)
		}

		// Check 1: the result signature of exported build functions.
		if inBuild {
			checkBuildSignatures(pass, file)
		}
	}
	return nil, nil
}

// checkSquirrelImports flags a squirrel import outside a build package.
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

// checkBuildSignatures checks the result of exported functions without a receiver.
func checkBuildSignatures(pass *analysis.Pass, file *ast.File) {
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		// Methods (with a receiver) and unexported helpers are not checked.
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

// isSingleQuerySig — result (string, []any, error).
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

// isBatchSig — result (*<...>.Batch, error): a pointer to a named type
// with the name Batch (any package).
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

// isSliceOfAny — []any (a slice with the empty interface as the element).
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

// isPtrToNamed — a pointer to a named type with the given name.
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
