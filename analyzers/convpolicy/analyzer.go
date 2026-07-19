// Package convpolicy implements rule GID-247 (slug convert-no-policy, linter
// gidconvpolicy). Source: converter.md — "a converter is a pure mapping,
// input → output"; it must not make business decisions.
//
// Scope: packages whose import path ends with the "convert" segment
// (pathseg.EndsWith(pkgPath, "convert")) — same layer detection as GID-235
// (gidconvpure). Generated files are skipped.
//
// What is flagged: a function that branches on one of its input parameters
// (an if/switch whose condition references a parameter) and, across the
// branches, assigns to the same local variable two or more DISTINCT constant
// values of a BASIC (non-named) type. That is policy selection — the
// converter invents a raw domain value (a codec name, a sample rate, a
// channel count) and picks it by input — rather than copying a ready value
// from its input.
//
// Why "basic, non-named" only: an enum-to-enum mapping written as an
// if/switch also selects among constants, but those constants are of a NAMED
// enum type — that is legitimate vocabulary mapping (GID-143/233 push it
// toward a map, not toward this layer). Restricting to basic types (int,
// string, bool, float — including untyped literals and constant conversions
// such as uint32(1)) isolates the "raw magic value" case with a low
// false-positive rate.
//
// The fix: move the decision to /domain/model (a factory or a named policy
// value) and let the converter copy the ready value from the input.
//
// Escape hatch: //nolint:gidconvpolicy or settings.exclude
// (a function name, or Type.Method for a method converter).
package convpolicy

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/exclude"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-247"

// Analyzer — GID-247 with default settings.
var Analyzer = NewAnalyzer(Settings{})

// Settings — linter settings from .golangci.yml.
type Settings struct {
	// Exclude — converter functions exempt from the rule: a bare function
	// name (e.g. "asrFormatFromSource") or a "Type.Method" pair for a method
	// converter.
	Exclude []string `json:"exclude"`
}

// NewAnalyzer builds the GID-247 analyzer from the linter settings.
func NewAnalyzer(s Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "gidconvpolicy",
		Doc: ruleID + ": a convert function must not branch on its input to select a raw " +
			"constant value — that is business policy, not mapping",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, s.Exclude)
		},
	}
}

func run(pass *analysis.Pass, excludes []string) (any, error) {
	if !pathseg.EndsWith(pass.Pkg.Path(), "convert") {
		return nil, nil
	}
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			if exclude.Match(excludes, receiverType(fn), fn.Name.Name) {
				continue
			}
			checkFunc(pass, fn)
		}
	}
	return nil, nil
}

// span is a lexical range [lo, hi) covering the body of an input-conditioned
// branch (an if/else block or a switch case clause).
type span struct {
	lo, hi token.Pos
}

// constAssign records one assignment of a basic-typed constant to a local
// variable, with whether it sits inside an input-conditioned branch.
type constAssign struct {
	value    constant.Value
	pos      token.Pos
	inBranch bool
}

func checkFunc(pass *analysis.Pass, fn *ast.FuncDecl) {
	// minDistinct — the number of distinct constant values a variable must
	// take across input-conditioned branches to count as a policy selection.
	const minDistinct = 2

	params := paramObjects(pass, fn.Type)
	if len(params) == 0 {
		return
	}
	branches := branchSpans(pass, fn.Body, params)
	if len(branches) == 0 {
		return
	}
	assigns := constAssigns(pass, fn.Body, branches)

	for obj, list := range assigns {
		if !hasInBranch(list) {
			continue
		}
		if distinctValues(list) < minDistinct {
			continue
		}
		pass.Reportf(firstInBranchPos(list),
			"%s: convert function %q branches on input to select a constant value for %q — this is "+
				"business policy, not mapping. Fix: move the decision to /domain/model (a factory or a "+
				"named value) and copy the ready value from the input",
			ruleID, fn.Name.Name, obj.Name())
	}
}

// paramObjects returns the set of parameter variable objects of the function.
func paramObjects(pass *analysis.Pass, ft *ast.FuncType) map[types.Object]bool {
	out := make(map[types.Object]bool)
	if ft.Params == nil {
		return out
	}
	for _, field := range ft.Params.List {
		for _, name := range field.Names {
			if name.Name == "_" {
				continue
			}
			if obj := pass.TypesInfo.Defs[name]; obj != nil {
				out[obj] = true
			}
		}
	}
	return out
}

