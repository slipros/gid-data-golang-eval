// Package logconstruct реализует правило GID-154: если сущность содержит
// logger (logrus), её конструктор обязан вызвать WithField(<entity>, <name>) —
// так по логам всегда видно, какая сущность их пишет.
package logconstruct

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/lgr"
)

const ruleID = "GID-154"

// Analyzer — правило GID-154: an entity constructor with a logger must call WithField. Fix: call logger.WithField(<entity>, <name>).
var Analyzer = &analysis.Analyzer{
	Name: "gidlogconstruct",
	Doc:  ruleID + ": an entity constructor with a logger must call WithField. Fix: call logger.WithField(<entity>, <name>)",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	withLogger := structsWithLogger(pass)
	if len(withLogger) == 0 {
		return nil, nil
	}
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || fn.Body == nil {
				continue
			}
			entity, ok := constructedEntity(fn.Name.Name, withLogger)
			if !ok {
				continue
			}
			if !callsWithField(pass, fn.Body) {
				pass.Reportf(fn.Name.Pos(),
					"%s: entity %q has a logger. Fix: constructor %q must call logger.WithField(<entity>, <name>)",
					ruleID, entity, fn.Name.Name)
			}
		}
	}
	return nil, nil
}

// structsWithLogger собирает имена структур пакета, содержащих logrus-поле.
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

// constructedEntity сопоставляет конструктор New<Entity> сущности с logger.
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

// callsWithField ищет в теле вызов WithField на logrus-типе.
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
		if !ok || sel.Sel.Name != "WithField" {
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
