// Package errswallow implements rule GID-258: in /domain/**, a function that
// handles an error must be able to RETURN one.
//
// Owner's principle: inside the application the signature is ours, so there is
// no reason for an error to end its life in a log line. A function that calls
// something failable and then only logs the failure decides, on the callee's
// behalf, that the failure does not matter — and the caller, reading a
// signature with no error in it, has no way to learn otherwise:
//
//	func (a *AdCabinetResolver) resolveChunk(ctx context.Context, chunk []uuid.UUID,
//		result map[uuid.UUID]model.AdCabinet) {
//		resp, err := a.registry.IntegrationsByIDs(ctx, req)
//		if err != nil {
//			a.logger.…Warn("resource registry unavailable…")
//			return
//		}
//
// A registry outage and an empty result are then indistinguishable upstream
// (incident 2026-08-06, advertising-api internal/domain/service/
// ad_cabinet_resolver.go — the same shape in AdCabinet, which returns
// (model.AdCabinet, bool) and drops the error in both branches).
//
// Detect, in packages under /domain/** only: a top-level FuncDecl whose result
// list has NO error, whose body checks an error value against nil
// (if err != nil, if err := f(); err != nil, switch err != nil) — that check is
// the proof the code KNOWS it can fail. Reported on the declaration: the fix is
// the signature, not the branch.
//
// Scope is /domain/** because that is where the signatures are ours to change.
// A transport handler answers to a foreign contract (http.HandlerFunc, a Kafka
// message handler, a gRPC interceptor) and has nowhere to return an error to —
// judging those would demand a nolint on every one of them.
//
// Not reported:
//   - a function that already returns error (it may still log, that is GID-155's
//     business, not this rule's);
//   - an error explicitly discarded (_ = f()) or never compared to nil — there
//     is no handling to speak of, and `errcheck` covers the discard;
//   - function LITERALS (a goroutine body, an errgroup closure, a defer):
//     their signature is fixed by whoever consumes them, so the fix this rule
//     asks for does not exist there. Only the enclosing FuncDecl is judged.
//   - _test.go: a test reports failures through *testing.T, not through a
//     result (GID-250).
//
// The soft-degradation case this rule collides with head-on — "an outage must
// NOT surface as an error, the caller substitutes a default" — is a real
// functional requirement now and then. It is declared per function
// (//nolint:giderrswallow or settings.exclude), so the decision is visible in
// the code instead of living in a doc comment.
package errswallow

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/exclude"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
	"github.com/slipros/gid-data-golang-eval/internal/srcfile"
)

const ruleID = "GID-258"

// domainScope — the layer this rule judges: /domain/** (service, usecase, and
// anything else the application layer holds).
var domainScope = []string{"domain"}

// Analyzer — rule GID-258 with no exclusions configured.
var Analyzer = NewAnalyzer(Settings{})

// Settings — linter settings from .golangci.yml.
type Settings struct {
	// Exclude — functions allowed to swallow the error because a functional
	// requirement demands soft degradation: "Function" (any receiver) or
	// "Type.Method".
	Exclude []string `json:"exclude"`
}

// NewAnalyzer builds the GID-258 analyzer from linter settings (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "giderrswallow",
		Doc: ruleID + ": in /domain/**, a function that checks an error must be able to return one — " +
			"logging it and moving on hides the failure from the caller. " +
			"Fix: add error to the results and return it",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, s.Exclude)
		},
	}
}

func run(pass *analysis.Pass, excluded []string) (any, error) {
	if !pathseg.HasLayer(pass.Pkg.Path(), domainScope...) {
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
			if exclude.Match(excluded, receiverType(fn), fn.Name.Name) {
				continue
			}
			checkFunc(pass, fn)
		}
	}
	return nil, nil
}

// checkFunc reports fn when it handles an error it cannot return.
func checkFunc(pass *analysis.Pass, fn *ast.FuncDecl) {
	if funcReturnsError(pass, fn) || !checksError(pass, fn.Body) {
		return
	}
	pass.Reportf(fn.Name.Pos(),
		"%s: the function checks an error but cannot return one — the failure ends in a log line and the "+
			"caller sees the same result as a successful empty answer. "+
			"Fix: add error to the results and return it (func resolve(ctx, ids) (map[K]V, error)); "+
			"if a functional requirement really demands soft degradation, declare it with "+
			"//nolint:giderrswallow on this function",
		ruleID)
}

// checksError reports whether body compares an error value against nil — the
// proof that the code knows the call can fail. Function literals are skipped:
// their signature belongs to whoever consumes them (a goroutine, an errgroup,
// a defer), so the fix this rule asks for is not available inside one.
func checksError(pass *analysis.Pass, body *ast.BlockStmt) bool {
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		if _, isLit := n.(*ast.FuncLit); isLit {
			return false
		}
		binary, ok := n.(*ast.BinaryExpr)
		if !ok {
			return true
		}
		if isErrorNilComparison(pass, binary) {
			found = true
		}
		return true
	})
	return found
}

// isErrorNilComparison reports whether expr compares an error-typed operand
// with nil (err != nil, err == nil, nil != err).
func isErrorNilComparison(pass *analysis.Pass, expr *ast.BinaryExpr) bool {
	if expr.Op.String() != "!=" && expr.Op.String() != "==" {
		return false
	}
	return isErrorAgainstNil(pass, expr.X, expr.Y) || isErrorAgainstNil(pass, expr.Y, expr.X)
}

func isErrorAgainstNil(pass *analysis.Pass, errSide, nilSide ast.Expr) bool {
	id, ok := nilSide.(*ast.Ident)
	if !ok || id.Name != "nil" {
		return false
	}
	return isErrorType(pass.TypesInfo.TypeOf(errSide))
}

// funcReturnsError reports whether fn's result list includes an error.
func funcReturnsError(pass *analysis.Pass, fn *ast.FuncDecl) bool {
	if fn.Type.Results == nil {
		return false
	}
	for _, field := range fn.Type.Results.List {
		if isErrorType(pass.TypesInfo.TypeOf(field.Type)) {
			return true
		}
	}
	return false
}

// receiverType returns the name of fn's receiver type ("Resolver" for both
// Resolver and *Resolver), or "" for a plain function — the form
// settings.exclude matches against.
func receiverType(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return ""
	}
	expr := fn.Recv.List[0].Type
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}
	if id, ok := expr.(*ast.Ident); ok {
		return id.Name
	}
	return ""
}

func isErrorType(t types.Type) bool {
	if t == nil {
		return false
	}
	errObj := types.Universe.Lookup("error")
	errType := errObj.Type()
	iface, ok := errType.Underlying().(*types.Interface)
	if !ok {
		return false
	}
	return types.Implements(t, iface)
}
