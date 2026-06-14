// Package bansymbol implements rule GID-217 (linter gidbansymbol):
// a configurable ban on specific symbols of third-party libraries.
//
// Source: repo.md — "Do not use gdpostgres.TQuery — direct conn methods are
// simpler and sufficient". By default the TQuery symbol from the
// gitlab.gid.team/gid-data/tech/golang/libs/postgres.git library is banned;
// the list can be overridden via settings.symbols in .golangci.yml.
//
// Detection: any *ast.SelectorExpr that resolves via pass.TypesInfo.Uses
// to an object with the given name from the given package. Generic
// instantiations (gdpostgres.TQuery[T](...)) resolve the same way and are caught too.
//
// Package match — by exact import path OR by a suffix of path segments
// (to cover versioned paths like .../v2). Generated code
// (ast.IsGenerated) is skipped.
package bansymbol

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-217"

// defaultSymbols — built-in list: ban on gdpostgres.TQuery.
var defaultSymbols = []Symbol{
	{
		Pkg:  "gitlab.gid.team/gid-data/tech/golang/libs/postgres.git",
		Name: "TQuery",
		Msg:  "gdpostgres.TQuery is banned. Fix: use conn methods directly: Select, ScanRow, NamedStruct or Transaction (repo.md)",
	},
}

// Analyzer — the variant with default settings.
var Analyzer = NewAnalyzer(Settings{})

// Symbol — a description of one banned symbol.
type Symbol struct {
	// Pkg — import path of the symbol's package. Matched exactly OR by a
	// suffix of path segments (e.g. ".../postgres.git" matches
	// ".../postgres.git/v2").
	Pkg string `json:"pkg"`
	// Name — name of the exported symbol (function, type, variable).
	Name string `json:"name"`
	// Msg — hint text for the diagnostic. Optional: without it a
	// generic wording is used.
	Msg string `json:"msg"`
}

// Settings — linter settings from .golangci.yml.
type Settings struct {
	// Symbols — the list of banned symbols. If empty, the built-in
	// default list is used.
	Symbols []Symbol `json:"symbols"`
}

// NewAnalyzer builds the GID-217 analyzer with the given settings.
func NewAnalyzer(cfg Settings) *analysis.Analyzer {
	symbols := cfg.Symbols
	if len(symbols) == 0 {
		symbols = defaultSymbols
	}
	return &analysis.Analyzer{
		Name: "gidbansymbol",
		Doc:  ruleID + ": ban specific library symbols (configurable). Fix: replace the banned symbol with the project-approved alternative.",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, symbols)
		},
	}
}

func run(pass *analysis.Pass, symbols []Symbol) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			obj := pass.TypesInfo.Uses[sel.Sel]
			if obj == nil || obj.Pkg() == nil {
				return true
			}
			pkg := obj.Pkg()
			objPkg := pkg.Path()
			objName := obj.Name()
			//nolint:gidallptr // the plugin does not depend on the internal gdhelper library
			for _, s := range symbols {
				if s.Name != objName {
					continue
				}
				if !pkgMatches(objPkg, s.Pkg) {
					continue
				}
				report(pass, sel.Sel.Pos(), s, pkg.Name(), objName)
				break
			}
			return true
		})
	}
	return nil, nil
}

// pkgMatches reports whether the symbol's package import path matches the setting:
// exact equality OR a suffix of path segments.
func pkgMatches(objPkg, want string) bool {
	if objPkg == want {
		return true
	}
	return pathseg.EndsWith(objPkg, pathseg.Segments(want)...)
}

func report(pass *analysis.Pass, pos token.Pos, s Symbol, pkgName, name string) {
	if s.Msg != "" {
		pass.Reportf(pos, "%s: %s", ruleID, s.Msg)
		return
	}
	pass.Reportf(pos,
		"%s: symbol %s.%s is banned by gidbansymbol. "+
			"Fix: replace it with the project-approved alternative.",
		ruleID, pkgName, name)
}
