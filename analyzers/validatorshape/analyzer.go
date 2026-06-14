// Package validatorshape implements rule GID-213 (validator-shape):
// the shape of a validator. A validator is a struct with a
// Validate(ctx context.Context, req *T) error method whose name matches the
// operation name.
//
// Scope: packages of the validate layer (the "validate" segment in the
// import path). Every EXPORTED struct type (except names with the Options
// suffix) must have a Validate method (pointer or value receiver) where:
//   - the first parameter has type context.Context;
//   - the only result has type error.
//
// Checking the first parameter (ctx) and the single result (error) is
// enough: the req request type may be anything, and the number of parameters
// after ctx is not limited (see validatorshape.feature, the boundary class).
//
// Generated code (ast.IsGenerated) is skipped. LoadMode — TypesInfo:
// go/types is needed to recognize context.Context, error, and the type's methods.
//
// Exceptions:
//   - targeted: //nolint:gidvalidatorshape
//   - centralized: settings.exclude in .golangci.yml — type names that are
//     not considered validators (e.g. "HealthCheck").
//
// Source: validator.md.
package validatorshape

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/exclude"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-213"

// Analyzer — the variant with default settings (no exclusions).
var Analyzer = NewAnalyzer(Settings{})

// Settings — the linter settings from .golangci.yml.
type Settings struct {
	// Exclude — names of struct types that are not considered validators
	// and are not required to have a Validate method.
	Exclude []string `json:"exclude"`
}

// NewAnalyzer builds the GID-213 analyzer from the linter settings (.golangci.yml).
func NewAnalyzer(cfg Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "gidvalidatorshape",
		Doc:  ruleID + ": a validator is a struct with a Validate(ctx context.Context, req *T) error method. Fix: add that method",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, cfg)
		},
	}
}

func run(pass *analysis.Pass, cfg Settings) (any, error) {
	// Scope: validate-layer packages only.
	if !pathseg.Contains(pass.Pkg.Path(), "validate") {
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
				// Exported struct types only.
				if !ts.Name.IsExported() {
					continue
				}
				if _, ok := ts.Type.(*ast.StructType); !ok {
					continue
				}
				// Settings types (*Options) are not validators.
				if strings.HasSuffix(ts.Name.Name, "Options") {
					continue
				}
				// Centralized exclusions by type name.
				if exclude.Match(cfg.Exclude, ts.Name.Name, ts.Name.Name) {
					continue
				}
				checkValidator(pass, ts)
			}
		}
	}
	return nil, nil
}

func checkValidator(pass *analysis.Pass, ts *ast.TypeSpec) {
	obj, ok := pass.TypesInfo.Defs[ts.Name].(*types.TypeName)
	if !ok {
		return
	}
	named, ok := obj.Type().(*types.Named)
	if !ok {
		return
	}
	if hasValidate(named) {
		return
	}
	pass.Reportf(ts.Name.Pos(),
		"%s: validator %q must have a Validate(ctx context.Context, req *T) error method. Fix: add it",
		ruleID, ts.Name.Name)
}

// hasValidate reports whether the type (or a pointer to it) has a Validate
// method with context.Context as the first parameter and error as the only
// result.
func hasValidate(named *types.Named) bool {
	// Look up the method on both T and *T: a pointer receiver is not in T's method set.
	mset := types.NewMethodSet(types.NewPointer(named))
	for i := 0; i < mset.Len(); i++ {
		sel := mset.At(i)
		fn, ok := sel.Obj().(*types.Func)
		if !ok || fn.Name() != "Validate" {
			continue
		}
		sig, ok := fn.Type().(*types.Signature)
		if !ok {
			continue
		}
		return validateShape(sig)
	}
	return false
}

func validateShape(sig *types.Signature) bool {
	// The first parameter is context.Context.
	params := sig.Params()
	if params.Len() < 1 {
		return false
	}
	first := params.At(0)
	if !isContext(first.Type()) {
		return false
	}
	// The only result is error.
	results := sig.Results()
	if results.Len() != 1 {
		return false
	}
	res := results.At(0)
	return isError(res.Type())
}

func isContext(t types.Type) bool {
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	pkg := obj.Pkg()
	return pkg != nil && pkg.Path() == "context" && obj.Name() == "Context"
}

func isError(t types.Type) bool {
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	return obj.Pkg() == nil && obj.Name() == "error"
}
