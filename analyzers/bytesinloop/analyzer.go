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
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-182"

// Analyzer — rule GID-182: conversion of a string literal/constant to []byte/[]rune inside a loop.
var Analyzer = &analysis.Analyzer{
	Name: "gidbytesinloop",
	Doc:  ruleID + ": converting a string literal/constant to []byte/[]rune inside a loop. Fix: compute the conversion once before the loop.",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}

		// Collect the positional ranges of all loop bodies (for/range).
		// Nested blocks and the bodies of closures declared in the loop are
		// lexically inside this range — and therefore count as "in the loop".
		var loopBodies []*ast.BlockStmt
		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.ForStmt:
				loopBodies = append(loopBodies, node.Body)
			case *ast.RangeStmt:
				loopBodies = append(loopBodies, node.Body)
			}
			return true
		})

		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if !insideAnyLoop(call.Pos(), loopBodies) {
				return true
			}
			checkConversion(pass, call)
			return true
		})
	}
	return nil, nil
}

// insideAnyLoop reports whether the position pos lies inside the body of at least one loop.
func insideAnyLoop(pos token.Pos, bodies []*ast.BlockStmt) bool {
	for _, b := range bodies {
		// Lbrace < pos < Rbrace — the position is strictly inside the body's braces.
		if pos > b.Lbrace && pos < b.Rbrace {
			return true
		}
	}
	return false
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
