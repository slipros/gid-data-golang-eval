// Package initclean implements the GID-180 rule: init() is deterministic.
// Inside func init() the following are forbidden:
//   - starting goroutines (a go statement) — directly in the init body, including
//     nested blocks and closures declared within init itself;
//   - calls to functions of I/O packages (os, net, net/http, database/sql,
//     io/ioutil, bufio, etc.) — do background work and I/O from
//     main/constructor/app.
//
// Reading environment variables (os.Getenv, os.LookupEnv) is allowed — it is not I/O.
//
// The list of I/O packages is configured via settings.packages (replaces the default).
package initclean

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-180"

// defaultPackages — packages whose function calls are considered I/O.
// Matched by package path via TypesInfo.
var defaultPackages = []string{
	"os",
	"net",
	"net/http",
	"database/sql",
	"io/ioutil",
	"bufio",
}

// allowedFuncs — functions of I/O packages that are allowed in init().
// Reading env is deterministic and carries no I/O effect.
var allowedFuncs = map[string]map[string]bool{
	"os": {
		"Getenv":    true,
		"LookupEnv": true,
	},
}

// Analyzer — the variant with the default list of I/O packages.
var Analyzer = NewAnalyzer(Settings{})

// Settings — linter settings from .golangci.yml.
type Settings struct {
	// Packages — I/O packages (import paths). Replaces the default list.
	Packages []string `json:"packages"`
}

// NewAnalyzer builds the GID-180 analyzer from linter settings (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	pkgs := s.Packages
	if len(pkgs) == 0 {
		pkgs = defaultPackages
	}
	ioPkgs := make(map[string]bool, len(pkgs))
	for _, p := range pkgs {
		ioPkgs[p] = true
	}
	return &analysis.Analyzer{
		Name: "gidinitclean",
		Doc:  ruleID + ": init() must be deterministic, without goroutines or I/O calls. Fix: move that work to main/constructor",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, ioPkgs)
		},
	}
}

func run(pass *analysis.Pass, ioPkgs map[string]bool) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || fn.Name.Name != "init" || fn.Body == nil {
				continue
			}
			checkInitBody(pass, fn.Body, ioPkgs)
		}
	}
	return nil, nil
}

// checkInitBody walks the entire init() body (including nested blocks and
// the bodies of closures declared directly in init) and reports forbidden
// constructs.
func checkInitBody(pass *analysis.Pass, body *ast.BlockStmt, ioPkgs map[string]bool) {
	ast.Inspect(body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.GoStmt:
			pass.Reportf(node.Pos(),
				"%s: a goroutine in init() is forbidden; init must be deterministic. "+
					"Fix: start background work from main/app", ruleID)
		case *ast.CallExpr:
			pkg, fn := selectorPkgFunc(pass, node)
			if pkg == "" || !ioPkgs[pkg] {
				return true
			}
			if allowedFuncs[pkg][fn] {
				return true
			}
			pass.Reportf(node.Pos(),
				"%s: an I/O call %s.%s in init() is forbidden. "+
					"Fix: do it in main/constructor", ruleID, pkg, fn)
		}
		return true
	})
}

// selectorPkgFunc returns the package path and function name for a call of the
// form pkg.Func(...). For non-package calls (methods, local functions) it
// returns an empty package string.
func selectorPkgFunc(pass *analysis.Pass, call *ast.CallExpr) (pkgPath, fnName string) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", ""
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return "", ""
	}
	pkgName, ok := pass.TypesInfo.Uses[ident].(*types.PkgName)
	if !ok {
		return "", ""
	}
	imported := pkgName.Imported()
	return imported.Path(), sel.Sel.Name
}
