// Package optsstyle implements rule GID-152: Options-type conventions.
//
//   - opts in function parameters is passed by pointer (*XxxOptions);
//   - opts in the entity body is stored as an unexported named field (opts Options / opts *Options);
//   - embedding an Options type (anonymous field) is a violation: it promotes option fields into the public API.
package optsstyle

import (
	"go/ast"
	"go/types"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-152"

// Analyzer — rule GID-152: opts by pointer in parameters, stored as an unexported named field in the struct.
var Analyzer = &analysis.Analyzer{
	Name: "gidoptsstyle",
	Doc:  ruleID + ": opts is passed by pointer in parameters and stored as an unexported named field in the struct. Embedding opts is forbidden.",
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

// checkParams: an Options parameter by value is a violation.
func checkParams(pass *analysis.Pass, fn *ast.FuncDecl) {
	if fn.Type.Params == nil {
		return
	}
	for _, field := range fn.Type.Params.List {
		t := pass.TypesInfo.TypeOf(field.Type)
		if name, ok := optionsName(t); ok {
			pass.Reportf(field.Pos(),
				"%s: opts must be passed by pointer. Fix: use *%s", ruleID, name)
		}
	}
}

// checkStructs inspects every struct field that involves an Options type:
//   - embedded (anonymous) Options field → violation: embedding promotes option fields into the public API;
//   - exported named Options field → violation: opts must be unexported;
//   - unexported named Options field → OK (this is the required pattern).
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
			// Resolve the Options type (strip pointer if any).
			t := pass.TypesInfo.TypeOf(field.Type)
			if ptr, ok2 := t.(*types.Pointer); ok2 {
				t = ptr.Elem()
			}
			name, ok2 := optionsName(t)
			if !ok2 {
				continue
			}

			if len(field.Names) == 0 {
				// Anonymous (embedded) field — this is a violation.
				pass.Reportf(field.Pos(),
					"%s: embedding %s is forbidden: it promotes option fields into the public API. Fix: use an unexported named field `opts %s`",
					ruleID, name, name)
				continue
			}

			// Named field: check visibility.
			fieldName := field.Names[0].Name
			if isExported(fieldName) {
				pass.Reportf(field.Pos(),
					"%s: Options field %q must be unexported. Fix: rename to `opts %s`",
					ruleID, fieldName, name)
			}
			// Unexported named field — OK, no diagnostic.
		}
	}
}

// optionsName returns the type name if it is a named Options type (value or pointer).
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

// isExported reports whether a Go identifier is exported.
func isExported(name string) bool {
	if name == "" {
		return false
	}
	r := []rune(name)
	return unicode.IsUpper(r[0])
}
