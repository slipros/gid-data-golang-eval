// Package lgr — распознавание типов и вызовов logrus
// (github.com/sirupsen/logrus) для logger-правил.
package lgr

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// terminalMethods — методы logrus, выводящие сообщение в лог.
var terminalMethods = map[string]struct{}{
	"Trace": {}, "Tracef": {}, "Traceln": {},
	"Debug": {}, "Debugf": {}, "Debugln": {},
	"Info": {}, "Infof": {}, "Infoln": {},
	"Print": {}, "Printf": {}, "Println": {},
	"Warn": {}, "Warnf": {}, "Warnln": {},
	"Warning": {}, "Warningf": {}, "Warningln": {},
	"Error": {}, "Errorf": {}, "Errorln": {},
	"Fatal": {}, "Fatalf": {}, "Fatalln": {},
	"Panic": {}, "Panicf": {}, "Panicln": {},
	"Log": {}, "Logf": {}, "Logln": {},
}

// IsType сообщает, относится ли тип к пакету logrus
// (*logrus.Entry, *logrus.Logger, logrus.FieldLogger и т.п.).
func IsType(t types.Type) bool {
	// pkgPath — путь пакета logrus.
	const pkgPath = "github.com/sirupsen/logrus"
	switch tt := t.(type) {
	case *types.Pointer:
		return IsType(tt.Elem())
	case *types.Alias:
		return IsType(types.Unalias(tt))
	case *types.Named:
		obj := tt.Obj()
		pkg := obj.Pkg()
		return pkg != nil && pkg.Path() == pkgPath
	}
	return false
}

// IsMethodSel сообщает, является ли селектор вызовом метода logrus-типа.
func IsMethodSel(pass *analysis.Pass, sel *ast.SelectorExpr) bool {
	fn, ok := pass.TypesInfo.ObjectOf(sel.Sel).(*types.Func)
	if !ok {
		return false
	}
	sig, ok := fn.Type().(*types.Signature)
	if !ok || sig.Recv() == nil {
		return false
	}
	recv := sig.Recv()
	return IsType(recv.Type())
}

// IsTerminal сообщает, является ли вызов терминальным logrus-вызовом
// (Info/Error/...), и возвращает имя метода.
func IsTerminal(pass *analysis.Pass, call *ast.CallExpr) (string, bool) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", false
	}
	if _, ok := terminalMethods[sel.Sel.Name]; !ok {
		return "", false
	}
	if !IsMethodSel(pass, sel) {
		return "", false
	}
	return sel.Sel.Name, true
}

// Chain собирает цепочку logrus-вызовов от терминального call вглубь:
// терминал + все последовательные With*-методы. Возвращает селекторы
// (от терминала к началу) и базовое выражение, на котором начата цепочка.
func Chain(pass *analysis.Pass, call *ast.CallExpr) (sels []*ast.SelectorExpr, base ast.Expr) {
	cur := ast.Expr(call)
	for {
		c, ok := cur.(*ast.CallExpr)
		if !ok {
			break
		}
		sel, ok := c.Fun.(*ast.SelectorExpr)
		if !ok || !IsMethodSel(pass, sel) {
			break
		}
		if len(sels) > 0 && !strings.HasPrefix(sel.Sel.Name, "With") {
			break
		}
		sels = append(sels, sel)
		cur = sel.X
	}
	if len(sels) == 0 {
		return nil, nil
	}
	return sels, sels[len(sels)-1].X
}

// ChainNames возвращает имена методов цепочки.
func ChainNames(sels []*ast.SelectorExpr) []string {
	names := make([]string, 0, len(sels))
	for _, s := range sels {
		names = append(names, s.Sel.Name)
	}
	return names
}
