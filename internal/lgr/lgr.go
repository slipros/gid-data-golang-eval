// Package lgr — recognition of logger types and calls for the logger rules
// (GID-153/154/155/156/196/214). Two stacks are recognized, and the rules are
// written against whichever one the service uses:
//
//   - logrus (github.com/sirupsen/logrus): *logrus.Entry, *logrus.Logger,
//     logrus.FieldLogger; a chain of With* methods ending in a terminal call
//     (Info/Error/...);
//   - slog (log/slog): *slog.Logger; With/WithGroup instead of WithField, and
//     the context travels in the terminal call itself (InfoContext(ctx, ...))
//     rather than through WithContext.
//
// Kind tells the caller which stack a call belongs to, so that a rule can
// demand the right shape (WithContext for logrus, a *Context method for slog)
// without pinning the service to one library.
package lgr

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// Level method names shared by both stacks — spelled once so the two terminal
// sets below stay in sync.
//
//nolint:gidconstscope // GID-194: lgr is the shared logger vocabulary of the analyzers, not model/entity
const (
	// KindNone — not a logger.
	KindNone Kind = iota
	// KindLogrus — github.com/sirupsen/logrus.
	KindLogrus
	// KindSlog — the stdlib log/slog.
	KindSlog

	debugName = "Debug"
	infoName  = "Info"
	warnName  = "Warn"
	errorName = "Error"
	logName   = "Log"
)

// logrusTerminals — logrus methods that emit a message to the log.
var logrusTerminals = map[string]struct{}{
	"Trace": {}, "Tracef": {}, "Traceln": {},
	debugName: {}, "Debugf": {}, "Debugln": {},
	infoName: {}, "Infof": {}, "Infoln": {},
	"Print": {}, "Printf": {}, "Println": {},
	warnName: {}, "Warnf": {}, "Warnln": {},
	"Warning": {}, "Warningf": {}, "Warningln": {},
	errorName: {}, "Errorf": {}, "Errorln": {},
	"Fatal": {}, "Fatalf": {}, "Fatalln": {},
	"Panic": {}, "Panicf": {}, "Panicln": {},
	logName: {}, "Logf": {}, "Logln": {},
}

// slogTerminals — slog.Logger methods that emit a record. The *Context
// variants carry the context in the call itself — slog has no WithContext.
var slogTerminals = map[string]struct{}{
	debugName: {}, "DebugContext": {},
	infoName: {}, "InfoContext": {},
	warnName: {}, "WarnContext": {},
	errorName: {}, "ErrorContext": {},
	logName: {}, "LogAttrs": {},
}

// Kind — the logger stack a type or a call belongs to.
type Kind int

// TypeKind reports which logger stack the type belongs to (KindNone if none).
func TypeKind(t types.Type) Kind {
	switch tt := t.(type) {
	case *types.Pointer:
		return TypeKind(tt.Elem())
	case *types.Alias:
		return TypeKind(types.Unalias(tt))
	case *types.Named:
		obj := tt.Obj()
		pkg := obj.Pkg()
		if pkg == nil {
			return KindNone
		}
		const (
			logrusPath = "github.com/sirupsen/logrus"
			slogPath   = "log/slog"
		)
		switch pkg.Path() {
		case logrusPath:
			return KindLogrus
		case slogPath:
			return KindSlog
		}
	}
	return KindNone
}

// IsType reports whether the type is a logger of either stack.
func IsType(t types.Type) bool {
	return TypeKind(t) != KindNone
}

// MethodKind reports which stack the selector's receiver belongs to
// (KindNone when the selector is not a logger method call).
func MethodKind(pass *analysis.Pass, sel *ast.SelectorExpr) Kind {
	fn, ok := pass.TypesInfo.ObjectOf(sel.Sel).(*types.Func)
	if !ok {
		return KindNone
	}
	sig, ok := fn.Type().(*types.Signature)
	if !ok {
		return KindNone
	}
	recv := sig.Recv()
	if recv == nil {
		return KindNone
	}
	return TypeKind(recv.Type())
}

// IsMethodSel reports whether the selector is a call to a logger method.
func IsMethodSel(pass *analysis.Pass, sel *ast.SelectorExpr) bool {
	return MethodKind(pass, sel) != KindNone
}

// IsTerminal reports whether the call emits a message to the log, and returns
// the method name.
func IsTerminal(pass *analysis.Pass, call *ast.CallExpr) (string, bool) {
	name, _, ok := Terminal(pass, call)
	return name, ok
}

// Terminal is IsTerminal plus the stack the call belongs to — a rule that
// demands a stack-specific shape (WithContext vs InfoContext) needs both.
func Terminal(pass *analysis.Pass, call *ast.CallExpr) (name string, kind Kind, ok bool) {
	sel, isSel := call.Fun.(*ast.SelectorExpr)
	if !isSel {
		return "", KindNone, false
	}
	kind = MethodKind(pass, sel)
	if !isTerminalName(sel.Sel.Name, kind) {
		return "", KindNone, false
	}
	return sel.Sel.Name, kind, true
}

func isTerminalName(name string, kind Kind) bool {
	switch kind {
	case KindLogrus:
		_, ok := logrusTerminals[name]
		return ok
	case KindSlog:
		_, ok := slogTerminals[name]
		return ok
	case KindNone:
		return false
	}
	return false
}

// Chain collects the chain of logger calls from the terminal call inward: the
// terminal plus all consecutive enrichment methods (logrus With*, slog
// With/WithGroup). Returns the selectors (from the terminal toward the start)
// and the base expression the chain begins on.
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
