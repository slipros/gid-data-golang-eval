// Package hastest implements rule GID-263: code that carries logic carries a
// test.
//
// The linter does not run anything. It reads the package's own _test.go files
// (a test lives in the same package as the code under test — GID-250) and asks
// one question about every non-trivial exported function: is it mentioned there
// at all? That is a deliberately one-sided proxy for coverage. "Never
// mentioned" proves "never executed"; "mentioned" proves nothing about which
// branches ran. The rule is built on the side that cannot raise a false alarm —
// the other side belongs to a coverage gate over `go test -coverprofile`, which
// needs the tests actually run and has no place in a linter's millisecond
// budget.
//
// Package variants, and why a package with NO tests goes unreported. With
// run.tests: true go/packages hands out two variants of a package that has
// tests: the base one (production files only) and "pkg [pkg.test]" (production
// files + _test.go). golangci-lint keeps only the second
// (filterDuplicatePackages), so a pass that holds _test.go files is the one to
// judge, and it is judged fully.
//
// A pass WITHOUT them is ambiguous: a package that has no tests looks exactly
// like a package whose tests were not handed over — the base variant, and every
// package of a run with run.tests: false. Separating the two takes a look at
// the file system, and looking there to judge tests the run deliberately
// excluded would be working around the setting rather than honouring it. So
// such a pass is left alone: under run.tests: false the rule reports nothing at
// all, and a package with no test file whatsoever is out of its reach in every
// mode. Naming those packages needs no source analysis anyway —
// `go list -f '{{if not .TestGoFiles}}{{.ImportPath}}{{end}}' ./...` prints
// them, and that belongs in CI next to the coverage gate.
//
// One transitive step, through a package-level var: a function called in the
// initializer of a var the tests use counts as exercised — var Analyzer =
// NewAnalyzer(Settings{}) runs NewAnalyzer at package initialization, so
// reporting it as untouched would be a false statement, and the value of this
// rule is that its statements hold. The step is not applied recursively:
// "reachable from a test" is real coverage, and no linter establishes that
// without running the tests.
//
// Candidates are top-level exported FuncDecls of production files with a
// NON-TRIVIAL body. Trivial — at most one statement, no binary operator, at
// most one call — covers the getter (return s.id), the enum String
// (return string(e)), the one-line delegation (return s.repo.Create(ctx, m))
// and the one-line constructor. A test over those asserts that the mock was
// called.
//
// Not judged: generated files, the composition root (package main, /app/**),
// the synthesized pkg.test package, packages under settings.exclude-paths, and
// functions on settings.exclude.
//
// _test.go is not a source of candidates: an exported helper of a suite is
// scaffolding, and demanding a test for the test is the reflex this rule must
// not have.
package hastest

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"

	"github.com/slipros/gid-data-golang-eval/internal/astwalk"
	"github.com/slipros/gid-data-golang-eval/internal/exclude"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
	"github.com/slipros/gid-data-golang-eval/internal/srcfile"
)

const ruleID = "GID-263"

// Analyzer — rule GID-263 with no exclusions configured.
var Analyzer = NewAnalyzer(Settings{})

// Settings — linter settings from .golangci.yml.
type Settings struct {
	// Exclude — functions that need no test of their own: "Function" (any
	// receiver) or "Type.Method".
	Exclude []string `json:"exclude"`
	// ExcludePaths — "/"-joined path-segment sequences; a package whose import
	// path contains such a sequence is not checked (generated clients, a
	// vendored subtree).
	ExcludePaths []string `json:"exclude-paths"`
}

// NewAnalyzer builds the GID-263 analyzer from linter settings (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "gidhastest",
		Doc: ruleID + ": a non-trivial exported function is exercised by a test of its own package. " +
			"Fix: add func TestCreateSegment(t *testing.T) calling it",
		Requires: astwalk.Requires,
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, s)
		},
	}
}

// candidate — an exported function that carries enough logic to deserve a test.
type candidate struct {
	decl *ast.FuncDecl
	obj  *types.Func
	recv string
}

// name renders the candidate the way the diagnostic names it: "Type.Method" for
// a method, the bare name for a function.
func (c candidate) name() string {
	if c.recv == "" {
		return c.decl.Name.Name
	}

	return c.recv + "." + c.decl.Name.Name
}

