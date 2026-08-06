// Package branchdispatch implements rule GID-261: a service method does not
// dispatch between two ways of fetching the same thing — it is one operation.
//
// The shape: an if/else whose branches call DIFFERENT methods of the SAME
// receiver and hand the results to the same place, so the method is really two
// methods behind one name, told apart by a condition on its input:
//
//	if in.OrganizationID.IsNil() {
//	    e, err = i.repo.Integration(ctx, in.ID)
//	} else {
//	    e, err = i.repo.IntegrationByOrganization(ctx, in.ID, in.OrganizationID)
//	}
//
// The caller already knows which of the two it wants — it is the one filling
// (or leaving empty) the field the condition reads. Splitting gives two honest
// methods (Integration, IntegrationByOrganization), each with the arguments its
// query actually needs, instead of a signature where a field is sometimes
// meaningful and sometimes not.
//
// Both branch forms are matched: assignment to the same left-hand side, and a
// bare return of the call. The receiver is compared as an expression
// (i.repo vs i.coreRepo are different receivers), and at least two distinct
// method names must appear — the same method called with different arguments is
// one operation with a prepared argument, not a dispatch.
//
// An else-if chain is judged as a whole and reported once, on the outermost if:
// three ways into one dependency is the same defect as two, only bigger.
//
// Scope: the root of /domain/service (owner decision 2026-08-06). A _test.go
// file is not judged: a double dispatching over its own state is test
// scaffolding (GID-250).
//
// Exceptions: //nolint:gidbranchdispatch, or settings.exclude
// ("Method" | "Type.Method").
package branchdispatch

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/exclude"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
	"github.com/slipros/gid-data-golang-eval/internal/srcfile"
)

const ruleID = "GID-261"

// Analyzer — rule GID-261 with default settings.
var Analyzer = NewAnalyzer(Settings{})

// Settings — settings of rule GID-261 from .golangci.yml.
type Settings struct {
	// Exclude — exclusions: "Method" (any type) or "Type.Method".
	Exclude []string `json:"exclude"`
}

// NewAnalyzer builds the GID-261 analyzer from the linter settings (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "gidbranchdispatch",
		Doc: ruleID + ": a service method must not dispatch between two methods of the same " +
			"dependency by a condition on its input. Fix: split it into two methods",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, s.Exclude)
		},
	}
}

func run(pass *analysis.Pass, excl []string) (any, error) {
	if !pathseg.EndsWith(pass.Pkg.Path(), "domain", "service") {
		return nil, nil
	}

	for _, file := range pass.Files {
		if ast.IsGenerated(file) || srcfile.IsTest(pass, file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			owner, method := ownerAndMethod(fn)
			if exclude.Match(excl, owner, method) {
				continue
			}
			checkBody(pass, fn.Body, owner, method)
		}
	}

	return nil, nil
}

// checkBody walks the function and reports every if/else that dispatches. An
// if that is itself the else of another one is skipped: the chain is judged
// from its outermost if, once.
func checkBody(pass *analysis.Pass, body *ast.BlockStmt, owner, method string) {
	chained := make(map[*ast.IfStmt]struct{})
	ast.Inspect(body, func(n ast.Node) bool {
		ifStmt, ok := n.(*ast.IfStmt)
		if !ok {
			return true
		}
		if next, ok := ifStmt.Else.(*ast.IfStmt); ok {
			chained[next] = struct{}{}
		}
		if _, isChained := chained[ifStmt]; isChained {
			return true
		}
		calls, ok := dispatchCalls(ifStmt)
		if !ok {
			return true
		}
		methods := make([]string, 0, len(calls))
		for _, c := range calls { //nolint:gidallptr // the plugin does not depend on the internal gdhelper library
			methods = append(methods, c.receiver+"."+c.method)
		}
		pass.Reportf(ifStmt.Pos(),
			"%s: %s picks between %s by a condition on its input — one method, several "+
				"operations. Fix: split it into a method per query (%s), each taking the arguments "+
				"its own query needs, and let the caller choose",
			ruleID, methodLabel(owner, method), joinAnd(methods), joinAnd(methodNames(calls)))

		return true
	})
}

