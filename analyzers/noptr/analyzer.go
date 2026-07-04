// Package noptr implements the rules about nullable pointers:
//
//   - GID-120: *uuid.UUID is forbidden everywhere — emptiness is checked with IsNil();
//   - GID-121: in /domain/model and /event/dto struct fields do not use
//     pointers to simple types — *time.Time or a pointer to any basic
//     numeric or string type (int*, uint*, float*, complex*, string,
//     including named types based on them) — the zero value expresses
//     absence (IsZero(), len == 0, == 0). *bool is exempt (false is itself
//     a meaningful value), and so is a pointer to a nested struct. When a
//     pointer cannot be avoided, escape with //nolint:gidnoptr.
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
	Doc:  ruleUUID + "/" + ruleZero + ": forbid *uuid.UUID everywhere, and pointers to simple types (time, numeric, string) in domain/model and event/dto — the zero value checks emptiness itself. Fix: use the value type; escape with //nolint:gidnoptr when unavoidable",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	pkgPath := pass.Pkg.Path()
	inScope := pathseg.Contains(pkgPath, "domain", "model") || pathseg.Contains(pkgPath, "event", "dto")
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		checkUUIDPointers(pass, file)
		if inScope {
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

// checkModelFields — GID-121: pointers to simple types (time, numeric, string)
// in /domain/model and /event/dto struct fields.
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
			"%s: *time.Time is unnecessary here. Fix: use time.Time and check absence with t.IsZero(); if a pointer is unavoidable, use //nolint:gidnoptr", ruleZero)
	case isSimpleValueType(elem):
		pass.Reportf(field.Pos(),
			"%s: a pointer to a simple type is unnecessary here. Fix: use the value and check the zero value (len(s) == 0 for strings, == 0 for numbers); if a pointer is unavoidable, use //nolint:gidnoptr", ruleZero)
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

// isSimpleValueType reports whether t's underlying type is a basic numeric or
// string type (including a named type based on one), excluding bool — a
// pointer to bool is exempt because false is itself a meaningful value.
func isSimpleValueType(t types.Type) bool {
	basic, ok := t.Underlying().(*types.Basic)
	if !ok {
		return false
	}
	if basic.Info()&types.IsBoolean != 0 {
		return false
	}
	return basic.Info()&(types.IsInteger|types.IsFloat|types.IsComplex|types.IsString) != 0
}