// kind is the word the diagnostic uses for the candidate.
func (c candidate) kind() string {
	if c.recv == "" {
		return "function"
	}

	return "method"
}

// testName is the test the fix example asks for: TestSegment_Rebuild for a
// method, TestBuildQuery for a function.
func (c candidate) testName() string {
	if c.recv == "" {
		return "Test" + c.decl.Name.Name
	}

	return "Test" + c.recv + "_" + c.decl.Name.Name
}

func run(pass *analysis.Pass, s Settings) (any, error) {
	if srcfile.IsTestBinaryPkg(pass) || isCompositionRoot(pass) || excludedPath(pass.Pkg.Path(), s.ExcludePaths) {
		return nil, nil
	}

	cands, prodFiles, hasTestFile := collect(pass, s.Exclude)
	if len(cands) == 0 {
		return nil, nil
	}

	if !hasTestFile {
		// The pass holds no _test.go. That is either a package without tests or
		// a run that did not hand them over (run.tests: false, and the base
		// variant of every package that does have tests), and nothing inside
		// the pass tells the two apart. Telling them apart would take a look at
		// the file system, which is precisely what the setting switched off, so
		// the rule judges only what it was given.
		return nil, nil
	}

	covered := coveredByTests(pass)
	spreadThroughVars(pass, prodFiles, covered)

	//nolint:gidallptr // the plugin does not depend on the internal gdhelper library
	for _, c := range cands {
		if covered.holds(c.obj) {
			continue
		}

		pass.Reportf(c.decl.Name.Pos(),
			"%s: exported %s %s is not exercised by any test of this package. "+
				"Fix: add func %s(t *testing.T) calling it",
			ruleID, c.kind(), c.name(), c.testName())
	}

	return nil, nil
}

// collect walks the package once and returns the candidates of its production
// files, those files themselves, and whether the pass holds a _test.go at all.
func collect(pass *analysis.Pass, excluded []string) (cands []candidate, prodFiles []*ast.File, hasTestFile bool) {
	skip := func(file *ast.File) bool {
		if srcfile.IsTest(pass, file) {
			hasTestFile = true

			return true
		}
		if ast.IsGenerated(file) {
			return true
		}
		prodFiles = append(prodFiles, file)

		return false
	}

	astwalk.NodesOf(pass, skip, func(_ *ast.File, fn *ast.FuncDecl) {
		if !fn.Name.IsExported() || trivial(fn) {
			return
		}

		recv := receiverType(fn)
		if exclude.Match(excluded, recv, fn.Name.Name) {
			return
		}

		obj, ok := pass.TypesInfo.Defs[fn.Name].(*types.Func)
		if !ok {
			return
		}

		cands = append(cands, candidate{decl: fn, obj: obj, recv: recv})
	})

	return cands, prodFiles, hasTestFile
}

// coverage — what the package's tests mention.
type coverage struct {
	objects map[types.Object]struct{}
	// ifaceMethods — names of methods called through an INTERFACE value in a
	// test. The object resolved there is the interface's method, not the
	// implementation's, so a concrete method reached that way would look
	// untouched. Matching by name lets such a call count: the rule gives up
	// precision to keep its no-false-alarm side.
	ifaceMethods map[string]struct{}
}

func (c coverage) holds(obj *types.Func) bool {
	if _, ok := c.objects[obj]; ok {
		return true
	}
	sig, ok := obj.Type().(*types.Signature)
	if !ok || sig.Recv() == nil {
		return false
	}
	_, ok = c.ifaceMethods[obj.Name()]

	return ok
}

// spreadThroughVars marks as covered the functions called in the initializer of
// a package-level var that the tests do use. One step only, and only through a
// var: the shape it exists for is
//
//	var Analyzer = NewAnalyzer(Settings{})
//
// where a test naming Analyzer runs NewAnalyzer at package initialization —
// saying it "is not exercised by any test" would be a false statement, and this
// rule's whole value is that its statements hold. The step is not applied
// recursively: "reachable from a test" is real coverage, which no linter can
// establish without running the tests.
func spreadThroughVars(pass *analysis.Pass, prodFiles []*ast.File, cov coverage) {
	for _, file := range prodFiles {
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.VAR {
				continue
			}

			for _, spec := range gen.Specs {
				value, isValue := spec.(*ast.ValueSpec)
				if !isValue || !anyNameCovered(pass, value.Names, cov) {
					continue
				}

				for _, expr := range value.Values {
					markCallees(pass, expr, cov)
				}
			}
		}
	}
}

