// Package optsstyle implements rule GID-152: Options-type conventions.
//
//   - opts in function parameters is passed by pointer (*XxxOptions);
//   - opts in the entity body is stored as an unexported named field (opts Options / opts *Options);
//   - embedding an Options type (anonymous field) is a violation: it promotes option fields into the public API.
//
// Scope (configurable via .golangci.yml, see Settings):
//
//   - the rule is enforced only in the layers that own a constructor with opts —
//     by default handler packages and the domain service / usecase layers.
//     The config / composition layer (e.g. internal/app) is out of scope: it
//     legitimately holds library Options structs (httpserver.Options, …).
//   - only Options types declared in the SAME package as the struct/func are
//     considered. Options types from the standard library, module-cache
//     dependencies, or other first-party packages are never reported: GID-152
//     governs how an entity stores ITS OWN options, and external types cannot
//     be fixed at the use site anyway.
package optsstyle

import (
	"go/ast"
	"go/types"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-152"

// defaultLeaf / defaultWithin — layers where GID-152 is enforced by default.
//   - leaf: the package IS a handler leaf (path ends with "handler");
//     handler/convert and handler/validate subpackages are out of scope.
//   - within: the path contains the domain service / usecase segments.
var (
	defaultLeaf   = [][]string{{"handler"}}
	defaultWithin = [][]string{{"domain", "service"}, {"domain", "usecase"}}
)

// Analyzer — variant with the default scope.
var Analyzer = NewAnalyzer(Settings{})

// Settings — linter settings from .golangci.yml. Both lists are matched by
// path segments (never strings.Contains). If both are empty the defaults apply;
// setting either one replaces the whole default scope.
type Settings struct {
	// Leaf — layers matched by the package's trailing path segments: the
	// package is the layer root, not a subpackage. Default: [["handler"]].
	Leaf [][]string `json:"leaf"`
	// Within — layers matched anywhere in the package path.
	// Default: [["domain","service"], ["domain","usecase"]].
	Within [][]string `json:"within"`
}

// NewAnalyzer builds the GID-152 analyzer from the linter settings (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	leaf, within := s.Leaf, s.Within
	if len(leaf) == 0 && len(within) == 0 {
		leaf, within = defaultLeaf, defaultWithin
	}
	return &analysis.Analyzer{
		Name: "gidoptsstyle",
		Doc:  ruleID + ": opts is passed by pointer in parameters and stored as an unexported named field in the struct. Embedding opts is forbidden.",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, leaf, within)
		},
	}
}

func run(pass *analysis.Pass, leaf, within [][]string) (any, error) {
	if !inScope(pass.Pkg.Path(), leaf, within) {
		return nil, nil
	}
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

// inScope reports whether the package path belongs to a layer the rule guards.
func inScope(path string, leaf, within [][]string) bool {
	for _, seq := range leaf {
		if pathseg.EndsWith(path, seq...) {
			return true
		}
	}
	for _, seq := range within {
		if pathseg.Contains(path, seq...) {
			return true
		}
	}
	return false
}

// checkParams: an Options parameter by value is a violation.
func checkParams(pass *analysis.Pass, fn *ast.FuncDecl) {
	if fn.Type.Params == nil {
		return
	}
	for _, field := range fn.Type.Params.List {
		t := pass.TypesInfo.TypeOf(field.Type)
		if name, ok := localOptionsName(pass, t); ok {
			pass.Reportf(field.Pos(),
				"%s: opts must be passed by pointer. Fix: use *%s", ruleID, name)
		}
	}
}

// checkStructs inspects every struct field that involves a local Options type:
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
			name, ok2 := localOptionsName(pass, t)
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

// localOptionsName returns the type name if t is a named Options type declared
// in the package currently being analyzed. Options types from other packages —
// the standard library, dependencies in the module cache, or other first-party
// packages — are never reported (see the package doc).
func localOptionsName(pass *analysis.Pass, t types.Type) (string, bool) {
	named, ok := t.(*types.Named)
	if !ok {
		return "", false
	}
	obj := named.Obj()
	if obj.Pkg() == nil || obj.Pkg() != pass.Pkg {
		return "", false
	}
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
