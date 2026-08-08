// Package bytesinloop implements rule GID-182 (Uber: avoid repeated
// string-to-byte conversions): converting a string literal or constant
// to []byte/[]rune inside a loop body must be computed once before the loop.
//
// What is matched:
//   - []byte("literal") inside a for/range body (including nested blocks);
//   - []rune("literal") in the same places;
//   - []byte(constStr), where constStr is a string constant (the value is
//     computed via pass.TypesInfo, constant value, types.String);
//   - a conversion inside the body of a closure declared in the loop (the
//     closure runs on every iteration).
//
// What is NOT matched:
//   - []byte(variable) — a conversion of a variable (not a constant): the value
//     may change, it cannot be hoisted;
//   - []byte("literal") outside a loop — it is computed once anyway;
//   - []byte(param), where param is a function/closure parameter.
//
// Generated code (ast.IsGenerated) is skipped. LoadMode — TypesInfo.
package bytesinloop

import (
	"go/ast"
	"go/types"

	"github.com/slipros/gid-data-golang-eval/internal/astwalk"
	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-182"

// Analyzer — rule GID-182: conversion of a string literal/constant to []byte/[]rune inside a loop.
var Analyzer = &analysis.Analyzer{
	Name:     "gidbytesinloop",
	Doc:      ruleID + ": converting a string literal/constant to []byte/[]rune inside a loop. Fix: compute the conversion once before the loop.",
	Requires: astwalk.Requires,
	Run:      run,
}

// loopFilter — the loop statements, the block that is their body, and the
// conversions being judged. The bodies of closures declared in a loop are
// lexically inside that body and so count as "in the loop", exactly as before.
var loopFilter = []ast.Node{
	(*ast.ForStmt)(nil),
	(*ast.RangeStmt)(nil),
	(*ast.BlockStmt)(nil),
	(*ast.CallExpr)(nil),
}

func run(pass *analysis.Pass) (any, error) {
	// depth counts the loop bodies currently open around the traversal. The
	// loop header (init/cond/post, or the ranged expression) is not part of the
	// body, so only the body block itself opens a level.
	var (
		depth      int
		loopBodies = map[*ast.BlockStmt]struct{}{}
	)

	astwalk.Around(pass, loopFilter, ast.IsGenerated, func(_ *ast.File, n ast.Node, push bool) bool {
		switch node := n.(type) {
		case *ast.ForStmt:
			if push {
				loopBodies[node.Body] = struct{}{}
			}
		case *ast.RangeStmt:
			if push {
				loopBodies[node.Body] = struct{}{}
			}
		case *ast.BlockStmt:
			if _, ok := loopBodies[node]; ok {
				if push {
					depth++
				} else {
					depth--
				}
			}
		case *ast.CallExpr:
			if push && depth > 0 {
				checkConversion(pass, node)
			}
		}

		return true
	})

	return nil, nil
}

// checkConversion: if call is a []byte(X)/[]rune(X) conversion where X is
// a string constant, emits a diagnostic.
func checkConversion(pass *analysis.Pass, call *ast.CallExpr) {
	kind, ok := sliceConversionKind(call.Fun)
	if !ok {
		return
	}
	if len(call.Args) != 1 {
		return
	}
	arg := call.Args[0]
	tv, ok := pass.TypesInfo.Types[arg]
	if !ok || tv.Value == nil {
		return // not a constant (a variable, parameter, call) — skip.
	}
	// The value is a constant; make sure its type is a string type.
	basic, ok := tv.Type.Underlying().(*types.Basic)
	if !ok || basic.Info()&types.IsString == 0 {
		return
	}
	pass.Reportf(call.Pos(),
		"%s: converting to %s inside a loop repeats the allocation. "+
			"Fix: compute it once before the loop.", ruleID, kind)
}

// sliceConversionKind: if fun is the type []byte or []rune (an ArrayType
// without a length whose element is byte/rune), returns the string "[]byte"
// or "[]rune".
func sliceConversionKind(fun ast.Expr) (string, bool) {
	arr, ok := fun.(*ast.ArrayType)
	if !ok || arr.Len != nil {
		return "", false // not a slice ([N]T is an array, not a conversion here).
	}
	elt, ok := arr.Elt.(*ast.Ident)
	if !ok {
		return "", false
	}
	switch elt.Name {
	case "byte":
		return "[]byte", true
	case "rune":
		return "[]rune", true
	default:
		return "", false
	}
}
