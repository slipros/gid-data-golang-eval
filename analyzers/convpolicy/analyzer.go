// Package convpolicy implements rule GID-247 (slug convert-no-policy, linter
// gidconvpolicy). Source: converter.md — "a converter is a pure mapping,
// input → output"; it must not make business decisions.
//
// Scope: packages whose import path ends with the "convert" segment
// (pathseg.EndsWith(pkgPath, "convert")) — same layer detection as GID-235
// (gidconvpure). Generated files and _test.go files are skipped (GID-250: a
// test double branching over its own state is scaffolding, not a converter
// deciding policy).
//
// What is flagged — two shapes of the same defect, a function that branches
// on one of its input parameters (an if/switch whose condition references a
// parameter) and, across the branches, assigns to the same local variable two
// or more DISTINCT constant values:
//
//  1. constants of a BASIC (non-named) type — the converter invents a raw
//     domain value (a codec name, a sample rate, a channel count) and picks
//     it by input, rather than copying a ready value from its input;
//  2. constants of one NAMED enum type, selected by a condition that does not
//     read an enum value — a length comparison, an emptiness/nil check, a
//     bool flag, a number or string comparison, a predicate call. The
//     decision "which reason do we show outside" is business policy, and the
//     converter is making it (incident 2026-07-xx, consent-webhook-trigger:
//     WebhookTriggerV2EmptyReason picked by len(UserIdentifiers) == 0).
//
// The criterion is what the condition READS, not the type of the selected
// constant: a branch whose condition reads only enum-typed values (a switch
// on an enum expression, a comparison against an enum constant) is an
// enum→enum mapping — legitimate vocabulary translation, GID-143/233 territory
// — and stays silent even though it selects among named enum constants. A
// condition that also reads any non-enum value (a basic-typed operand, a
// slice/pointer/nil check, a predicate call) is not a mapping, whatever else
// it reads.
//
// Basic-typed constants keep the original, looser behavior: they are flagged
// in any parameter-referencing branch, whether the condition reads an enum or
// not — the raw magic value is a defect on its own.
//
// The fix: move the decision to /domain/model (a clean predicate or a
// factory) and pass the converter the ready result.
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

	"github.com/slipros/gid-data-golang-eval/internal/astwalk"
	"github.com/slipros/gid-data-golang-eval/internal/exclude"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
	"github.com/slipros/gid-data-golang-eval/internal/srcfile"
)

const ruleID = "GID-247"

const (
	constNone constKind = iota
	constBasic
	constEnum
)

// Analyzer — GID-247 with default settings.
var Analyzer = NewAnalyzer(Settings{})

// constKind classifies a compile-time constant the rule can judge.
type constKind int

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
		Doc: ruleID + ": a convert function must not branch on its input to select a constant " +
			"value — a raw basic-typed value or a named enum constant picked by a non-enum " +
			"condition. That is business policy, not mapping",
		Requires: astwalk.Requires,
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, s.Exclude)
		},
	}
}

func run(pass *analysis.Pass, excludes []string) (any, error) {
	if !pathseg.EndsWith(pass.Pkg.Path(), "convert") {
		return nil, nil
	}
	astwalk.NodesOf(pass, func(file *ast.File) bool {
		return ast.IsGenerated(file) || srcfile.IsTest(pass, file)
	}, func(_ *ast.File, fn *ast.FuncDecl) {
		if fn.Body == nil {
			return
		}
		if exclude.Match(excludes, receiverType(fn), fn.Name.Name) {
			return
		}
		checkFunc(pass, fn)
	})
	return nil, nil
}

// span is a lexical range [lo, hi) covering the body of an input-conditioned
// branch (an if/else block or a switch case clause), with what its selecting
// condition reads.
type span struct {
	lo, hi token.Pos
	// nonEnum — the condition reads at least one non-enum-typed value (a
	// basic operand, a length/emptiness/nil check, a bool flag, a predicate
	// call). A condition reading only enum-typed values is enum→enum
	// mapping, not policy.
	nonEnum bool
}

// constAssign records one assignment of a constant to a local variable, with
// whether it sits inside a policy-worthy branch (for enum constants — a
// branch whose condition reads a non-enum value; for basic ones — any
// parameter-referencing branch).
type constAssign struct {
	value    constant.Value
	pos      token.Pos
	inBranch bool
}

func checkFunc(pass *analysis.Pass, fn *ast.FuncDecl) {
	// minDistinct — the number of distinct constant values a variable must
	// take across the branches to count as a policy selection.
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
		if basic := list.basic; basic != nil && hasInBranch(basic) && distinctValues(basic) >= minDistinct {
			pass.Reportf(firstInBranchPos(basic),
				"%s: convert function %q branches on input to select a constant value for %q — this is "+
					"business policy, not mapping. Fix: move the decision to /domain/model (a factory or a "+
					"named value) and copy the ready value from the input",
				ruleID, fn.Name.Name, obj.Name())
		}
		for named, enum := range list.enum {
			if !hasInBranch(enum) || distinctValues(enum) < minDistinct {
				continue
			}
			pass.Reportf(firstInBranchPos(enum),
				"%s: convert function %q branches on a non-enum condition to select among enum constants "+
					"of %q for %q — the converter decides a business question, not mapping. Fix: move the "+
					"decision to /domain/model (a clean predicate or a factory) and pass the converter the "+
					"ready result",
				ruleID, fn.Name.Name, named.Name(), obj.Name())
		}
	}
}