// anyNameCovered reports whether the tests mention any of the names the var
// spec declares.
func anyNameCovered(pass *analysis.Pass, names []*ast.Ident, cov coverage) bool {
	for _, name := range names {
		obj := pass.TypesInfo.Defs[name]
		if obj == nil {
			continue
		}
		if _, ok := cov.objects[obj]; ok {
			return true
		}
	}

	return false
}

// markCallees adds every function called inside expr to the coverage set.
func markCallees(pass *analysis.Pass, expr ast.Expr, cov coverage) {
	// The walk is narrowed to one initializer expression, so the shared
	// inspector would cost more than it saves.
	ast.Inspect(expr, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if fn := typeutil.StaticCallee(pass.TypesInfo, call); fn != nil {
			cov.objects[fn] = struct{}{}
		}

		return true
	})
}

func coveredByTests(pass *analysis.Pass) coverage {
	cov := coverage{
		objects:      make(map[types.Object]struct{}),
		ifaceMethods: make(map[string]struct{}),
	}

	skip := func(file *ast.File) bool {
		return !srcfile.IsTest(pass, file)
	}

	astwalk.NodesOf(pass, skip, func(_ *ast.File, id *ast.Ident) {
		obj := pass.TypesInfo.Uses[id]
		if obj == nil {
			return
		}

		cov.objects[obj] = struct{}{}

		fn, ok := obj.(*types.Func)
		if !ok {
			return
		}
		sig, isSig := fn.Type().(*types.Signature)
		if !isSig {
			return
		}

		recv := sig.Recv()
		if recv == nil {
			return
		}
		if types.IsInterface(recv.Type()) {
			cov.ifaceMethods[fn.Name()] = struct{}{}
		}
	})

	return cov
}

// trivial reports whether the body carries no logic worth a test: at most one
// statement, no binary operator, no function literal, at most one call. That is
// the getter, the enum String, the one-line delegation and the one-line
// constructor — a test over them asserts that the mock was called.
func trivial(fn *ast.FuncDecl) bool {
	if fn.Body == nil || len(fn.Body.List) == 0 {
		return true
	}
	if len(fn.Body.List) > 1 {
		return false
	}

	switch stmt := fn.Body.List[0].(type) {
	case *ast.ReturnStmt:
		return simple(stmt.Results)
	case *ast.ExprStmt:
		return simple([]ast.Expr{stmt.X})
	case *ast.AssignStmt:
		return simple(stmt.Rhs)
	default:
		return false
	}
}

// simple reports whether the expressions hold no decision of their own: no
// binary operator, no closure, and at most one call in total (the delegation
// itself, or a conversion).
func simple(exprs []ast.Expr) bool {
	calls := 0
	plain := true

	for _, expr := range exprs {
		// The walk is already narrowed to a single expression, so the shared
		// inspector would cost more than it saves here.
		ast.Inspect(expr, func(n ast.Node) bool {
			switch n.(type) {
			case *ast.BinaryExpr, *ast.FuncLit:
				plain = false
			case *ast.CallExpr:
				calls++
			}

			return plain
		})
	}

	return plain && calls <= 1
}

// isCompositionRoot reports whether the package is wiring rather than logic:
// package main and /app/** assemble the dependencies, and that they assemble
// correctly is proved by the service starting, not by a unit test.
func isCompositionRoot(pass *analysis.Pass) bool {
	return pass.Pkg.Name() == "main" || pathseg.HasLayer(pass.Pkg.Path(), "app")
}

// excludedPath reports whether the package is exempted through
// settings.exclude-paths (a "/"-joined sequence of path segments).
func excludedPath(pkgPath string, excludes []string) bool {
	for _, e := range excludes {
		if pathseg.Contains(pkgPath, pathseg.Segments(strings.Trim(e, "/"))...) {
			return true
		}
	}

	return false
}

// receiverType is the name of the receiver's type, or "" for a function.
func receiverType(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return ""
	}

	expr := fn.Recv.List[0].Type
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}
	if index, ok := expr.(*ast.IndexExpr); ok { // a generic receiver: Cache[T]
		expr = index.X
	}
	if id, ok := expr.(*ast.Ident); ok {
		return id.Name
	}

	return ""
}
