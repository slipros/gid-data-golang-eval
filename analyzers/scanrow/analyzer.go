// Package scanrow implements GID-245 (gidscanrow): an epgx
// conn.Select(ctx, &out, ...) where out is a struct with exactly one field
// reads a single column — conn.ScanRow(ctx, []any{&field}, ...) with the column
// pointer directly is idiomatic instead of mapping into a one-field struct.
//
//	// bad — Select maps one column into a one-field struct
//	var out struct {
//	    MemberID uuid.NullUUID `db:"member_id"`
//	}
//	if err := t.conn.Select(ctx, &out, sql, args); err != nil { ... }
//
//	// good — ScanRow reads the single column into the value directly
//	var out uuid.NullUUID
//	if err := t.conn.ScanRow(ctx, []any{&out}, sql, args); err != nil { ... }
//
// The receiver of Select is confirmed to be an epgx connection by exposing a
// sibling method ScanRow(context.Context, []any, string, ...any) error — a
// specific fingerprint that avoids false positives on any unrelated .Select
// call. The flagged target is only a struct pointee with exactly one field;
// slices (*[]T, Select's multi-row form) and multi-field structs are out of
// scope. Generated code (ast.IsGenerated) is skipped.
package scanrow

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/exclude"
)

const ruleID = "GID-245"

// Analyzer — GID-245 with default settings (no exclusions).
var Analyzer = NewAnalyzer(Settings{})

// Settings — linter settings from .golangci.yml.
type Settings struct {
	// Exclude — methods exempted from the rule: "Function" / "Method" or
	// "Type.Method".
	Exclude []string `json:"exclude"`
}

// NewAnalyzer builds the GID-245 analyzer.
func NewAnalyzer(s Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "gidscanrow",
		Doc: ruleID + ": an epgx Select into a one-field struct reads a single column — " +
			"use ScanRow with the column pointer directly",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, s)
		},
	}
}

func run(pass *analysis.Pass, s Settings) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			if exclude.Match(s.Exclude, recvTypeName(fn), fn.Name.Name) {
				continue
			}
			checkFunc(pass, fn)
		}
	}
	return nil, nil
}

func checkFunc(pass *analysis.Pass, fn *ast.FuncDecl) {
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		const diagMessage = ruleID + ": Select into a single-field struct reads one column — use ScanRow with the field pointer. " +
			"Fix: var out T; conn.ScanRow(ctx, []any{&out}, sql, args...)"
		if isSelectIntoSingleFieldStruct(pass, call) {
			pass.Reportf(call.Pos(), diagMessage)
		}
		return true
	})
}

// isSelectIntoSingleFieldStruct reports whether call is an epgx
// conn.Select(ctx, &out, ...) whose out is a struct with exactly one field.
func isSelectIntoSingleFieldStruct(pass *analysis.Pass, call *ast.CallExpr) bool {
	const selectMethod = "Select"
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != selectMethod {
		return false
	}
	// Select(ctx, ptr, sql, args...) — need at least the ctx and ptr arguments.
	if len(call.Args) < 2 {
		return false
	}
	recvType := pass.TypesInfo.TypeOf(sel.X)
	if recvType == nil || !hasScanRowMethod(pass, recvType) {
		return false
	}
	return isSingleFieldStructPtr(pass.TypesInfo.TypeOf(call.Args[1]))
}

// hasScanRowMethod reports whether t exposes a method
// ScanRow(context.Context, []any, string, ...any) error — the epgx fingerprint.
func hasScanRowMethod(pass *analysis.Pass, t types.Type) bool {
	const scanRowMethod = "ScanRow"
	obj, _, _ := types.LookupFieldOrMethod(t, true, pass.Pkg, scanRowMethod)
	fn, ok := obj.(*types.Func)
	if !ok {
		return false
	}
	sig, ok := fn.Type().(*types.Signature)
	if !ok || !sig.Variadic() {
		return false
	}
	params := sig.Params()
	if params.Len() < 2 {
		return false
	}
	// The second parameter is the scan slice ([]any) — the distinctive shape.
	second := params.At(1)
	paramType := second.Type()
	_, isSlice := paramType.Underlying().(*types.Slice)
	return isSlice
}

// isSingleFieldStructPtr reports whether t is a pointer to a struct with
// exactly one field (Select's single-row, single-column target).
func isSingleFieldStructPtr(t types.Type) bool {
	if t == nil {
		return false
	}
	ptr, ok := t.Underlying().(*types.Pointer)
	if !ok {
		return false
	}
	elem := ptr.Elem()
	st, ok := elem.Underlying().(*types.Struct)
	return ok && st.NumFields() == 1
}

// recvTypeName returns the name of fn's receiver type ("" for a free function).
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
