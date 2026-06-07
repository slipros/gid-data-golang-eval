package errplace

import (
	"go/ast"
	"go/token"
	"path/filepath"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const errFileRuleID = "GID-169"

// errFileScopes — корневые пакеты слоёв, для которых действует GID-169.
// Только корни слоя (pathseg.EndsWith): подпакеты model/filter, entity/...
// не задеваются.
var errFileScopes = [][]string{
	{"domain", "model"},
	{"dal", "entity"},
}

// FileAnalyzer — вариант с дефолтным списком файлов (error.go, errors.go, err.go).
var FileAnalyzer = NewFileAnalyzer(Settings{})

// Settings — настройки правила GID-169 из .golangci.yml.
type Settings struct {
	// Files — имена файлов, где разрешены объявления ошибок слоя.
	// Пусто → дефолт ["error.go", "errors.go", "err.go"]
	// (err.go — каноничное имя из entity.md).
	Files []string `json:"files"`
}

// NewFileAnalyzer строит анализатор GID-169 с заданным списком файлов.
func NewFileAnalyzer(s Settings) *analysis.Analyzer {
	files := s.Files
	if len(files) == 0 {
		files = []string{"error.go", "errors.go", "err.go"}
	}
	allowed := make(map[string]struct{}, len(files))
	for _, f := range files {
		allowed[f] = struct{}{}
	}
	return &analysis.Analyzer{
		Name: "giderrfile",
		Doc:  errFileRuleID + ": layer errors live in a dedicated file (error.go/errors.go/err.go). Fix: move errors into error.go",
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

// checkErrFileVars сообщает о package-level var типа error в неположенном файле.
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
