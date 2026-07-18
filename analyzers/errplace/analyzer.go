// Package errplace implements the rules for placing errors by layer:
//
//   - GID-144 (giddomainerrors): all domain errors live in /domain/model.
//     service and usecase neither declare nor create errors — they exchange
//     received errors for errors from model.
//   - GID-145 (giddalerrors): all dal errors live in /dal/entity.
//     The repository exchanges connection errors for errors from entity
//     (if no exchange happened — it passes the original through, which is not creation).
//   - GID-169 (giderrfile): layer errors live in a dedicated file.
//     In the root packages /domain/model and /dal/entity, package-level
//     variables of type error are declared only in a file from settings.files
//     (default: error.go). Refines GID-144/GID-145: those define
//     the "home" layer for errors, while GID-169 picks the exact file inside it.
//
// Forbidden outside the allowed package: declaring package-level variables
// of type error and calling error constructors (errors.New, fmt.Errorf,
// errors.Errorf). Exchange and enrichment — errors.Wrap/WithStack/WithMessage
// (github.com/pkg/errors) and typed gderror errors — are allowed.
package errplace

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

// forbiddenCtors — error constructors: package -> function names.
var forbiddenCtors = map[string]map[string]struct{}{
	"errors":                {"New": {}},
	"fmt":                   {"Errorf": {}},
	"github.com/pkg/errors": {"New": {}, "Errorf": {}},
}

// DomainAnalyzer — rule GID-144: all domain errors live in /domain/model. Fix: declare them in /domain/model.
var DomainAnalyzer = &analysis.Analyzer{
	Name: "giddomainerrors",
	Doc:  "GID-144: all domain errors live in /domain/model. Fix: declare them in /domain/model",
	Run: newRun(&config{
		ruleID:  "GID-144",
		tree:    []string{"domain"},
		allowed: []string{"domain", "model"},
		home:    "/domain/model",
	}),
}

// DALAnalyzer — rule GID-145: all dal errors live in /dal/entity. Fix: declare them in /dal/entity.
var DALAnalyzer = &analysis.Analyzer{
	Name: "giddalerrors",
	Doc:  "GID-145: all dal errors live in /dal/entity. Fix: declare them in /dal/entity",
	Run: newRun(&config{
		ruleID:  "GID-145",
		tree:    []string{"dal"},
		allowed: []string{"dal", "entity"},
		home:    "/dal/entity",
	}),
}

type config struct {
	ruleID  string
	tree    []string // layer tree in which the rule applies
	allowed []string // subpath where errors are allowed
	home    string   // human-readable home of errors for the message
}

func newRun(cfg *config) func(*analysis.Pass) (any, error) {
	return func(pass *analysis.Pass) (any, error) {
		pkgPath := pass.Pkg.Path()
		if !pathseg.HasLayer(pkgPath, cfg.tree...) || pathseg.HasLayer(pkgPath, cfg.allowed...) {
			return nil, nil
		}
		for _, file := range pass.Files {
			if ast.IsGenerated(file) {
				continue
			}
			checkErrorVars(pass, cfg, file)
			checkErrorCtors(pass, cfg, file)
		}
		return nil, nil
	}
}

// checkErrorVars looks for package-level variables implementing error.
func checkErrorVars(pass *analysis.Pass, cfg *config, file *ast.File) {
	for _, decl := range file.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok || gd.Tok != token.VAR {
			continue
		}
		for _, spec := range gd.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for _, name := range vs.Names {
				if name.Name == "_" {
					continue
				}
				obj := pass.TypesInfo.Defs[name]
				if obj == nil || !implementsError(obj.Type()) {
					continue
				}
				pass.Reportf(name.Pos(),
					"%s: error %q is declared in %q. Fix: keep this layer's errors in %s",
					cfg.ruleID, name.Name, pass.Pkg.Path(), cfg.home)
			}
		}
	}
}

// checkErrorCtors looks for calls to error constructors.
func checkErrorCtors(pass *analysis.Pass, cfg *config, file *ast.File) {
	ast.Inspect(file, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		fn := typeutil.Callee(pass.TypesInfo, call)
		f, ok := fn.(*types.Func)
		if !ok || f.Pkg() == nil {
			return true
		}
		fPkg := f.Pkg()
		names, ok := forbiddenCtors[fPkg.Path()]
		if !ok {
			return true
		}
		if _, ok := names[f.Name()]; !ok {
			return true
		}
		pass.Reportf(call.Pos(),
			"%s: creating an error via %s.%s is forbidden. Fix: exchange it for an error from %s (Wrap/WithStack are allowed)",
			cfg.ruleID, fPkg.Name(), f.Name(), cfg.home)
		return true
	})
}

func implementsError(t types.Type) bool {
	errObj := types.Universe.Lookup("error")
	errType := errObj.Type()
	errIface, ok := errType.Underlying().(*types.Interface)
	if !ok {
		return false
	}
	return types.Implements(t, errIface)
}
