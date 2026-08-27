// Package closurector implements rule GID-273: a function that builds and
// returns a closure is a constructor of it and carries the constructor prefix.
//
// Owner requirement 2026-08-27, consent-webhook-trigger:
//
//	func (w *WebhookTriggerV2) contactFilter(
//		scope model.WebhookTriggerV2ContactFilterScope,
//	) func(identifier *model.ConsentEventV2UserIdentifier) bool
//
// The method returns nothing of the domain — it assembles a predicate and
// hands it over. That is an inline constructor of a filter, so by GID-104 it
// is named newContactFilter: the reader sees at the call site that a thing is
// being built, not that a value is being fetched.
//
// Scope: /domain/** (every package under the layer — internal/domain/... and
// pkg/<module>/domain/... alike, matched by path segments, the GID-272
// boundary). The constructor convention is about the domain vocabulary; the
// transport layer names its closure factories after the ecosystem it plugs
// into — chi middleware is RequestID(), a gRPC interceptor is
// UnaryServerInterceptor(), a module publishes Router() — and renaming those
// would diverge from the library convention, not fix a defect. Measured on six
// udmp repos: judging every layer gave 52 findings, 24 of them Module.Router
// and 19 middleware factories, against 2 in /domain/** — both of the shape the
// rule exists for (DatasetHistory.buildAuthorEmailByID in lk-api).
//
// Judged: a function or method whose result list holds a FUNCTION TYPE (a bare
// func type or a named type over one) AND whose body actually builds it — a
// return of a function literal, directly or through a local variable bound to
// one. The second half is what separates a constructor from an accessor:
// "return w.filter" hands out a stored callback and is not judged, while
// "return func(...) bool { … }" makes one on the spot.
//
// Not judged:
//
//   - a name already carrying the constructor prefix (New/new by GID-104) —
//     that is the fix;
//   - the options convention (GID-126): a With-prefixed function, and any
//     function whose result is a named type with an Option/Options/Opt suffix.
//     WithTimeout(d) Option builds a closure on purpose and is named by its
//     own convention;
//   - a _test.go file (GID-250): a test helper returning a cleanup func
//     (setup() func()) is scaffolding named by the testing convention;
//   - generated code.
//
// Exceptions: //nolint:gidclosurector, or settings.exclude ("Function" |
// "Type.Method").
package closurector

import (
	"go/ast"
	"go/types"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/astwalk"
	"github.com/slipros/gid-data-golang-eval/internal/exclude"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
	"github.com/slipros/gid-data-golang-eval/internal/srcfile"
)

const ruleID = "GID-273"

// defaultOptionSuffixes — name suffixes of the options convention (GID-126): a
// function building a value of such a type is named by that convention, not by
// the constructor one.
var defaultOptionSuffixes = []string{"Options", "Option", "Opt"}

// Analyzer — rule GID-273 with default settings.
var Analyzer = NewAnalyzer(Settings{})

// Settings — settings of rule GID-273 from .golangci.yml.
type Settings struct {
	// OptionSuffixes — suffixes of the option types whose builders are named
	// by the options convention. Replaces the defaults when set.
	OptionSuffixes []string `json:"option-suffixes"`

	// Exclude — exclusions: "Function" (any receiver) or "Type.Method".
	Exclude []string `json:"exclude"`
}

// NewAnalyzer builds the GID-273 analyzer from the linter settings (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	suffixes := s.OptionSuffixes
	if len(suffixes) == 0 {
		suffixes = defaultOptionSuffixes
	}

	return &analysis.Analyzer{
		Name: "gidclosurector",
		Doc: ruleID + ": in /domain/**, a function building and returning a closure is an inline " +
			"constructor of it and takes the New/new prefix (GID-104). Fix: rename contactFilter to newContactFilter",
		Requires: astwalk.Requires,
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, suffixes, s.Exclude)
		},
	}
}

func run(pass *analysis.Pass, suffixes, excluded []string) (any, error) {
	if !pathseg.HasLayer(pass.Pkg.Path(), "domain") {
		return nil, nil
	}

	skip := func(file *ast.File) bool {
		return ast.IsGenerated(file) || srcfile.IsTest(pass, file)
	}

	astwalk.NodesOf(pass, skip, func(_ *ast.File, fn *ast.FuncDecl) {
		if fn.Body == nil || isCtorName(fn.Name.Name) || hasPrefixWord(fn.Name.Name, "With") {
			return
		}
		if exclude.Match(excluded, receiverType(fn), fn.Name.Name) {
			return
		}
		if !returnsFuncType(pass, fn, suffixes) || !buildsClosure(fn.Body) {
			return
		}

		pass.Reportf(fn.Name.Pos(),
			"%s: %s builds and returns a closure — an inline constructor of it, not a value of the layer. "+
				"Fix: rename it to %s (GID-104) (or //nolint:gidclosurector when the name is fixed by a convention)",
			ruleID, methodLabel(fn), ctorName(fn.Name.Name))
	})

	return nil, nil
}

