// Package convnaming implements the converter rules:
//
//   - GID-105: converter names in convert packages are <Dst><Type>From<Src>
//     (EntityCreateSnapshotFromModel, ModelHelloOutFromEntity);
//   - GID-135: converter functions (the ...From... pattern) live in the
//     convert/ subpackage of their layer, not in service/repo/handler themselves.
//
// Exception: ctx helpers <Name>FromContext (GID-166) are not converters.
package convnaming

import (
	"go/ast"
	"regexp"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const (
	ruleNaming = "GID-105"
	rulePlace  = "GID-135"
)

// converterName: <Dst><Type>From<Src> — words before and after From.
var converterName = regexp.MustCompile(`^[A-Z][A-Za-z0-9]*From[A-Z][A-Za-z0-9]*$`)

// scopes — the layers in which converters must live in convert/.
var scopes = [][]string{
	{"dal"},
	{"domain"},
	{"server"},
	{"event"},
}

// Analyzer — GID rule: see Doc.
var Analyzer = &analysis.Analyzer{
	Name: "gidconvnaming",
	Doc:  ruleNaming + "/" + rulePlace + ": converters are named <Dst><Type>From<Src> and live in convert/ packages. Fix: rename to <Dst><Type>From<Src> and move into a convert/ subpackage",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	inConvert := pathseg.EndsWith(pass.Pkg.Path(), "convert")
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || !fn.Name.IsExported() {
				continue
			}
			if inConvert {
				checkConverterName(pass, fn)
				continue
			}
			checkConverterPlace(pass, fn)
		}
	}
	return nil, nil
}

// checkConverterName — GID-105: exported functions of a convert package
// are named <Dst><Type>From<Src>.
func checkConverterName(pass *analysis.Pass, fn *ast.FuncDecl) {
	if converterName.MatchString(fn.Name.Name) {
		return
	}
	pass.Reportf(fn.Name.Pos(),
		"%s: converter %q must be named <Dst><Type>From<Src>. Fix: rename it, e.g. EntityCreateSnapshotFromModel",
		ruleNaming, fn.Name.Name)
}

// checkConverterPlace — GID-135: a converter function outside a convert package.
func checkConverterPlace(pass *analysis.Pass, fn *ast.FuncDecl) {
	name := fn.Name.Name
	if !converterName.MatchString(name) || strings.HasSuffix(name, "FromContext") {
		return
	}
	if !inScope(pass.Pkg.Path()) {
		return
	}
	pass.Reportf(fn.Name.Pos(),
		"%s: converter %q must live in a convert/ subpackage of its layer. Fix: move it into convert/", rulePlace, name)
}

func inScope(pkgPath string) bool {
	for _, scope := range scopes {
		if pathseg.Contains(pkgPath, scope...) {
			return true
		}
	}
	return false
}
