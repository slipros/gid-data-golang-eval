package errplace

import (
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"

	"golang.org/x/tools/go/analysis"
)

const errFileRuleID = "GID-169"

// FileAnalyzer — variant with the default file list (error.go).
var FileAnalyzer = NewFileAnalyzer(Settings{})

// Settings — settings of rule GID-169 from .golangci.yml.
type Settings struct {
	// Files — names of files where package error declarations are allowed.
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
		Doc:  errFileRuleID + ": package errors live in a dedicated file (error.go). Fix: move errors into error.go",
		Run: func(pass *analysis.Pass) (any, error) {
			return runErrFile(pass, allowed)
		},
	}
}

func runErrFile(pass *analysis.Pass, allowed map[string]struct{}) (any, error) {
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
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				if d.Tok != token.TYPE {
					continue
				}
				for _, spec := range d.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok || ts.Assign.IsValid() {
						continue
					}
					obj, ok := pass.TypesInfo.Defs[ts.Name].(*types.TypeName)
					if ok && isErrorType(obj) {
						pass.Reportf(ts.Pos(), "%s: error type %q is declared in %s. Fix: move the type and its methods to error.go", errFileRuleID, ts.Name.Name, fname)
					}
				}
			case *ast.FuncDecl:
				if d.Recv == nil || len(d.Recv.List) != 1 {
					continue
				}
				t := pass.TypesInfo.TypeOf(d.Recv.List[0].Type)
				if ptr, ok := t.(*types.Pointer); ok {
					t = ptr.Elem()
				}
				if named, ok := t.(*types.Named); ok && isErrorType(named.Obj()) {
					obj := named.Obj()
					pass.Reportf(d.Name.Pos(), "%s: method of error type %q is declared in %s. Fix: move the method to error.go", errFileRuleID, obj.Name(), fname)
				}
			}
		}
	}
	return nil, nil
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
					"%s: error %q is declared in %s. Fix: keep package errors in error.go",
					errFileRuleID, name.Name, fname)
			}
		}
	}
}
