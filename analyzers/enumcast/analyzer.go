// Package enumcast implements rule GID-233 (linter gidenumcast): a direct
// cast between two string-based enums declared in different packages is
// forbidden — an enum crosses a layer boundary only via a map conversion
// with comma-ok and gderror.NewUnhandledValueError (GID-143). A direct cast
// silently passes unknown values across the boundary.
//
// Detection (type-based): a conversion expression DstEnum(x) where
//
//   - DstEnum is a string-based enum — a named type with underlying string
//     that has at least one typed const declared in its package (GID-123);
//   - the static type of x is another string-based enum declared in a
//     different package (a cross-package enum→enum cast is the layer
//     boundary smell, so the rule applies everywhere in the codebase).
//
// Allowed (boundary cases): same-package casts, casts from/to plain string,
// casts of untyped constants/literals (in DstEnum("active") the operand is
// typed as DstEnum itself, so it never matches). Named string types without
// consts are not enums and are not matched. The GID-143 map converter
// (map literal + indexing) involves no conversion expression and stays
// clean. _test.go files and generated code (ast.IsGenerated) are skipped.
// Per-case suppression: //nolint:gidenumcast. LoadMode — TypesInfo.
package enumcast

import (
	"go/ast"
	"go/types"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-233"

// Analyzer — rule GID-233: no direct cast between enum types from different
// packages; convert via map with gderror.NewUnhandledValueError (GID-143).
var Analyzer = &analysis.Analyzer{
	Name: "gidenumcast",
	Doc:  ruleID + ": direct cast between enum types crosses a layer boundary unchecked. Fix: convert via map with comma-ok + gderror.NewUnhandledValueError (see GID-143)",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	// enumCache memoizes the "string-based enum with consts" check per named type.
	enumCache := map[*types.Named]bool{}
	for _, file := range pass.Files {
		if ast.IsGenerated(file) || isTestFile(pass, file) {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || len(call.Args) != 1 {
				return true
			}
			tv, ok := pass.TypesInfo.Types[call.Fun]
			if !ok || !tv.IsType() {
				return true // a regular call, not a conversion
			}
			dst := stringEnum(tv.Type, enumCache)
			if dst == nil {
				return true
			}
			src := stringEnum(pass.TypesInfo.TypeOf(call.Args[0]), enumCache)
			if src == nil {
				return true
			}
			srcObj, dstObj := src.Obj(), dst.Obj()
			srcPkg, dstPkg := srcObj.Pkg(), dstObj.Pkg()
			if srcPkg == nil || dstPkg == nil || srcPkg == dstPkg {
				return true // same-package cast is allowed (boundary case)
			}
			pass.Reportf(call.Pos(),
				"%s: direct cast between enum types crosses a layer boundary unchecked. "+
					"Fix: convert via map with comma-ok + gderror.NewUnhandledValueError (see GID-143)",
				ruleID)
			return true
		})
	}
	return nil, nil
}

// stringEnum returns t as *types.Named when t is a string-based enum:
// a named type with underlying string that has at least one typed const
// declared in its package (GID-123). Otherwise returns nil.
func stringEnum(t types.Type, cache map[*types.Named]bool) *types.Named {
	if t == nil {
		return nil
	}
	named, ok := types.Unalias(t).(*types.Named)
	if !ok {
		return nil
	}
	basic, ok := named.Underlying().(*types.Basic)
	if !ok || basic.Kind() != types.String {
		return nil
	}
	isEnum, ok := cache[named]
	if !ok {
		isEnum = hasTypedConst(named)
		cache[named] = isEnum
	}
	if !isEnum {
		return nil
	}
	return named
}

// hasTypedConst reports whether the package of named declares at least one
// package-level const of exactly this type.
func hasTypedConst(named *types.Named) bool {
	obj := named.Obj()
	pkg := obj.Pkg()
	if pkg == nil {
		return false // universe type, cannot be an enum
	}
	scope := pkg.Scope()
	for _, name := range scope.Names() {
		c, ok := scope.Lookup(name).(*types.Const)
		if !ok {
			continue
		}
		if types.Identical(c.Type(), named) {
			return true
		}
	}
	return false
}

// isTestFile reports whether file is a _test.go file.
func isTestFile(pass *analysis.Pass, file *ast.File) bool {
	return strings.HasSuffix(filepath.Base(pass.Fset.Position(file.Pos()).Filename), "_test.go")
}
