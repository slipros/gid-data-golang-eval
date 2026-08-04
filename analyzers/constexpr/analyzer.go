// Package constexpr implements rule GID-254: a local variable initialized
// with a CONSTANT string expression that is not a lone literal — a
// concatenation of constants, or a reference to another constant — must be
// declared as a const.
//
// The case that produced the rule (resource-registry, 2026-08-04):
//
//	const integrationColumns = `id, organization_id, …`
//
//	sqlQuery := "SELECT " + integrationColumns + " FROM integration WHERE id = @id"
//
// Every operand is a constant, so the whole expression is a constant
// expression the compiler folds at build time — nothing about it is dynamic.
// Written with ":=" it reads as if the query were assembled at run time, it
// can be reassigned by mistake, and it drifts away from the sibling methods
// of the same repository, which already say `const sqlQuery = …`. The fix is
// one keyword: `const sqlQuery = "SELECT " + integrationColumns + " FROM …"`.
//
// Detect: inside a function body, a declaration (`x := <expr>` or
// `var x = <expr>`) where ALL of
//   - the initializer's constant VALUE is known (types.Info.Types[expr].Value
//     != nil — the whole expression is a constant expression), AND
//   - its type is string-kinded (a named string type counts: a typed constant
//     is still a constant), AND
//   - the initializer is NOT a lone *ast.BasicLit — a bare `msg := "hi"` is
//     left alone on purpose: the rule targets expressions that LOOK dynamic
//     (a concatenation, a constant reference) while being constant, not every
//     local string in the codebase, AND
//   - the variable is never assigned again and its address is never taken —
//     either would make `const` impossible.
//
// A declaration in the init statement of if/for/switch is skipped: Go has no
// place for a const declaration there.
//
// Only string constants are in scope. Numeric constant expressions
// (`timeout := 5 * 60`) are deliberately left out for now — they are rare in
// our code and carry a higher noise risk; extending the kind check is the
// obvious next step if the team wants them.
//
// Generated code and _test.go files (fixtures and table data legitimately
// build strings this way) are skipped.
package constexpr

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-254"

// Analyzer — rule GID-254.
var Analyzer = NewAnalyzer()

// NewAnalyzer builds the GID-254 analyzer.
func NewAnalyzer() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "gidconstexpr",
		Doc: ruleID + ": a local variable initialized with a constant string expression (a concatenation of " +
			"constants) must be a const. Fix: const sqlQuery = \"SELECT \" + columns + \" FROM t\"",
		Run: run,
	}
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) || isTestFile(pass, file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			checkFunc(pass, fn)
		}
	}
	return nil, nil
}

func checkFunc(pass *analysis.Pass, fn *ast.FuncDecl) {
	mutated := mutatedObjects(pass, fn)
	inits := initStatements(fn)
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch stmt := n.(type) {
		case *ast.AssignStmt:
			if stmt.Tok != token.DEFINE || inits[stmt] {
				return true
			}
			checkPairs(pass, stmt.Lhs, stmt.Rhs, mutated)
		case *ast.DeclStmt:
			genDecl, ok := stmt.Decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.VAR {
				return true
			}
			for _, spec := range genDecl.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				names := make([]ast.Expr, 0, len(valueSpec.Names))
				for _, name := range valueSpec.Names {
					names = append(names, name)
				}
				checkPairs(pass, names, valueSpec.Values, mutated)
			}
		}
		return true
	})
}

// checkPairs reports every name whose paired initializer is a constant string
// expression worth a const declaration. A multi-value initializer (a, b :=
// f()) has no per-name expression and is skipped.
func checkPairs(pass *analysis.Pass, names, values []ast.Expr, mutated map[types.Object]bool) {
	if len(names) != len(values) {
		return
	}
	for i, name := range names {
		ident, ok := name.(*ast.Ident)
		if !ok || ident.Name == "_" {
			continue
		}
		obj := pass.TypesInfo.Defs[ident]
		if obj == nil || mutated[obj] {
			continue
		}
		if !isConstStringExpr(pass, values[i]) {
			continue
		}
		pass.Reportf(ident.Pos(),
			"%s: %q is initialized with a constant string expression — every operand is a constant, so the "+
				"value is folded at compile time and nothing here is dynamic. Fix: declare it as a const "+
				"(const %s = \"SELECT \" + columns + \" FROM t\")",
			ruleID, ident.Name, ident.Name)
	}
}

// isConstStringExpr reports whether expr is a constant expression of string
// kind that is not a lone literal — a concatenation of constants
// ("SELECT " + columns) or a reference to a constant (columns).
func isConstStringExpr(pass *analysis.Pass, expr ast.Expr) bool {
	if _, isLiteral := expr.(*ast.BasicLit); isLiteral {
		return false
	}
	tv, ok := pass.TypesInfo.Types[expr]
	if !ok || tv.Value == nil {
		return false
	}
	basic, ok := tv.Type.Underlying().(*types.Basic)
	return ok && basic.Info()&types.IsString != 0
}

// mutatedObjects collects the objects that cannot become a const: those
// assigned after their declaration (including ++/--) and those whose address
// is taken.
func mutatedObjects(pass *analysis.Pass, fn *ast.FuncDecl) map[types.Object]bool {
	out := map[types.Object]bool{}
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.AssignStmt:
			if node.Tok == token.DEFINE {
				return true
			}
			for _, lhs := range node.Lhs {
				markObject(pass, lhs, out)
			}
		case *ast.IncDecStmt:
			markObject(pass, node.X, out)
		case *ast.UnaryExpr:
			if node.Op == token.AND {
				markObject(pass, node.X, out)
			}
		}
		return true
	})
	return out
}

// markObject records the object expr refers to when it is a plain identifier.
func markObject(pass *analysis.Pass, expr ast.Expr, out map[types.Object]bool) {
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return
	}
	if obj := pass.TypesInfo.Uses[ident]; obj != nil {
		out[obj] = true
		return
	}
	if obj := pass.TypesInfo.Defs[ident]; obj != nil {
		out[obj] = true
	}
}

// initStatements collects the assignments that serve as the init statement of
// an if/for/switch — Go allows no const declaration in that position.
func initStatements(fn *ast.FuncDecl) map[*ast.AssignStmt]bool {
	out := map[*ast.AssignStmt]bool{}
	mark := func(stmt ast.Stmt) {
		if assign, ok := stmt.(*ast.AssignStmt); ok {
			out[assign] = true
		}
	}
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.IfStmt:
			mark(node.Init)
		case *ast.ForStmt:
			mark(node.Init)
			mark(node.Post)
		case *ast.SwitchStmt:
			mark(node.Init)
		case *ast.TypeSwitchStmt:
			mark(node.Init)
			mark(node.Assign)
		}
		return true
	})
	return out
}

// isTestFile reports whether the file is a _test.go file.
func isTestFile(pass *analysis.Pass, file *ast.File) bool {
	return strings.HasSuffix(pass.Fset.Position(file.Pos()).Filename, "_test.go")
}
