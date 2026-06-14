// Package metricstruct implements rule GID-174 (gidmetricstruct):
// the service metrics package is standardized.
//
// Convention for all backend-go services: the package path is /metric, the
// package name is metric, and metrics are aggregated in an exported Prometheus
// struct aggregator (fields are metrics per protocol/subsystem: HTTP, GRPC,
// Kafka, …) with a Register method.
//
// The rule applies only to a package whose import path ends with the segment
// metric or metrics (pathseg.EndsWith). Other packages are left alone.
//
// Checks:
//  1. the path ends with metrics → the package is named incorrectly;
//  2. the path ends with metric, but there is no Prometheus type;
//  3. Prometheus exists but has no Register method;
//  4. Prometheus is declared but is not a struct.
//
// Grouping convention (extra checks, only for .../metric paths):
//   - additional metrics live in separate files, grouped functionally into
//     structs (one group per file);
//   - prometheus.go does the wiring: the Prometheus type is declared exactly
//     in prometheus.go, and its Register method registers the groups by
//     calling their Register method.
//
// Extra checks:
//  5. the Prometheus type is declared outside prometheus.go;
//  6. prometheus.go declares other exported struct types;
//  7. a file of the metric package (other than prometheus.go) declares ≥2
//     exported struct types — report on the second and subsequent ones;
//  8. a Prometheus field whose type has a Register method is not registered
//     in the body of Prometheus.Register (no <field>.Register(...) call).
package metricstruct

import (
	"go/ast"
	"go/types"
	"path/filepath"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const (
	ruleID     = "GID-174"
	typeName   = "Prometheus"
	regMethod  = "Register"
	wiringFile = "prometheus.go"
)

// Analyzer — rule GID-174. Requires type information: the Register method is
// detected via types (accounting for value/pointer receivers).
var Analyzer = &analysis.Analyzer{
	Name: "gidmetricstruct",
	Doc:  ruleID + ": the metrics package is standardized: path/name metric, a Prometheus struct with a Register method. Fix: follow that layout",
	Run:  run,
}

func run(pass *analysis.Pass) (any, error) {
	path := pass.Pkg.Path()

	// Not applicable: the package is not the root of a metric/metrics layer.
	switch {
	case pathseg.EndsWith(path, "metrics"):
		// Check 1: the package is named metrics instead of metric.
		reportOnPackageClause(pass,
			"%s: the metrics package must be named metric, not metrics. Fix: rename it to metric", ruleID)
		return nil, nil
	case pathseg.EndsWith(path, "metric"):
		// continue with checks 2-4
	default:
		return nil, nil
	}

	ts, named := findPrometheus(pass)
	if named == nil {
		// Check 2: there is no Prometheus type.
		reportOnPackageClause(pass,
			"%s: the metric package must declare a metrics aggregator: struct %s with a %s method. Fix: add it",
			ruleID, typeName, regMethod)
		return nil, nil
	}

	// Check 4: Prometheus exists but is not a struct.
	st, ok := named.Underlying().(*types.Struct)
	if !ok {
		pass.Reportf(ts.Name.Pos(),
			"%s: %s must be a metrics aggregator struct. Fix: make it a struct", ruleID, typeName)
		return nil, nil
	}

	// Check 5: Prometheus is declared outside prometheus.go.
	if filepath.Base(pass.Fset.Position(ts.Name.Pos()).Filename) != wiringFile {
		pass.Reportf(ts.Name.Pos(),
			"%s: the %s aggregator must live in %s. Fix: move it there", ruleID, typeName, wiringFile)
	}

	// Checks 6 and 7: grouping of struct types across files.
	checkGrouping(pass)

	// Check 3: struct Prometheus without a Register method.
	if !hasRegisterMethod(named) {
		pass.Reportf(ts.Name.Pos(),
			"%s: struct %s must have a %s method. Fix: add it", ruleID, typeName, regMethod)
		return nil, nil
	}

	// Check 8: every group field is registered in Prometheus.Register.
	checkRegisterWiring(pass, named, st)

	return nil, nil
}

// checkGrouping implements checks 6 and 7:
//   - 6: prometheus.go must not contain other exported struct types
//     (besides Prometheus);
//   - 7: other files of the package must have at most one exported struct
//     group per file (report on the second and subsequent ones).
func checkGrouping(pass *analysis.Pass) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		fname := filepath.Base(pass.Fset.Position(file.Pos()).Filename)
		if isTestFile(fname) {
			continue
		}
		isWiring := fname == wiringFile
		groupsInFile := 0
		for _, ts := range exportedStructTypes(file) {
			if isWiring {
				if ts.Name.Name == typeName {
					continue // the aggregator itself in prometheus.go is fine
				}
				pass.Reportf(ts.Name.Pos(),
					"%s: a metrics group must live in its own file; %s is wiring only. Fix: move the group out",
					ruleID, wiringFile)
				continue
			}
			groupsInFile++
			if groupsInFile >= 2 {
				pass.Reportf(ts.Name.Pos(),
					"%s: one functional metrics group per file. Fix: split groups into separate files", ruleID)
			}
		}
	}
}

