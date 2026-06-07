// Package exitonce реализует правило GID-181 (exit once / exit in main):
// процесс завершается ровно в одном месте — в функции main() пакета main.
//
//   - os.Exit, log.Fatal* (std log), logrus.Fatal*/logrus.Exit
//     (github.com/sirupsen/logrus, включая методы Entry/Logger)
//     разрешены ТОЛЬКО в пакете main и только в функции main();
//   - в самой func main допускается не более ОДНОГО такого вызова.
//
// Любой exit-вызов вне func main означает, что ошибка не возвращается
// наверх; повторный вызов в main размывает единственную точку выхода.
package exitonce

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-181"

// logrusPkgPath — путь пакета logrus.
const logrusPkgPath = "github.com/sirupsen/logrus"

// Analyzer — правило GID-181: os.Exit/log.Fatal*/logrus.Fatal* only once and only in func main. Fix: return an error up the call stack instead.
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
		// mainBody — тело верхнеуровневой функции main() в пакете main.
		var mainBody *ast.BlockStmt
		if isMainPkg {
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if ok && fn.Recv == nil && fn.Name.Name == "main" && fn.Body != nil {
					mainBody = fn.Body
				}
			}
		}

		// Сначала отмечаем все exit-вызовы, лежащие внутри тела main(),
		// чтобы при общем обходе файла отличать их от вызовов вне main.
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

		// mainCount — порядковый счётчик exit-вызовов внутри main().
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

// exitName распознаёт exit-вызов (os.Exit / log.Fatal* / logrus.Fatal* / logrus.Exit,
// в т.ч. методы *logrus.Entry и *logrus.Logger) и возвращает читаемое имя для
// диагностики (например "os.Exit", "log.Fatal", "logrus.Fatalf").
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

	// Метод logrus-типа (*logrus.Entry / *logrus.Logger): Fatal*.
	if recv := sig.Recv(); recv != nil {
		if isLogrusType(recv.Type()) && isFatalName(fn.Name()) {
			return "logrus." + fn.Name(), true
		}
		return "", false
	}

	// Пакетная функция.
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

// isFatalName сообщает, что имя метода/функции — Fatal-семейство
// (Fatal, Fatalf, Fatalln).
func isFatalName(name string) bool {
	switch name {
	case "Fatal", "Fatalf", "Fatalln":
		return true
	}
	return false
}

// isLogrusType сообщает, относится ли тип к пакету logrus.
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
