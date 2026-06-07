// Package optsstyle реализует правило GID-152: конвенции Options-типов.
//
//   - opts в параметрах функций передаётся указателем (*XxxOptions);
//   - opts в теле сущности встраивается (embedded), а не хранится
//     именованным полем.
package optsstyle

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-152"

// Analyzer — правило GID-152: opts указателем в параметрах, embedded в структуре.
var Analyzer = &analysis.Analyzer{
	Name: "gidoptsstyle",
	Doc:  ruleID + ": opts передаётся указателем и встраивается в структуру",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				checkParams(pass, d)
			case *ast.GenDecl:
				checkStructs(pass, d)
			}
		}
	}
	return nil, nil
}

// checkParams: параметр-Options по значению — нарушение.
func checkParams(pass *analysis.Pass, fn *ast.FuncDecl) {
	if fn.Type.Params == nil {
		return
	}
	for _, field := range fn.Type.Params.List {
		t := pass.TypesInfo.TypeOf(field.Type)
		if name, ok := optionsName(t); ok {
			pass.Reportf(field.Pos(),
				"%s: opts передаётся указателем — используйте *%s", ruleID, name)
		}
	}
}

// checkStructs: именованное поле Options — нарушение, opts встраивается.
func checkStructs(pass *analysis.Pass, gd *ast.GenDecl) {
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
			if len(field.Names) == 0 {
				continue // embedded — норма
			}
			t := pass.TypesInfo.TypeOf(field.Type)
			if ptr, ok := t.(*types.Pointer); ok {
				t = ptr.Elem()
			}
			if name, ok := optionsName(t); ok {
				pass.Reportf(field.Pos(),
					"%s: opts встраивается в тело сущности (embedded %s), а не хранится именованным полем",
					ruleID, name)
			}
		}
	}
}

// optionsName возвращает имя типа, если это именованный Options-тип
// (по значению, без указателя).
func optionsName(t types.Type) (string, bool) {
	named, ok := t.(*types.Named)
	if !ok {
		return "", false
	}
	obj := named.Obj()
	name := obj.Name()
	if !strings.HasSuffix(name, "Options") {
		return "", false
	}
	return name, true
}
