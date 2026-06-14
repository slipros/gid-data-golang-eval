// Package noptr implements the rules about nullable pointers:
//
//   - GID-120: *uuid.UUID is forbidden everywhere — emptiness is checked with IsNil();
//   - GID-121: in /domain/model struct fields do not use *time.Time or
//     pointers to string types — the zero value expresses absence
//     (IsZero(), len == 0). A pointer is justified only when the zero value
//     is meaningful (e.g. *bool).
package noptr

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const (
	ruleUUID = "GID-120"
	ruleZero = "GID-121"
)

// Analyzer — the GID rule: see Doc.
var Analyzer = &analysis.Analyzer{
	Name: "gidnoptr",
	Doc:  ruleUUID + "/" + ruleZero + ": forbid pointers where the type checks emptiness itself (uuid, time, string). Fix: use the value type",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	inModel := pathseg.Contains(pass.Pkg.Path(), "domain", "model")
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		checkUUIDPointers(pass, file)
		if inModel {
			checkModelFields(pass, file)
		}
	}
	return nil, nil
}

// checkUUIDPointers — GID-120: *uuid.UUID in any type position.
func checkUUIDPointers(pass *analysis.Pass, file *ast.File) {
	ast.Inspect(file, func(n ast.Node) bool {
		star, ok := n.(*ast.StarExpr)
		if !ok {
			return true
		}
		tv, ok := pass.TypesInfo.Types[star]
		if !ok || !tv.IsType() {
			return true // a dereference, not a type
		}
		ptr, ok := tv.Type.(*types.Pointer)
		if !ok {
			return true
		}
		if isUUID(ptr.Elem()) {
			pass.Reportf(star.Pos(),
				"%s: *uuid.UUID is forbidden. Fix: use uuid.UUID and check emptiness with IsNil()", ruleUUID)
		}
		return true
	})
}

// checkModelFields — GID-121: pointers to time.Time/string types in model fields.
func checkModelFields(pass *analysis.Pass, file *ast.File) {
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
				checkModelField(pass, field)
			}
		}
	}
}

func checkModelField(pass *analysis.Pass, field *ast.Field) {
	ptr, ok := pass.TypesInfo.TypeOf(field.Type).(*types.Pointer)
	if !ok {
		return
	}
	elem := ptr.Elem()
	switch {
	case isTime(elem):
		pass.Reportf(field.Pos(),
			"%s: *time.Time is unnecessary in model. Fix: use time.Time and check absence with t.IsZero()", ruleZero)
	case isStringBased(elem):
		pass.Reportf(field.Pos(),
			"%s: a pointer to a string type is unnecessary in model. Fix: use the value and check len(s) == 0", ruleZero)
	}
}

func isUUID(t types.Type) bool {
	const uuidPkg = "github.com/gofrs/uuid"
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	if obj.Pkg() == nil || obj.Name() != "UUID" {
		return false
	}
	pkg := obj.Pkg()
	path := pkg.Path()
	return path == uuidPkg || pathseg.Contains(path, "gofrs", "uuid")
}

func isTime(t types.Type) bool {
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	pkg := obj.Pkg()
	return pkg != nil && pkg.Path() == "time" && obj.Name() == "Time"
}

func isStringBased(t types.Type) bool {
	basic, ok := t.Underlying().(*types.Basic)
	return ok && basic.Kind() == types.String
}
