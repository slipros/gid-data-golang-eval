// Package arglimit implements rule GID-272: a function or method in the
// domain layer takes at most 3 substantive arguments.
//
// Owner's principle: a function in /domain/** that accepts many arguments
// stopped being a conversion and became assembly — it must be split, or a type
// carrying the related fields must be introduced. Incident 2026-08-27:
// WebhooksTriggersV2FromConsentEventV2 in consent-webhook-trigger takes six
// arguments, four of them maps keyed by the same organizationID — the four
// maps are one thing (the events of one organization) and ask to be grouped.
//
// Counted: package-level functions AND methods (the receiver does not count);
// each named parameter is one argument, a variadic tail is one. context.Context
// is not counted — it is a technical parameter, not a substantive one (owner
// decision). Constructors (the New/new prefix by GID-104) are not judged — a
// constructor takes as many dependencies as its entity has.
//
// Scope: every package under /domain/** — internal/domain/... and
// pkg/<module>/domain/... alike. A _test.go file is not judged (GID-250).
//
// Exceptions: //nolint:gidarglimit, or settings.exclude ("Function" |
// "Type.Method").
package arglimit

import (
	"go/ast"
	"go/types"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/astwalk"
	"github.com/slipros/gid-data-golang-eval/internal/exclude"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
	"github.com/slipros/gid-data-golang-eval/internal/srcfile"
)

const ruleID = "GID-272"

// Analyzer — rule GID-272 with default settings.
var Analyzer = NewAnalyzer(Settings{})

// Settings — settings of rule GID-272 from .golangci.yml.
type Settings struct {
	// MaxArgs — the maximum allowed number of substantive arguments
	// (context.Context excluded, receiver not counted, variadic tail one).
	// 0 → default (3). A violation begins at MaxArgs+1.
	MaxArgs int `json:"max-args"`

	// Exclude — exclusions: "Function" (any receiver) or "Type.Method".
	Exclude []string `json:"exclude"`
}

// NewAnalyzer builds the GID-272 analyzer from the linter settings (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	// defaultMaxArgs — the default maximum allowed number of substantive
	// arguments: 3 is allowed, a violation begins at 4.
	const defaultMaxArgs = 3

	maxArgs := s.MaxArgs
	if maxArgs < 1 {
		maxArgs = defaultMaxArgs
	}
	return &analysis.Analyzer{
		Name:     "gidarglimit",
		Doc:      ruleID + ": in /domain/**, a function or method takes more than " + strconv.Itoa(maxArgs) + " substantive arguments — a function with that many parameters stopped being a conversion and became assembly. Fix: split it or introduce a type that carries the related fields",
		Requires: astwalk.Requires,
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, maxArgs, s.Exclude)
		},
	}
}

func run(pass *analysis.Pass, maxArgs int, excluded []string) (any, error) {
	if !pathseg.HasLayer(pass.Pkg.Path(), "domain") {
		return nil, nil
	}
	skip := func(file *ast.File) bool {
		return ast.IsGenerated(file) || srcfile.IsTest(pass, file)
	}
	astwalk.NodesOf(pass, skip, func(_ *ast.File, fn *ast.FuncDecl) {
		if fn.Body == nil || isCtorName(fn.Name.Name) {
			return
		}
		if exclude.Match(excluded, receiverType(fn), fn.Name.Name) {
			return
		}
		count := countArgs(pass, fn)
		if count <= maxArgs {
			return
		}
		pass.Reportf(fn.Name.Pos(),
			"%s: %s takes %d substantive arguments (allowed: %d) — a function with that many parameters "+
				"stopped being a conversion and became assembly. "+
				"Fix: split it into smaller functions or introduce a type that groups the related fields "+
				"(func %s(in In) Out instead of func %s(a, b, c, d int) Out)",
			ruleID, methodLabel(fn), count, maxArgs, fn.Name.Name, fn.Name.Name)
	})

	return nil, nil
}

// countArgs counts the substantive arguments of fn: each named parameter is
// one argument, an unnamed one is one too, a variadic tail is one; the
// receiver and context.Context parameters are not counted.
func countArgs(pass *analysis.Pass, fn *ast.FuncDecl) int {
	if fn.Type.Params == nil {
		return 0
	}
	count := 0
	for _, field := range fn.Type.Params.List {
		if isContextType(pass.TypesInfo.TypeOf(field.Type)) {
			continue
		}
		if len(field.Names) == 0 {
			count++
			continue
		}
		count += len(field.Names)
	}
	return count
}

// isContextType reports whether t is the standard context.Context type.
func isContextType(t types.Type) bool {
	if t == nil {
		return false
	}
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	pkg := obj.Pkg()

	return pkg != nil && pkg.Path() == "context" && obj.Name() == "Context"
}

// isCtorName reports whether the name is a constructor by the GID-104
// convention: the New/new prefix followed by the entity name.
func isCtorName(name string) bool {
	rest, ok := strings.CutPrefix(name, "New")
	if !ok {
		if rest, ok = strings.CutPrefix(name, "new"); !ok {
			return false
		}
	}
	if rest == "" {
		return true
	}

	r, _ := utf8.DecodeRuneInString(rest)

	return unicode.IsUpper(r)
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
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return ""
	}

	return ident.Name
}

// methodLabel spells the judged function for the diagnostic.
func methodLabel(fn *ast.FuncDecl) string {
	if recv := receiverType(fn); recv != "" {
		return "method " + recv + "." + fn.Name.Name
	}

	return "function " + fn.Name.Name
}
