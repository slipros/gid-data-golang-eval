// Package cliflags implements two urfave/cli/v3 flag hygiene rules.
//
// GID-238 (cli-flag-naming):
//   - the Name field of a flag literal must be kebab-case:
//     [a-z0-9]+(-[a-z0-9]+)*;
//   - environment variable names — cli.EnvVars("...") passed to the Sources
//     field, or a literal EnvVars []string field (both v3 styles are
//     accepted) — must be UPPER_SNAKE_CASE: [A-Z0-9]+(_[A-Z0-9]+)*.
//
// GID-239 (cli-flag-required):
//   - a flag literal with neither Required: true nor an explicit Value
//     field is flagged. Agents wiring a flag into the app tend to forget
//     Required, and a flag consumed downstream with no default silently
//     zero-values. A Value field counts as a default no matter what it
//     holds — an explicit zero (Value: 0) is still a deliberate default.
//
// Applicability: any composite literal (bare or "&cli.XxxFlag{...}") whose
// type is a named or aliased type — StringFlag, IntFlag, BoolFlag, the
// generic FlagBase instantiations they alias, or a project's own *Flag
// type — declared in a package whose path ends in "urfave/cli/v3" or
// "urfave/cli", with a type name ending in "Flag". Only keyed struct
// fields are inspected; a Name/EnvVars value that is not a string literal
// (computed at runtime) is left unchecked — FP-safe skip.
//
// Exceptions:
//   - per-line: //nolint:gidcliflags
//   - centralized (GID-239 only): settings.exclude in .golangci.yml — flag
//     names exempt from the required-or-default check.
package cliflags

import (
	"go/ast"
	"go/token"
	"go/types"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/exclude"
)

const (
	ruleNaming   = "GID-238"
	ruleRequired = "GID-239"
)

var (
	kebabCaseRe  = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	upperSnakeRe = regexp.MustCompile(`^[A-Z0-9]+(_[A-Z0-9]+)*$`)
)

// Analyzer — default variant without exclusions.
var Analyzer = NewAnalyzer(Settings{})

// Settings — linter settings from .golangci.yml.
type Settings struct {
	// Exclude — flag names exempt from GID-239 (required-or-default check).
	Exclude []string `json:"exclude"`
}

// NewAnalyzer builds the GID-238/GID-239 analyzer from linter settings (.golangci.yml).
func NewAnalyzer(cfg Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "gidcliflags",
		Doc: ruleNaming + "/" + ruleRequired + ": urfave/cli/v3 flag literals — Name must be " +
			"kebab-case and EnvVars/Sources(cli.EnvVars) must be UPPER_SNAKE_CASE (" + ruleNaming + "); " +
			"a flag must carry Required or a default Value (" + ruleRequired + ")",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, cfg)
		},
	}
}

func run(pass *analysis.Pass, cfg Settings) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			lit, ok := n.(*ast.CompositeLit)
			if !ok || !isCliFlagLit(pass, lit) {
				return true
			}
			checkFlagLit(pass, cfg, lit)
			return true
		})
	}
	return nil, nil
}

// isCliFlagLit reports whether lit's type is a *Flag type from the cli package.
func isCliFlagLit(pass *analysis.Pass, lit *ast.CompositeLit) bool {
	tv, ok := pass.TypesInfo.Types[lit]
	if !ok {
		return false
	}
	obj := flagTypeObj(tv.Type)
	return obj != nil && isCliFlagObj(obj)
}

// flagTypeObj unwraps a pointer and returns the type name object of a named
// or aliased type, or nil. Checking the object at this level (rather than
// resolving through to its underlying type) is what lets a project alias
// such as "type StringFlag = FlagBase[string, StringConfig, stringValue]"
// still be recognized by its own declared name.
func flagTypeObj(t types.Type) *types.TypeName {
	switch tt := t.(type) {
	case *types.Pointer:
		return flagTypeObj(tt.Elem())
	case *types.Named:
		return tt.Obj()
	case *types.Alias:
		return tt.Obj()
	default:
		return nil
	}
}