// returnsFuncType reports whether one of fn's results is a function type — a
// bare func type or a named type over one. A result of an option type
// (settings.option-suffixes) is the options convention and does not count.
func returnsFuncType(pass *analysis.Pass, fn *ast.FuncDecl, suffixes []string) bool {
	if fn.Type.Results == nil {
		return false
	}

	for _, res := range fn.Type.Results.List {
		t := pass.TypesInfo.TypeOf(res.Type)
		if t == nil {
			continue
		}
		if named, ok := t.(*types.Named); ok {
			obj := named.Obj()
			if hasSuffix(obj.Name(), suffixes) {
				continue // an option type: the builder follows the options convention
			}
		}
		if _, ok := t.Underlying().(*types.Signature); ok {
			return true
		}
	}

	return false
}

// buildsClosure reports whether the body BUILDS the function it returns: a
// returned function literal, or a returned local variable bound to one. A body
// handing out a stored callback ("return w.filter") builds nothing and is an
// accessor, not a constructor.
//
// The walk does not descend into a function literal: the returns inside a
// closure belong to that closure, not to the function under judgement. It is
// narrowed to one function body, so it stays on ast.Inspect.
func buildsClosure(body *ast.BlockStmt) bool {
	bound := map[string]struct{}{}
	built := false

	ast.Inspect(body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.AssignStmt:
			for _, name := range assignedLiterals(node) {
				bound[name] = struct{}{}
			}
		case *ast.ValueSpec:
			for _, name := range declaredLiterals(node) {
				bound[name] = struct{}{}
			}
		case *ast.ReturnStmt:
			built = built || returnsLiteral(node, bound)
		case *ast.FuncLit:
			return false // the body of a closure is not this function's body
		}

		return !built
	})

	return built
}

// assignedLiterals returns the local names assigned a function literal.
func assignedLiterals(stmt *ast.AssignStmt) []string {
	var names []string

	for i, rhs := range stmt.Rhs {
		if _, ok := rhs.(*ast.FuncLit); !ok || i >= len(stmt.Lhs) {
			continue
		}
		if ident, ok := stmt.Lhs[i].(*ast.Ident); ok {
			names = append(names, ident.Name)
		}
	}

	return names
}

// declaredLiterals returns the names of a var declaration initialised with a
// function literal.
func declaredLiterals(spec *ast.ValueSpec) []string {
	var names []string

	for i, value := range spec.Values {
		if _, ok := value.(*ast.FuncLit); !ok || i >= len(spec.Names) {
			continue
		}
		names = append(names, spec.Names[i].Name)
	}

	return names
}

// returnsLiteral reports whether the return hands over a function built here.
func returnsLiteral(stmt *ast.ReturnStmt, bound map[string]struct{}) bool {
	for _, res := range stmt.Results {
		switch r := res.(type) {
		case *ast.FuncLit:
			return true
		case *ast.Ident:
			if _, ok := bound[r.Name]; ok {
				return true
			}
		}
	}

	return false
}

// ctorName spells the name the fix asks for: New for an exported function,
// new for an unexported one (GID-104).
func ctorName(name string) string {
	r, size := utf8.DecodeRuneInString(name)
	if size == 0 {
		return name
	}

	prefix := "new"
	if unicode.IsUpper(r) {
		prefix = "New"
	}

	return prefix + string(unicode.ToUpper(r)) + name[size:]
}

// isCtorName reports whether the name is a constructor by the GID-104
// convention: the New/new prefix followed by the entity name.
func isCtorName(name string) bool {
	return hasPrefixWord(name, "New") || hasPrefixWord(name, "new")
}

// hasPrefixWord reports whether name starts with prefix as a whole word — the
// next rune is uppercase, so "newest" is not a constructor and "Without" is not
// an option builder.
func hasPrefixWord(name, prefix string) bool {
	rest, ok := strings.CutPrefix(name, prefix)
	if !ok {
		return false
	}
	if rest == "" {
		return true
	}

	r, _ := utf8.DecodeRuneInString(rest)

	return unicode.IsUpper(r)
}

// hasSuffix reports whether name ends with one of the option-type suffixes.
func hasSuffix(name string, suffixes []string) bool {
	for _, s := range suffixes {
		if s != "" && strings.HasSuffix(name, s) {
			return true
		}
	}

	return false
}

// receiverType returns the name of fn's receiver type ("Trigger" for both
// Trigger and *Trigger), or "" for a plain function — the form
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