// checkRegisterWiring implements check 8: a Prometheus field whose type (or a
// pointer to it) has a Register method must be registered inside the body of
// Prometheus.Register with a <field>.Register(...) call.
func checkRegisterWiring(pass *analysis.Pass, named *types.Named, st *types.Struct) {
	body, recv := registerMethodBody(pass, named)
	if body == nil {
		return // the Register method is declared in another package/has no body — do not report
	}
	called := registeredFields(body, recv)

	// Deterministic order: by struct field index.
	for i := 0; i < st.NumFields(); i++ {
		f := st.Field(i)
		if !fieldTypeHasRegister(f.Type()) {
			continue
		}
		if _, ok := called[f.Name()]; ok {
			continue
		}
		pass.Reportf(f.Pos(),
			"%s: %s.%s registers group %s. Fix: call its %s",
			ruleID, typeName, regMethod, f.Name(), regMethod)
	}
}

// findPrometheus looks up the Prometheus type declaration in the package and its *types.Named.
func findPrometheus(pass *analysis.Pass) (*ast.TypeSpec, *types.Named) {
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range gd.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok || ts.Name.Name != typeName {
					continue
				}
				obj, ok := pass.TypesInfo.Defs[ts.Name].(*types.TypeName)
				if !ok {
					continue
				}
				named, ok := obj.Type().(*types.Named)
				if !ok {
					continue
				}
				return ts, named
			}
		}
	}
	return nil, nil
}

// hasRegisterMethod reports whether the type has a Register method (any
// signature, value or pointer receiver). The pointer method set covers both cases.
func hasRegisterMethod(named *types.Named) bool {
	return lookupRegister(named) != nil
}

// lookupRegister returns the type's Register method (via the pointer method set) or nil.
func lookupRegister(named *types.Named) *types.Func {
	mset := types.NewMethodSet(types.NewPointer(named))
	obj := named.Obj()
	sel := mset.Lookup(obj.Pkg(), regMethod)
	if sel == nil {
		return nil
	}
	fn, ok := sel.Obj().(*types.Func)
	if !ok {
		return nil
	}
	return fn
}

// fieldTypeHasRegister reports whether the field type (or a pointer to it)
// has a Register method. The field type is reduced to *types.Named.
func fieldTypeHasRegister(t types.Type) bool {
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	return hasRegisterMethod(named)
}

// exportedStructTypes returns the TypeSpecs of the file's exported struct
// types in declaration order (deterministic).
func exportedStructTypes(file *ast.File) []*ast.TypeSpec {
	var out []*ast.TypeSpec
	for _, decl := range file.Decls {
		gd, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, spec := range gd.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok || !ts.Name.IsExported() {
				continue
			}
			if _, ok := ts.Type.(*ast.StructType); !ok {
				continue
			}
			out = append(out, ts)
		}
	}
	return out
}

// registerMethodBody finds the AST body of Prometheus's Register method and
// the name of its receiver (to recognize calls like p.Field.Register(...)).
// Returns (nil, "") if the method is not found in the package files or has no body.
func registerMethodBody(pass *analysis.Pass, named *types.Named) (body *ast.BlockStmt, recv string) {
	fn := lookupRegister(named)
	if fn == nil {
		return nil, ""
	}
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Name.Name != regMethod || fd.Recv == nil || fd.Body == nil {
				continue
			}
			obj, ok := pass.TypesInfo.Defs[fd.Name].(*types.Func)
			if !ok || obj != fn {
				continue
			}
			recv := ""
			if len(fd.Recv.List) > 0 && len(fd.Recv.List[0].Names) > 0 {
				recv = fd.Recv.List[0].Names[0].Name
			}
			return fd.Body, recv
		}
	}
	return nil, ""
}

// registeredFields collects the names of fields for which the Register body
// contains a <field>.Register(...) call. Recognized forms: p.Field.Register(...)
// (via the recv receiver) and Field.Register(...) (directly).
func registeredFields(body *ast.BlockStmt, recv string) map[string]struct{} {
	out := map[string]struct{}{}
	ast.Inspect(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		// top-level selector: X.Register
		topSel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || topSel.Sel.Name != regMethod {
			return true
		}
		switch x := topSel.X.(type) {
		case *ast.SelectorExpr: // p.Field.Register(...)
			if id, ok := x.X.(*ast.Ident); ok && (recv == "" || id.Name == recv) {
				out[x.Sel.Name] = struct{}{}
			}
		case *ast.Ident: // Field.Register(...) directly
			out[x.Name] = struct{}{}
		}
		return true
	})
	return out
}

// isTestFile reports whether the file is a _test.go file.
func isTestFile(name string) bool {
	const suffix = "_test.go"
	return len(name) > len(suffix) && name[len(name)-len(suffix):] == suffix
}

// reportOnPackageClause reports on the package clause of the non-generated
// file with the smallest name — deterministic regardless of pass.Files order.
func reportOnPackageClause(pass *analysis.Pass, format string, args ...any) {
	var target *ast.File
	var targetName string
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		tokenFile := pass.Fset.File(file.Pos())
		name := tokenFile.Name()
		if target == nil || name < targetName {
			target, targetName = file, name
		}
	}
	if target == nil {
		return
	}
	pass.Reportf(target.Name.Pos(), format, args...)
}
