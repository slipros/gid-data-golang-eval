// Package flagmain implements rule GID-192 (flags):
//
//   - Registering flags via the stdlib "flag" package (flag.String/Int/
//     Bool/.../flag.Parse/flag.Var functions and flag.FlagSet methods) is
//     allowed only in package main: Fix: declare flags in the binary, let libraries take parameters.
//   - In package main the flag name (the first string constant argument of
//     flag.String/Int/Bool/Duration/Float64/Var, etc.) must be in
//     snake_case: capital letters and hyphens are forbidden, digits and `_` are allowed.
//
// A camel-case name of the VARIABLE that receives the flag is not checked
// here — that is revive/ST1003 territory.
//
// The flag package is detected via TypesInfo (package path "flag"), so a
// local package named "flag" does not fall under the rule.
//
// Test files (*_test.go) and packages with the _test suffix are skipped: flag
// can be legitimate in tests. Generated code (ast.IsGenerated) is also
// skipped. LoadMode — TypesInfo.
package flagmain

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-192"

// flagPkgPath — the path of the stdlib flag registration package.
const flagPkgPath = "flag"

// Analyzer — rule GID-192: flag.* only in package main; flag names in snake_case.
var Analyzer = &analysis.Analyzer{
	Name: "gidflagmain",
	Doc:  ruleID + ": flags are registered only in package main, flag names in snake_case. Fix: register flags in main and use snake_case names",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	// A package with the _test suffix (an external test package) — skipped entirely.
	if strings.HasSuffix(pass.Pkg.Name(), "_test") {
		return nil, nil
	}
	isMain := pass.Pkg.Name() == "main"

	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		// *_test.go files are skipped — flag is legitimate in tests.
		if isTestFile(pass, file) {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			fn := flagFunc(pass, call)
			if fn == nil {
				return true
			}

			if !isMain {
				pass.Reportf(call.Pos(),
					"%s: registering a flag outside package main is forbidden. "+
						"Fix: declare flags in the binary, let libraries take parameters", ruleID)
				return true
			}

			// In package main the flag name is checked for snake_case.
			pos, name, ok := flagName(pass, fn, call)
			if !ok {
				return true
			}
			if !isSnakeCase(name) {
				pass.Reportf(pos, "%s: flag name %q. Fix: use snake_case", ruleID, name)
			}
			return true
		})
	}
	return nil, nil
}

// isTestFile reports that the file is *_test.go.
func isTestFile(pass *analysis.Pass, file *ast.File) bool {
	name := pass.Fset.Position(file.Pos()).Filename
	return strings.HasSuffix(name, "_test.go")
}

// flagFunc returns *types.Func if the call is a function of the flag package
// or a method of a type from the flag package (e.g. *flag.FlagSet). Otherwise nil.
func flagFunc(pass *analysis.Pass, call *ast.CallExpr) *types.Func {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil
	}
	fn, ok := pass.TypesInfo.ObjectOf(sel.Sel).(*types.Func)
	if !ok {
		return nil
	}
	sig, ok := fn.Type().(*types.Signature)
	if !ok {
		return nil
	}

	// A method of a type from the flag package (e.g. *flag.FlagSet).
	if recv := sig.Recv(); recv != nil {
		if isFlagType(recv.Type()) {
			return fn
		}
		return nil
	}

	// A package-level flag.* function.
	if pkg := fn.Pkg(); pkg != nil && pkg.Path() == flagPkgPath {
		return fn
	}
	return nil
}

// isFlagType reports whether the type belongs to the flag package (e.g. flag.FlagSet).
func isFlagType(t types.Type) bool {
	switch tt := t.(type) {
	case *types.Pointer:
		return isFlagType(tt.Elem())
	case *types.Alias:
		return isFlagType(types.Unalias(tt))
	case *types.Named:
		obj := tt.Obj()
		pkg := obj.Pkg()
		return pkg != nil && pkg.Path() == flagPkgPath
	}
	return false
}

// flagName extracts the flag name from a call to a flag-registering function.
// It returns the position of the name argument, the name itself and ok=true,
// only when the name is given as a string constant. For functions without a
// flag name (Parse, Parsed, Args, NArg, …) and for dynamic (non-constant)
// names ok=false.
func flagName(pass *analysis.Pass, fn *types.Func, call *ast.CallExpr) (token.Pos, string, bool) {
	idx, ok := nameArgIndex(fn.Name())
	if !ok || idx >= len(call.Args) {
		return token.NoPos, "", false
	}
	arg := call.Args[idx]
	tv, ok := pass.TypesInfo.Types[arg]
	if !ok || tv.Value == nil || tv.Value.Kind() != constant.String {
		return token.NoPos, "", false
	}
	return arg.Pos(), constant.StringVal(tv.Value), true
}

// nameArgIndex returns the index of the flag-name argument for a flag
// registration function/method. Groups:
//   - String/Int/Int64/Uint/Uint64/Float64/Bool/Duration/Func/BoolFunc
//     (name first) — index 0;
//   - Var/StringVar/IntVar/Int64Var/UintVar/Uint64Var/BoolVar/Float64Var/
//     DurationVar/TextVar (a pointer or Value comes first, the name second) —
//     index 1.
func nameArgIndex(name string) (int, bool) {
	switch name {
	case "String", "Int", "Int64", "Uint", "Uint64", "Float64",
		"Bool", "Duration", "Func", "BoolFunc":
		return 0, true
	case "Var", "StringVar", "IntVar", "Int64Var", "UintVar", "Uint64Var",
		"BoolVar", "Float64Var", "DurationVar", "TextVar":
		return 1, true
	}
	return 0, false
}

// isSnakeCase reports that the flag name is in snake_case: only lowercase
// letters, digits and `_`. Capital letters and hyphens are forbidden. An empty
// name is considered valid (not our concern).
func isSnakeCase(name string) bool {
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9':
		case r == '_':
		default:
			return false
		}
	}
	return true
}
