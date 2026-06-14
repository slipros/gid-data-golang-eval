// Package optsnaming implements rule GID-126: naming and defaults conventions
// of the Options pattern.
//
//   - a struct type named EXACTLY Options outside the app layer is a violation:
//     a settings type is named with an entity prefix (JobOptions), not bare
//     Options. In the app layer a bare Options is the norm (composition of
//     GRPCOptions/KafkaOptions).
//   - a package-level var of type <X>Options (including a pointer) whose name
//     does not start with Default is a violation: defaults live in a
//     Default<X>Options variable. Vars in the app layer are checked too
//     (defaults are Default* there as well).
package optsnaming

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-126"

// Analyzer — rule GID-126: an Options type name has an entity prefix,
// defaults are a Default<X>Options variable.
var Analyzer = &analysis.Analyzer{
	Name: "gidoptsnaming",
	Doc:  ruleID + ": an options type has an entity prefix; defaults are a Default<X>Options variable. Fix: rename accordingly",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	inApp := pathseg.Contains(pass.Pkg.Path(), "app")
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			switch gd.Tok {
			case token.TYPE:
				if !inApp {
					checkTypeNames(pass, gd)
				}
			case token.VAR:
				checkDefaultNames(pass, gd)
			}
		}
	}
	return nil, nil
}

// checkTypeNames: a struct type named exactly Options outside the app layer
// is a violation. Non-struct types (an alias to Options, an interface) are not affected.
func checkTypeNames(pass *analysis.Pass, gd *ast.GenDecl) {
	for _, spec := range gd.Specs {
		ts, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		if ts.Name.Name != "Options" {
			continue
		}
		if ts.Assign.IsValid() {
			continue // an alias (type Options = X) — not affected
		}
		if _, ok := ts.Type.(*ast.StructType); !ok {
			continue // struct types only
		}
		pass.Reportf(ts.Name.Pos(),
			"%s: an options type must have an entity prefix. Fix: use JobOptions, not bare Options", ruleID)
	}
}

// checkDefaultNames: a package-level var of type <X>Options (including a
// pointer) whose name does not start with Default is a violation. Local
// variables do not get here (only top-level GenDecls with Tok==var are checked).
func checkDefaultNames(pass *analysis.Pass, gd *ast.GenDecl) {
	for _, spec := range gd.Specs {
		vs, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		for _, name := range vs.Names {
			if name.Name == "_" {
				continue
			}
			obj := pass.TypesInfo.Defs[name]
			if obj == nil {
				continue
			}
			if !isOptionsType(obj.Type()) {
				continue
			}
			if strings.HasPrefix(name.Name, "Default") {
				continue
			}
			pass.Reportf(name.Pos(),
				"%s: option defaults must be a Default<X>Options variable. Fix: rename it", ruleID)
		}
	}
}

// isOptionsType reports whether the type is a named <X>Options (with an
// entity prefix) — by value or by pointer. A bare Options without a prefix
// does not count (it is the settings type itself, not its default).
func isOptionsType(t types.Type) bool {
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	name := obj.Name()
	return strings.HasSuffix(name, "Options") && name != "Options"
}
