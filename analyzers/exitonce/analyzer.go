// Package exitonce implements rule GID-181 (exit once / exit in main):
// the process terminates in exactly one place — in func main() of package main.
//
//   - os.Exit, log.Fatal* (std log), logrus.Fatal*/logrus.Exit
//     (github.com/sirupsen/logrus, including Entry/Logger methods)
//     are allowed ONLY in package main and only inside func main();
//   - within func main itself at most ONE such call is allowed.
//
// Any exit call outside func main means the error is not returned
// upward; a repeated call in main blurs the single exit point.
package exitonce

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-181"

// logrusPkgPath — the logrus package path.
const logrusPkgPath = "github.com/sirupsen/logrus"

// Analyzer — rule GID-181: os.Exit/log.Fatal*/logrus.Fatal* only once and only in func main. Fix: return an error up the call stack instead.
var Analyzer = &analysis.Analyzer{
	Name: "gidexitonce",
	Doc:  ruleID + ": os.Exit/log.Fatal*/logrus.Fatal* only once and only in func main. Fix: return an error up the call stack instead",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	isMainPkg := pass.Pkg.Name() == "main"
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		// mainBody — the body of the top-level func main() in package main.
		var mainBody *ast.BlockStmt
		if isMainPkg {
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if ok && fn.Recv == nil && fn.Name.Name == "main" && fn.Body != nil {
					mainBody = fn.Body
				}
			}
		}

		// First mark all exit calls located inside the body of main(),
		// so that the full file walk can tell them apart from calls outside main.
		inMain := map[*ast.CallExpr]struct{}{}
		if mainBody != nil {
			ast.Inspect(mainBody, func(n ast.Node) bool {
				if call, ok := n.(*ast.CallExpr); ok {
					if name, ok := exitName(pass, call); ok {
						_ = name
						inMain[call] = struct{}{}
					}
				}
				return true
			})
		}

		// mainCount — an ordinal counter of exit calls inside main().
		mainCount := 0
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			name, ok := exitName(pass, call)
			if !ok {
				return true
			}
			if _, ok := inMain[call]; ok {
				mainCount++
				if mainCount > 1 {
					pass.Reportf(call.Pos(),
						"%s: duplicate %s in main. Fix: exit the program in a single place",
						ruleID, name)
				}
				return true
			}
			pass.Reportf(call.Pos(),
				"%s: %s is forbidden outside func main. Fix: return an error up the call stack", ruleID, name)
			return true
		})
	}
	return nil, nil
}

// exitName recognizes an exit call (os.Exit / log.Fatal* / logrus.Fatal* / logrus.Exit,
// including methods of *logrus.Entry and *logrus.Logger) and returns a readable
// name for the diagnostic (e.g. "os.Exit", "log.Fatal", "logrus.Fatalf").
func exitName(pass *analysis.Pass, call *ast.CallExpr) (string, bool) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", false
	}
	fn, ok := pass.TypesInfo.ObjectOf(sel.Sel).(*types.Func)
	if !ok {
		return "", false
	}
	sig, ok := fn.Type().(*types.Signature)
	if !ok {
		return "", false
	}

	// A method of a logrus type (*logrus.Entry / *logrus.Logger): Fatal*.
	if recv := sig.Recv(); recv != nil {
		if isLogrusType(recv.Type()) && isFatalName(fn.Name()) {
			return "logrus." + fn.Name(), true
		}
		return "", false
	}

	// A package-level function.
	if fn.Pkg() == nil {
		return "", false
	}
	pkg := fn.Pkg()
	switch pkg.Path() {
	case "os":
		if fn.Name() == "Exit" {
			return "os.Exit", true
		}
	case "log":
		if isFatalName(fn.Name()) {
			return "log." + fn.Name(), true
		}
	case logrusPkgPath:
		if isFatalName(fn.Name()) || fn.Name() == "Exit" {
			return "logrus." + fn.Name(), true
		}
	}
	return "", false
}

// isFatalName reports that the method/function name belongs to the Fatal
// family (Fatal, Fatalf, Fatalln).
func isFatalName(name string) bool {
	switch name {
	case "Fatal", "Fatalf", "Fatalln":
		return true
	}
	return false
}

// isLogrusType reports whether the type belongs to the logrus package.
func isLogrusType(t types.Type) bool {
	switch tt := t.(type) {
	case *types.Pointer:
		return isLogrusType(tt.Elem())
	case *types.Alias:
		return isLogrusType(types.Unalias(tt))
	case *types.Named:
		obj := tt.Obj()
		pkg := obj.Pkg()
		return pkg != nil && pkg.Path() == logrusPkgPath
	}
	return false
}
