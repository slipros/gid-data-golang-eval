// Package lgr — recognition of logrus types and calls
// (github.com/sirupsen/logrus) for logger rules.
package lgr

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// terminalMethods — logrus methods that emit a message to the log.
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

// IsType reports whether the type belongs to the logrus package
// (*logrus.Entry, *logrus.Logger, logrus.FieldLogger, etc.).
func IsType(t types.Type) bool {
	// pkgPath — the logrus package path.
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

// IsMethodSel reports whether the selector is a call to a method of a logrus type.
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

// IsTerminal reports whether the call is a terminal logrus call
// (Info/Error/...), and returns the method name.
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

// Chain collects the chain of logrus calls from the terminal call inward:
// the terminal plus all consecutive With* methods. Returns the selectors
// (from the terminal toward the start) and the base expression on which the chain begins.
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

// ChainNames returns the method names of the chain.
func ChainNames(sels []*ast.SelectorExpr) []string {
	names := make([]string, 0, len(sels))
	for _, s := range sels {
		names = append(names, s.Sel.Name)
	}
	return names
}
