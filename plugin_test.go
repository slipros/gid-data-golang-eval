package gidrules

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// typeSymbols — what an analyzer can only touch when the package is
// type-checked. Under register.LoadModeSyntax golangci-lint leaves pass.Pkg nil
// and pass.TypesInfo empty, so reading them panics the whole run — and only
// when the linter runs alone (--enable-only), because any other linter asking
// for LoadModeTypesInfo silently fills them in. Twice already a rule was
// shipped that way; this test is the gate.
var typeSymbols = []string{"pass.Pkg", "pass.TypesInfo", "modlayout.", "typeutil."}

// TestLoadModeMatchesAnalyzerNeeds checks every register.Plugin call in
// plugin.go: a linter whose analyzer package reads type information must be
// registered with LoadModeTypesInfo.
func TestLoadModeMatchesAnalyzerNeeds(t *testing.T) {
	plugins := parsePlugins(t, "plugin.go")
	if len(plugins) == 0 {
		t.Fatal("no register.Plugin calls found in plugin.go")
	}

	//nolint:gidallptr // the plugin does not depend on the internal gdhelper library
	for _, p := range plugins {
		t.Run(p.linter, func(t *testing.T) {
			if p.loadMode != "LoadModeSyntax" {
				return
			}

			if used := typeUsage(t, p.pkg); used != "" {
				t.Errorf("linter %q is registered with LoadModeSyntax, but package analyzers/%s reads %s — "+
					"a standalone run (--enable-only=%s) will panic on a nil pass.Pkg. "+
					"Fix: register it with register.LoadModeTypesInfo",
					p.linter, p.pkg, used, p.linter)
			}
		})
	}
}

// pluginReg — one register.Plugin call: the linter name, the analyzer package
// it comes from and the load mode it was registered with.
type pluginReg struct {
	linter   string
	pkg      string
	loadMode string
}

func parsePlugins(t *testing.T, filename string) []pluginReg {
	t.Helper()

	file, err := parser.ParseFile(token.NewFileSet(), filename, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", filename, err)
	}

	var out []pluginReg

	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || !isRegisterPlugin(call) || len(call.Args) != 2 {
			return true
		}

		name, ok := stringLit(call.Args[0])
		if !ok {
			return true
		}

		inner, ok := call.Args[1].(*ast.CallExpr)
		if !ok || len(inner.Args) < 2 {
			return true
		}

		out = append(out, pluginReg{
			linter:   name,
			pkg:      analyzerPkg(inner.Args[0]),
			loadMode: selectorName(inner.Args[len(inner.Args)-1]),
		})

		return true
	})

	return out
}

func isRegisterPlugin(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Plugin" {
		return false
	}

	ident, ok := sel.X.(*ast.Ident)

	return ok && ident.Name == "register"
}

// analyzerPkg — the analyzer package of an argument like patterns.TimeNowAnalyzer
// or entitygroup.Analyzer.
func analyzerPkg(arg ast.Expr) string {
	sel, ok := arg.(*ast.SelectorExpr)
	if !ok {
		return ""
	}

	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return ""
	}

	return ident.Name
}

func selectorName(arg ast.Expr) string {
	sel, ok := arg.(*ast.SelectorExpr)
	if !ok {
		return ""
	}

	return sel.Sel.Name
}

func stringLit(arg ast.Expr) (string, bool) {
	lit, ok := arg.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}

	v, err := strconv.Unquote(lit.Value)

	return v, err == nil
}

// typeUsage — the first type-dependent symbol the analyzer package reads, or an
// empty string when it is pure AST.
func typeUsage(t *testing.T, pkg string) string {
	t.Helper()

	if pkg == "" {
		return ""
	}

	dir := filepath.Join("analyzers", pkg)

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}

		code := string(data)
		for _, symbol := range typeSymbols {
			if strings.Contains(code, symbol) {
				return symbol
			}
		}
	}

	return ""
}
