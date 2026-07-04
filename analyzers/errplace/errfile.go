package errplace

import (
	"go/ast"
	"go/token"
	"path/filepath"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const errFileRuleID = "GID-169"

// errFileScopes — root layer packages where GID-169 applies.
// Layer roots only (pathseg.EndsWith): subpackages model/filter, entity/...
// are not touched.
var errFileScopes = [][]string{
	{"domain", "model"},
	{"dal", "entity"},
}

// FileAnalyzer — variant with the default file list (error.go).
var FileAnalyzer = NewFileAnalyzer(Settings{})

// Settings — settings of rule GID-169 from .golangci.yml.
type Settings struct {
	// Files — names of files where layer error declarations are allowed.
	// Empty → default ["error.go"] (the canonical layer-errors file name).
	Files []string `json:"files"`
}

// NewFileAnalyzer builds the GID-169 analyzer with the given file list.
func NewFileAnalyzer(s Settings) *analysis.Analyzer {
	files := s.Files
	if len(files) == 0 {
		files = []string{"error.go"}
	}
	allowed := make(map[string]struct{}, len(files))
	for _, f := range files {
		allowed[f] = struct{}{}
	}
	return &analysis.Analyzer{
		Name: "giderrfile",
		Doc:  errFileRuleID + ": layer errors live in a dedicated file (error.go). Fix: move errors into error.go",
		Run: func(pass *analysis.Pass) (any, error) {
			return runErrFile(pass, allowed)
		},
	}
}

func runErrFile(pass *analysis.Pass, allowed map[string]struct{}) (any, error) {
	if !inErrFileScope(pass.Pkg.Path()) {
		return nil, nil
	}
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		fname := filepath.Base(pass.Fset.Position(file.Pos()).Filename)
		if isTestFile(fname) {
			continue
		}
		if _, ok := allowed[fname]; ok {
			continue
		}
		checkErrFileVars(pass, fname, file)
	}
	return nil, nil
}

func inErrFileScope(pkgPath string) bool {
	for _, scope := range errFileScopes {
		if pathseg.EndsWith(pkgPath, scope...) {
			return true
		}
	}
	return false
}

func isTestFile(fname string) bool {
	return len(fname) > len("_test.go") &&
		fname[len(fname)-len("_test.go"):] == "_test.go"
}

// checkErrFileVars reports a package-level var of type error in a wrong file.
func checkErrFileVars(pass *analysis.Pass, fname string, file *ast.File) {
	for _, decl := range file.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.VAR {
			continue
		}
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
				if obj == nil || !implementsError(obj.Type()) {
					continue
				}
				pass.Reportf(name.Pos(),
					"%s: error %q is declared in %s. Fix: keep layer errors in error.go",
					errFileRuleID, name.Name, fname)
			}
		}
	}
}
