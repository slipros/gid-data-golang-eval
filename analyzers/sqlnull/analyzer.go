// Package sqlnull implements rule GID-122: nullable entity (DAL) fields are
// described with database/sql types — sql.NullString, sql.NullTime,
// sql.NullInt32/64, or the generic sql.Null[T] — not with pointers.
package sqlnull

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-122"

// Analyzer — rule GID-122: nullable entity fields use sql.Null*, not pointers. Fix: use sql.Null*.
var Analyzer = &analysis.Analyzer{
	Name: "gidsqlnull",
	Doc:  ruleID + ": nullable entity fields use sql.Null*, not pointers. Fix: use sql.Null*",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	if !pathseg.EndsWith(pass.Pkg.Path(), "dal", "entity") {
		return nil, nil
	}
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				st, ok := ts.Type.(*ast.StructType)
				if !ok {
					continue
				}
				for _, field := range st.Fields.List {
					checkField(pass, field)
				}
			}
		}
	}
	return nil, nil
}

func checkField(pass *analysis.Pass, field *ast.Field) {
	ptr, ok := pass.TypesInfo.TypeOf(field.Type).(*types.Pointer)
	if !ok {
		return
	}
	if hint, ok := nullableHint(ptr.Elem()); ok {
		pass.Reportf(field.Pos(),
			"%s: a nullable entity field must use %s, not a pointer. Fix: replace the pointer with it", ruleID, hint)
	}
}

// nullableHint — the suitable sql type for the pointer's element.
func nullableHint(t types.Type) (string, bool) {
	if named, ok := t.(*types.Named); ok {
		obj := named.Obj()
		pkg := obj.Pkg()
		if pkg != nil && pkg.Path() == "time" && obj.Name() == "Time" {
			return "sql.NullTime", true
		}
	}
	basic, ok := t.Underlying().(*types.Basic)
	if !ok {
		// A non-standard type (a struct, etc.) — the generic sql.Null[T].
		if _, isStruct := t.Underlying().(*types.Struct); isStruct {
			return "sql.Null[T]", true
		}
		return "", false
	}
	switch basic.Kind() {
	case types.String:
		return "sql.NullString", true
	case types.Int32:
		return "sql.NullInt32", true
	case types.Int, types.Int64:
		return "sql.NullInt64", true
	case types.Float32, types.Float64:
		return "sql.NullFloat64", true
	case types.Bool:
		return "sql.NullBool", true
	default:
		return "sql.Null[T]", true
	}
}