// assigns is the per-variable collection of judged constant assignments:
// basic-typed constants in one list, named-enum constants grouped by their
// named type (a selection is a choice within one enum, not across types).
type assigns struct {
	basic []*constAssign
	enum  map[types.Object][]*constAssign
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
// condition references an input parameter, tagging whether the condition
// reads a non-enum value.
func branchSpans(pass *analysis.Pass, body *ast.BlockStmt, params map[types.Object]bool) []*span {
	var spans []*span
	ast.Inspect(body, func(n ast.Node) bool {
		switch s := n.(type) {
		case *ast.IfStmt:
			if refersParam(pass, s.Cond, params) {
				spans = append(spans, &span{s.Body.Pos(), s.Body.End(), readsNonEnum(pass, s.Cond)})
				if s.Else != nil {
					// An else has no condition of its own; its verdict
					// follows the if — the pair is one decision.
					spans = append(spans, &span{s.Else.Pos(), s.Else.End(), readsNonEnum(pass, s.Cond)})
				}
			}
		case *ast.SwitchStmt:
			tagOnParam := s.Tag != nil && refersParam(pass, s.Tag, params)
			for _, stmt := range s.Body.List {
				clause, ok := stmt.(*ast.CaseClause)
				if !ok {
					continue
				}
				if !tagOnParam && !caseRefersParam(pass, clause.List, params) {
					continue
				}
				nonEnum := (tagOnParam && readsNonEnum(pass, s.Tag)) || caseReadsNonEnum(pass, clause.List)
				spans = append(spans, &span{clause.Pos(), clause.End(), nonEnum})
			}
		}
		return true
	})
	return spans
}

// constAssigns collects, per local variable, every assignment of a constant,
// split into basic-typed and per-enum-type lists, tagging those inside a
// policy-worthy branch.
func constAssigns(pass *analysis.Pass, body *ast.BlockStmt, branches []*span) map[types.Object]*assigns {
	out := make(map[types.Object]*assigns)
	ast.Inspect(body, func(n ast.Node) bool {
		as, ok := n.(*ast.AssignStmt)
		if !ok || len(as.Lhs) != 1 || len(as.Rhs) != 1 {
			return true
		}
		obj := localVar(pass, as)
		if obj == nil {
			return true
		}
		kind, val, named := classifyConst(pass, as.Rhs[0])
		if kind == constNone {
			return true
		}
		a := &constAssign{
			value:    val,
			pos:      as.Pos(),
			inBranch: inPolicyBranch(as.Pos(), branches, kind),
		}
		rec, ok := out[obj]
		if !ok {
			rec = &assigns{}
			out[obj] = rec
		}
		switch kind {
		case constBasic:
			rec.basic = append(rec.basic, a)
		case constEnum:
			key := named.Obj()
			if rec.enum == nil {
				rec.enum = make(map[types.Object][]*constAssign)
			}
			rec.enum[key] = append(rec.enum[key], a)
		}
		return true
	})
	return out
}

// inPolicyBranch reports whether pos sits inside a branch that makes the
// assignment a policy selection: for enum constants — a branch whose
// condition reads a non-enum value; for basic ones — any parameter-referencing
// branch.
func inPolicyBranch(pos token.Pos, spans []*span, kind constKind) bool {
	for _, s := range spans {
		if pos < s.lo || pos >= s.hi {
			continue
		}
		if kind == constBasic || s.nonEnum {
			return true
		}
	}
	return false
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

// classifyConst reports what a compile-time constant expression is: a raw
// basic-typed value (untyped literals and constant conversions like
// uint32(1)), a named-enum constant, or nothing the rule judges. A basic
// type under a named wrapper counts as the named (enum) side — the rule
// splits by what the condition reads, not by how the constant was spelled.
func classifyConst(pass *analysis.Pass, expr ast.Expr) (constKind, constant.Value, *types.Named) {
	tv := pass.TypesInfo.Types[expr]
	if tv.Value == nil || tv.Type == nil {
		return constNone, nil, nil
	}
	if _, ok := tv.Type.Underlying().(*types.Basic); !ok {
		return constNone, nil, nil
	}
	named, isNamed := tv.Type.(*types.Named)
	if isNamed {
		return constEnum, tv.Value, named
	}
	return constBasic, tv.Value, nil
}

// readsNonEnum reports whether expr contains a subexpression whose type is
// not a named (enum) type: a basic operand (length, bool flag, number or
// string comparison), a slice/pointer/map/struct/interface read
// (emptiness/nil check), a predicate call. The boolean result of a comparison
// or a logical operator is not a read — its operands are judged instead, so
// `source == SourceMeeting` reads only the enum.
func readsNonEnum(pass *analysis.Pass, expr ast.Expr) bool {
	if expr == nil {
		return false
	}
	found := false
	ast.Inspect(expr, func(n ast.Node) bool {
		if found {
			return false
		}
		switch n.(type) {
		case *ast.BinaryExpr, *ast.ParenExpr:
			return true
		}
		expr, ok := n.(ast.Expr)
		if !ok {
			return true
		}
		tv, ok := pass.TypesInfo.Types[expr]
		if !ok || tv.Type == nil {
			return true
		}
		if _, named := tv.Type.(*types.Named); !named {
			found = true
			return false
		}
		return true
	})
	return found
}

func caseReadsNonEnum(pass *analysis.Pass, list []ast.Expr) bool {
	for _, expr := range list {
		if readsNonEnum(pass, expr) {
			return true
		}
	}
	return false
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
