// Package logconstruct implements rule GID-154: a constructor that gets hold
// of a logger must name the entity on it — so the logs always show which
// entity writes them.
//
// Two shapes trigger the requirement, and either one is enough:
//
//   - the constructed entity has a logger field (New<Entity> matched against a
//     struct of the same name);
//   - a New* constructor takes a logger as a parameter — whatever it stores it
//     in (an entity of a differently spelled name, a closure, another
//     constructor), the logger it passes on must already carry the entity name.
//
// The rule is not pinned to one stack: logrus names the entity with
// WithField(<entity>, <name>), slog with With("<entity>", <name>) or
// With(slog.String(...)); either shape satisfies it.
package logconstruct

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/lgr"
)

const ruleID = "GID-154"

// enrichMethods — the calls that attach the entity name to a logger:
// WithField for logrus, With for slog (WithGroup nests a group, which names
// the entity just as well).
var enrichMethods = map[string]struct{}{
	"WithField": {}, "WithFields": {}, "With": {}, "WithGroup": {},
}

// Analyzer — rule GID-154: an entity constructor with a logger must name the entity on it. Fix: call logger.WithField(<entity>, <name>) (logrus) or logger.With("<entity>", <name>) (slog).
var Analyzer = &analysis.Analyzer{
	Name: "gidlogconstruct",
	Doc: ruleID + ": an entity constructor with a logger must name the entity on it. " +
		"Fix: call logger.WithField(<entity>, <name>) (logrus) or logger.With(\"<entity>\", <name>) (slog)",
	Run: run,
}

func run(pass *analysis.Pass) (any, error) {
	withLogger := structsWithLogger(pass)
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || fn.Body == nil {
				continue
			}
			checkConstructor(pass, fn, withLogger)
		}
	}
	return nil, nil
}

// checkConstructor reports a New* constructor that gets hold of a logger —
// through the entity's field or through its own parameter — without naming the
// entity on it.
func checkConstructor(pass *analysis.Pass, fn *ast.FuncDecl, withLogger map[string]struct{}) {
	entity, viaField := constructedEntity(fn.Name.Name, withLogger)
	viaParam := takesLogger(pass, fn.Type)
	if !viaField && !viaParam {
		return
	}
	if callsWithField(pass, fn.Body) {
		return
	}
	if !viaField {
		entity, _ = cutNew(fn.Name.Name)
	}
	pass.Reportf(fn.Name.Pos(),
		"%s: entity %q has a logger. Fix: constructor %q must name the entity on it — "+
			"logger.WithField(<entity>, <name>) (logrus) or logger.With(\"<entity>\", <name>) (slog)",
		ruleID, entity, fn.Name.Name)
}

// takesLogger reports whether the constructor accepts a logger parameter — the
// second shape that triggers the rule: whatever the logger is stored in, it
// must leave the constructor already carrying the entity name.
func takesLogger(pass *analysis.Pass, fnType *ast.FuncType) bool {
	if fnType.Params == nil {
		return false
	}
	for _, param := range fnType.Params.List {
		if lgr.IsType(pass.TypesInfo.TypeOf(param.Type)) {
			return true
		}
	}
	return false
}

// structsWithLogger collects the names of package structs that hold a logger
// field, of either stack.
func structsWithLogger(pass *analysis.Pass) map[string]struct{} {
	out := map[string]struct{}{}
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				st, ok := ts.Type.(*ast.StructType)
				if !ok {
					continue
				}
				for _, field := range st.Fields.List {
					if lgr.IsType(pass.TypesInfo.TypeOf(field.Type)) {
						out[ts.Name.Name] = struct{}{}
						break
					}
				}
			}
		}
	}
	return out
}

// constructedEntity matches a New<Entity> constructor to an entity with a logger.
func constructedEntity(fnName string, withLogger map[string]struct{}) (string, bool) {
	entity, ok := cutNew(fnName)
	if !ok {
		return "", false
	}
	if _, ok := withLogger[entity]; !ok {
		return "", false
	}
	return entity, true
}

func cutNew(name string) (string, bool) {
	if len(name) <= 3 || name[:3] != "New" {
		return "", false
	}
	return name[3:], true
}

// callsWithField looks in the body for a call that names the entity on a
// logger of either stack.
func callsWithField(pass *analysis.Pass, body *ast.BlockStmt) bool {
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		if found {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		if _, isEnrich := enrichMethods[sel.Sel.Name]; !isEnrich {
			return true
		}
		if lgr.IsMethodSel(pass, sel) {
			found = true
			return false
		}
		return true
	})
	return found
}
