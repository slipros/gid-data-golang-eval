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
//
// The pair is spelled the same way everywhere:
//
//   - the KEY is the layer the entity lives in — "client" in /client/**,
//     "service" in /domain/service, "usecase", "repository", "event", "job",
//     "handler" in /server/**. It is not a free-form word: "component" says
//     nothing a log reader can filter on, while the layer plus the entity name
//     locates the writer exactly. The layer→key map is settings.keys.
//   - the VALUE is the entity name in lower snake_case ("device_access",
//     "sink"), never CamelCase.
//
// Both are checked only when spelled as literals; a key or a value computed at
// runtime is left alone.
//
// The composition root — package main and the app layer (internal/app,
// anchored to the module root) — is exempt, as it is in GID-104 and GID-214:
// it does not build an entity of its own, it wires the service together and
// hands the logger to the components, and each of those names itself.
package logconstruct

import (
	"go/ast"
	"go/token"
	"regexp"
	"strconv"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/lgr"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-154"

// entityNameRe — the spelling of an entity name: lower snake_case
// ("device_access"), digits allowed after the first letter.
var entityNameRe = regexp.MustCompile(`^[a-z][a-z0-9]*(_[a-z0-9]+)*$`)

// layerKeys — the log key each layer names its entities under. The key is the
// layer itself, so a log line always says both where the writer sits and which
// entity it is.
var layerKeys = []*layerKey{
	{segments: []string{"domain", "service"}, key: "service"},
	{segments: []string{"domain", "usecase"}, key: "usecase"},
	{segments: []string{"dal", "repository"}, key: "repository"},
	{segments: []string{"client"}, key: "client"},
	{segments: []string{"event"}, key: "event"},
	{segments: []string{"job"}, key: "job"},
	{segments: []string{"schedule"}, key: "schedule"},
	{segments: []string{"server"}, key: "handler"},
}

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

// layerKey — one layer and the log key its entities are named under.
type layerKey struct {
	segments []string
	key      string
}

// namingPair — one key/value pair of an enrichment call; valuePos is NoPos
// when the value is computed rather than spelled as a literal.
type namingPair struct {
	key      string
	value    string
	valuePos token.Pos
}

func run(pass *analysis.Pass) (any, error) {
	// composition root: package main or the app layer — wiring hands the logger
	// down to the components, which name themselves; naming the application
	// itself here adds nothing.
	if pass.Pkg.Name() == "main" || pathseg.HasLayer(pass.Pkg.Path(), "app") {
		return nil, nil
	}

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
		checkEntityNaming(pass, fn.Body)
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

// checkEntityNaming reports the naming pairs of the enrichment calls: the key
// must be the layer the entity lives in, the value the entity name in lower
// snake_case. A constructor in a package outside the known layers is only
// checked for the value.
func checkEntityNaming(pass *analysis.Pass, body *ast.BlockStmt) {
	layerKey, known := layerOf(pass.Pkg.Path())
	keyed := false
	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, isSel := call.Fun.(*ast.SelectorExpr)
		if !isSel {
			return true
		}
		if _, isEnrich := enrichMethods[sel.Sel.Name]; !isEnrich || !lgr.IsMethodSel(pass, sel) {
			return true
		}
		for _, pair := range namingPairs(call.Args) {
			if pair.key == layerKey {
				keyed = true
			}
			checkValue(pass, pair)
		}
		return true
	})
	if known && !keyed {
		pass.Reportf(body.Pos(),
			"%s: the entity is named under a key other than its layer. Fix: use the layer as the key — "+
				"logger.With(%q, \"<entity_name>\"); a free-form key (\"component\") is not filterable in the logs",
			ruleID, layerKey)
	}
}

// checkValue reports an entity name that is not lower snake_case.
func checkValue(pass *analysis.Pass, pair *namingPair) {
	if pair.value == "" || pair.valuePos == token.NoPos || entityNameRe.MatchString(pair.value) {
		return
	}
	pass.Reportf(pair.valuePos,
		"%s: the entity name %q is not lower snake_case. Fix: spell it as the log fields do — "+
			"\"device_access\", not \"DeviceAccess\" or \"deviceAccess\"",
		ruleID, pair.value)
}

// namingPairs collects the key/value pairs of an enrichment call: the
// arguments themselves (With("k", "v"), WithField("k", "v")), the arguments of
// an attribute constructor (slog.String("k", "v")) and the entries of a map
// literal (WithFields(logrus.Fields{"k": "v"})).
func namingPairs(args []ast.Expr) []*namingPair {
	var out []*namingPair
	for i, arg := range args {
		switch a := arg.(type) {
		case *ast.BasicLit:
			// With(key, value): the keys sit at the even positions.
			if a.Kind == token.STRING && i%2 == 0 {
				var value ast.Expr
				if i+1 < len(args) {
					value = args[i+1]
				}
				out = append(out, pairOf(a, value))
			}
		case *ast.CallExpr:
			// slog.String("key", "value") and the other attribute constructors.
			if len(a.Args) > 0 {
				if lit, ok := a.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
					var value ast.Expr
					if len(a.Args) > 1 {
						value = a.Args[1]
					}
					out = append(out, pairOf(lit, value))
				}
			}
		case *ast.CompositeLit:
			// logrus.Fields{"key": "value"}
			for _, elt := range a.Elts {
				kv, ok := elt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				if lit, ok := kv.Key.(*ast.BasicLit); ok && lit.Kind == token.STRING {
					out = append(out, pairOf(lit, kv.Value))
				}
			}
		}
	}
	return out
}

func pairOf(key *ast.BasicLit, value ast.Expr) *namingPair {
	pair := &namingPair{valuePos: token.NoPos}
	if k, err := strconv.Unquote(key.Value); err == nil {
		pair.key = k
	}
	lit, ok := value.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return pair
	}
	if v, err := strconv.Unquote(lit.Value); err == nil {
		pair.value = v
		pair.valuePos = lit.Pos()
	}
	return pair
}

// layerOf returns the log key of the layer the package lives in.
func layerOf(pkgPath string) (string, bool) {
	for _, layer := range layerKeys {
		if pathseg.HasLayer(pkgPath, layer.segments...) {
			return layer.key, true
		}
	}
	return "", false
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