// dispatchCalls collects the dependency calls of an if/else(-if) chain and
// reports whether they form a dispatch: every branch is a single call on the
// same receiver, and at least two distinct methods are involved.
func dispatchCalls(ifStmt *ast.IfStmt) ([]call, bool) {
	var calls []call
	for {
		c, ok := branchCall(ifStmt.Body)
		if !ok {
			return nil, false
		}
		calls = append(calls, c)
		switch next := ifStmt.Else.(type) {
		case *ast.IfStmt:
			ifStmt = next
		case *ast.BlockStmt:
			c, ok := branchCall(next)
			if !ok {
				return nil, false
			}
			calls = append(calls, c)

			return calls, isDispatch(calls)
		default:
			return nil, false // no else: a guard, not a dispatch
		}
	}
}

// isDispatch reports whether the collected calls go to one receiver through at
// least two distinct methods.
func isDispatch(calls []call) bool {
	first := calls[0]
	distinct := make(map[string]struct{}, len(calls))
	for _, c := range calls { //nolint:gidallptr // the plugin does not depend on the internal gdhelper library
		if !sameReceiver(first, c) {
			return false
		}
		distinct[c.method] = struct{}{}
	}

	return len(distinct) > 1
}

// methodNames returns the distinct method names in order of appearance.
func methodNames(calls []call) []string {
	seen := make(map[string]struct{}, len(calls))
	names := make([]string, 0, len(calls))
	for _, c := range calls { //nolint:gidallptr // the plugin does not depend on the internal gdhelper library
		if _, ok := seen[c.method]; ok {
			continue
		}
		seen[c.method] = struct{}{}
		names = append(names, c.method)
	}

	return names
}

// joinAnd spells a list the way the diagnostic reads: "a and b", "a, b and c".
func joinAnd(items []string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	default:
		return strings.Join(items[:len(items)-1], ", ") + " and " + items[len(items)-1]
	}
}

// call — a dependency call made by a branch.
type call struct {
	receiver string
	method   string
}

// branchCall returns the single call a branch consists of: either an
// assignment of one call, or a bare return of one call. A branch doing anything
// else is not a dispatch arm.
func branchCall(block *ast.BlockStmt) (call, bool) {
	if len(block.List) != 1 {
		return call{}, false
	}
	switch stmt := block.List[0].(type) {
	case *ast.AssignStmt:
		if len(stmt.Rhs) != 1 {
			return call{}, false
		}

		return selectorCall(stmt.Rhs[0])
	case *ast.ReturnStmt:
		if len(stmt.Results) != 1 {
			return call{}, false
		}

		return selectorCall(stmt.Results[0])
	default:
		return call{}, false
	}
}

// selectorCall reads a <receiver>.<method>(...) call off an expression.
func selectorCall(expr ast.Expr) (call, bool) {
	callExpr, ok := expr.(*ast.CallExpr)
	if !ok {
		return call{}, false
	}
	sel, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return call{}, false
	}

	return call{receiver: types.ExprString(sel.X), method: sel.Sel.Name}, true
}

// sameReceiver reports whether both branches call the same dependency.
func sameReceiver(a, b call) bool {
	return a.receiver != "" && a.receiver == b.receiver
}

// ownerAndMethod returns the receiver type name and the name of the function
// under analysis — the pair settings.exclude is written in.
func ownerAndMethod(fn *ast.FuncDecl) (owner, method string) {
	name := fn.Name.Name
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return "", name
	}
	expr := fn.Recv.List[0].Type
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return "", name
	}

	return ident.Name, name
}

// methodLabel spells the judged function for the diagnostic.
func methodLabel(owner, method string) string {
	if owner == "" {
		return "function " + method
	}

	return "method " + owner + "." + method
}