// branchSpans collects the body ranges of every branch whose selecting
// condition references an input parameter.
func branchSpans(pass *analysis.Pass, body *ast.BlockStmt, params map[types.Object]bool) []*span {
	var spans []*span
	ast.Inspect(body, func(n ast.Node) bool {
		switch s := n.(type) {
		case *ast.IfStmt:
			if refersParam(pass, s.Cond, params) {
				spans = append(spans, &span{s.Body.Pos(), s.Body.End()})
				if s.Else != nil {
					spans = append(spans, &span{s.Else.Pos(), s.Else.End()})
				}
			}
		case *ast.SwitchStmt:
			tagOnParam := s.Tag != nil && refersParam(pass, s.Tag, params)
			for _, stmt := range s.Body.List {
				clause, ok := stmt.(*ast.CaseClause)
				if !ok {
					continue
				}
				if tagOnParam || caseRefersParam(pass, clause.List, params) {
					spans = append(spans, &span{clause.Pos(), clause.End()})
				}
			}
		}
		return true
	})
	return spans
}

// constAssigns collects, per local variable, every assignment of a
// basic-typed constant, tagging those inside an input-conditioned branch.
func constAssigns(pass *analysis.Pass, body *ast.BlockStmt, branches []*span) map[types.Object][]*constAssign {
	out := make(map[types.Object][]*constAssign)
	ast.Inspect(body, func(n ast.Node) bool {
		as, ok := n.(*ast.AssignStmt)
		if !ok || len(as.Lhs) != 1 || len(as.Rhs) != 1 {
			return true
		}
		obj := localVar(pass, as)
		if obj == nil {
			return true
		}
		val, ok := basicConst(pass, as.Rhs[0])
		if !ok {
			return true
		}
		out[obj] = append(out[obj], &constAssign{
			value:    val,
			pos:      as.Pos(),
			inBranch: inAnySpan(as.Pos(), branches),
		})
		return true
	})
	return out
}

// localVar returns the variable object assigned by a single-target
// assignment (`:=` or `=`), or nil if the target is not a plain local.
func localVar(pass *analysis.Pass, as *ast.AssignStmt) types.Object {
	ident, ok := as.Lhs[0].(*ast.Ident)
	if !ok || ident.Name == "_" {
		return nil
	}
	var obj types.Object
	if as.Tok == token.DEFINE {
		obj = pass.TypesInfo.Defs[ident]
	} else {
		obj = pass.TypesInfo.Uses[ident]
	}
	if _, ok := obj.(*types.Var); !ok {
		return nil
	}
	return obj
}

// basicConst reports whether expr is a compile-time constant of a basic,
// non-named type (int/string/bool/float, incl. untyped literals and constant
// conversions like uint32(1)) and returns its value. Named types (enums) are
// rejected.
func basicConst(pass *analysis.Pass, expr ast.Expr) (constant.Value, bool) {
	tv := pass.TypesInfo.Types[expr]
	if tv.Value == nil || tv.Type == nil {
		return nil, false
	}
	if _, named := tv.Type.(*types.Named); named {
		return nil, false
	}
	if _, ok := tv.Type.Underlying().(*types.Basic); !ok {
		return nil, false
	}
	return tv.Value, true
}

func refersParam(pass *analysis.Pass, expr ast.Expr, params map[types.Object]bool) bool {
	found := false
	ast.Inspect(expr, func(n ast.Node) bool {
		ident, ok := n.(*ast.Ident)
		if !ok {
			return true
		}
		if params[pass.TypesInfo.Uses[ident]] {
			found = true
			return false
		}
		return true
	})
	return found
}

func caseRefersParam(pass *analysis.Pass, list []ast.Expr, params map[types.Object]bool) bool {
	for _, expr := range list {
		if refersParam(pass, expr, params) {
			return true
		}
	}
	return false
}

func inAnySpan(pos token.Pos, spans []*span) bool {
	for _, s := range spans {
		if pos >= s.lo && pos < s.hi {
			return true
		}
	}
	return false
}

func hasInBranch(list []*constAssign) bool {
	for _, a := range list {
		if a.inBranch {
			return true
		}
	}
	return false
}

// distinctValues counts the distinct constant values in the assignment list,
// keyed by ExactString (so uint32(1) and untyped 1 collapse to one value).
func distinctValues(list []*constAssign) int {
	seen := make(map[string]bool, len(list))
	for _, a := range list {
		seen[a.value.ExactString()] = true
	}
	return len(seen)
}

func firstInBranchPos(list []*constAssign) token.Pos {
	for _, a := range list {
		if a.inBranch {
			return a.pos
		}
	}
	return token.NoPos
}

// receiverType returns the receiver type name (without a pointer star) for a
// method, or "" for a plain function.
func receiverType(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return ""
	}
	expr := fn.Recv.List[0].Type
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name
	}
	return ""
}