// isCliFlagObj reports whether obj is a *Flag type declared in a package
// path ending in "urfave/cli/v3" or "urfave/cli".
func isCliFlagObj(obj *types.TypeName) bool {
	if !strings.HasSuffix(obj.Name(), "Flag") {
		return false
	}
	pkg := obj.Pkg()
	return pkg != nil && isCliPackagePath(pkg.Path())
}

func isCliPackagePath(path string) bool {
	return strings.HasSuffix(path, "urfave/cli/v3") || strings.HasSuffix(path, "urfave/cli")
}

// checkFlagLit runs GID-238 (naming) and GID-239 (required-or-default) on a
// single flag composite literal.
func checkFlagLit(pass *analysis.Pass, cfg Settings, lit *ast.CompositeLit) {
	var (
		name            string
		hasLiteralName  bool
		hasRequiredTrue bool
		hasValue        bool
	)
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue // unkeyed field — not the wiring style in practice, skip (FP-safe)
		}
		key, ok := kv.Key.(*ast.Ident)
		if !ok {
			continue
		}
		switch key.Name {
		case "Name":
			if s, ok := stringLit(kv.Value); ok {
				name, hasLiteralName = s, true
				checkKebabCase(pass, kv.Value, s)
			}
		case "EnvVars":
			checkEnvVarsSlice(pass, kv.Value)
		case "Sources":
			checkSourcesEnvVars(pass, kv.Value)
		case "Required":
			hasRequiredTrue = isTrueIdent(kv.Value)
		case "Value":
			hasValue = true // any value, including an explicit zero, counts as a default
		}
	}
	if hasRequiredTrue || hasValue {
		return
	}
	flagName := "<flag>"
	if hasLiteralName {
		if exclude.Match(cfg.Exclude, "", name) {
			return
		}
		flagName = name
	}
	pass.Reportf(lit.Pos(),
		"%s: flag %q has neither Required nor a default Value — a flag consumed by wiring must be required or carry a default",
		ruleRequired, flagName)
}

func checkKebabCase(pass *analysis.Pass, value ast.Expr, name string) {
	if !kebabCaseRe.MatchString(name) {
		pass.Reportf(value.Pos(), "%s: cli flag name %q must be kebab-case", ruleNaming, name)
	}
}

// checkEnvVarsSlice checks a v2-style "EnvVars: []string{...}" field.
func checkEnvVarsSlice(pass *analysis.Pass, value ast.Expr) {
	lit, ok := value.(*ast.CompositeLit)
	if !ok {
		return // not a slice literal (e.g. a variable) — FP-safe skip
	}
	for _, elt := range lit.Elts {
		if s, ok := stringLit(elt); ok {
			checkUpperSnake(pass, elt, s)
		}
	}
}

// checkSourcesEnvVars checks a v3-style "Sources: cli.EnvVars(...)" field.
func checkSourcesEnvVars(pass *analysis.Pass, value ast.Expr) {
	call, ok := value.(*ast.CallExpr)
	if !ok || !isCliEnvVarsCall(pass, call) {
		return // not a cli.EnvVars(...) call — FP-safe skip
	}
	for _, arg := range call.Args {
		if s, ok := stringLit(arg); ok {
			checkUpperSnake(pass, arg, s)
		}
	}
}

// isCliEnvVarsCall reports whether call is cli.EnvVars(...) from the cli package.
func isCliEnvVarsCall(pass *analysis.Pass, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	fn, ok := pass.TypesInfo.Uses[sel.Sel].(*types.Func)
	if !ok || fn.Name() != "EnvVars" {
		return false
	}
	pkg := fn.Pkg()
	return pkg != nil && isCliPackagePath(pkg.Path())
}

func checkUpperSnake(pass *analysis.Pass, node ast.Expr, name string) {
	if !upperSnakeRe.MatchString(name) {
		pass.Reportf(node.Pos(), "%s: env var %q must be UPPER_SNAKE_CASE", ruleNaming, name)
	}
}

func isTrueIdent(expr ast.Expr) bool {
	ident, ok := expr.(*ast.Ident)
	return ok && ident.Name == "true"
}

func stringLit(expr ast.Expr) (string, bool) {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}
	s, err := strconv.Unquote(lit.Value)
	if err != nil {
		return "", false
	}
	return s, true
}
