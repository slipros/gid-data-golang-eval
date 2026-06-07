// Package sqlnull реализует правило GID-122: nullable-поля entity (DAL)
// описываются типами database/sql — sql.NullString, sql.NullTime,
// sql.NullInt32/64 или обобщённым sql.Null[T] — а не указателями.
package sqlnull

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-122"

// Analyzer — правило GID-122: nullable-поля entity — sql.Null*, не указатели.
var Analyzer = &analysis.Analyzer{
	Name: "gidsqlnull",
	Doc:  ruleID + ": nullable-поля entity — sql.Null*, не указатели",
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
			"%s: nullable-поле entity описывается типом %s, не указателем", ruleID, hint)
	}
}

// nullableHint — подходящий sql-тип для элемента указателя.
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
		// Нестандартный тип (структура и т.п.) — обобщённый sql.Null[T].
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
