// Package protorequired implements rule GID-232 (proto-enum-required):
// validator.NewRequired() must not be used on proto3 enum fields.
//
// A proto3 non-optional enum is a named int32; its zero value
// (*_UNSPECIFIED = 0) is not treated as "empty" by
// github.com/raoptimus/validator.go/v2 because the value implements
// fmt.Stringer and String() returns a non-empty "..._UNSPECIFIED" string —
// so NewRequired() silently passes for the UNSPECIFIED value. The
// styleguide mandates validator.NewInRange(...) with the allowed values
// instead: UNSPECIFIED is not in the list and gets rejected.
//
// Scope: packages with a "validate" path segment (internal/pathseg). The
// analyzer inspects validator.RuleSet composite literals and resolves the
// validated struct type:
//   - from the enclosing Validate(ctx context.Context, req T) method, or
//   - from the enclosing constructor whose result type has such a
//     Validate method (rules built in the constructor, used in Validate).
//
// For every RuleSet key whose rules contain a validator.NewRequired()
// call (possibly chained, e.g. NewRequired().When(...)) the analyzer
// resolves the struct field with that name. A field is a proto3 enum if
// its type is a named type with underlying int32 that has a
// String() string method and an EnumDescriptor/Descriptor method
// (generated protobuf code). Nested rule sets are followed through
// validator.NewNested(...) and validator.NewEach(validator.NewNested(...)).
//
// FP-safety: if the validated struct, the field, or the nested type cannot
// be resolved confidently, the rule set (or key) is skipped. Pointer enum
// fields (proto3 optional) are not flagged: nil is genuinely empty there.
//
// Exceptions:
//   - per-line: //nolint:gidprotorequired
//   - centralized: settings.exclude in .golangci.yml — entries are
//     "Field" (any request type) or "RequestType.Field".
//
// Source: validator.md#proto3-enum.
package protorequired

import (
	"go/ast"
	"go/token"
	"go/types"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/exclude"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-232"

// Analyzer — default variant without exclusions.
var Analyzer = NewAnalyzer(Settings{})

// Settings — linter settings from .golangci.yml.
type Settings struct {
	// Exclude — enum fields exempt from the rule:
	// "Field" (any request type) or "RequestType.Field".
	Exclude []string `json:"exclude"`
}

// NewAnalyzer builds the GID-232 analyzer from linter settings (.golangci.yml).
func NewAnalyzer(cfg Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "gidprotorequired",
		Doc:  ruleID + ": NewRequired on a proto3 enum field treats *_UNSPECIFIED=0 as empty. Fix: validator.NewInRange(pb.Status_ACTIVE, pb.Status_CLOSED)",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, cfg)
		},
	}
}

func run(pass *analysis.Pass, cfg Settings) (any, error) {
	// Scope: only packages of the validate layer.
	if !pathseg.Contains(pass.Pkg.Path(), "validate") {
		return nil, nil
	}
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Body == nil {
				continue
			}
			target := targetStruct(pass, fd)
			if target == nil {
				continue // validated struct not resolvable — skip (FP-safe)
			}
			visited := make(map[*ast.CompositeLit]bool)
			ast.Inspect(fd.Body, func(n ast.Node) bool {
				lit, ok := n.(*ast.CompositeLit)
				if !ok || visited[lit] || !isRuleSet(pass, lit) {
					return true
				}
				checkRuleSet(pass, cfg, lit, target, visited)
				return true
			})
		}
	}
	return nil, nil
}

// targetStruct resolves the struct type validated by the rule sets built in
// fd: either fd is a Validate(ctx, req) method, or fd is a constructor of a
// validator type that has such a method.
func targetStruct(pass *analysis.Pass, fd *ast.FuncDecl) *types.Named {
	fn, ok := pass.TypesInfo.Defs[fd.Name].(*types.Func)
	if !ok {
		return nil
	}
	sig, ok := fn.Type().(*types.Signature)
	if !ok {
		return nil
	}
	if fd.Recv != nil {
		if fd.Name.Name != "Validate" {
			return nil
		}
		return requestStruct(sig)
	}
	// Constructor: a result type declared in this package whose method set
	// contains Validate(ctx, req).
	results := sig.Results()
	for i := 0; i < results.Len(); i++ {
		result := results.At(i)
		named := namedStruct(result.Type())
		if named == nil {
			continue
		}
		obj := named.Obj()
		if obj.Pkg() != pass.Pkg {
			continue
		}
		if t := validateMethodRequest(named); t != nil {
			return t
		}
	}
	return nil
}

// validateMethodRequest returns the request struct of the Validate(ctx, req)
// method of the given validator type, or nil.
func validateMethodRequest(named *types.Named) *types.Named {
	// Look on both T and *T: pointer receivers are not in the method set of T.
	mset := types.NewMethodSet(types.NewPointer(named))
	for i := 0; i < mset.Len(); i++ {
		sel := mset.At(i)
		fn, ok := sel.Obj().(*types.Func)
		if !ok || fn.Name() != "Validate" {
			continue
		}
		sig, ok := fn.Type().(*types.Signature)
		if !ok {
			return nil
		}
		return requestStruct(sig)
	}
	return nil
}

// requestStruct returns the struct type of the second parameter of a
// Validate(ctx context.Context, req T) signature, or nil.
func requestStruct(sig *types.Signature) *types.Named {
	params := sig.Params()
	if params.Len() < 2 {
		return nil
	}
	param := params.At(1)
	return namedStruct(param.Type())
}

// namedStruct unwraps a pointer and returns the named struct type, or nil.
func namedStruct(t types.Type) *types.Named {
	t = types.Unalias(t)
	if ptr, ok := t.(*types.Pointer); ok {
		t = types.Unalias(ptr.Elem())
	}
	named, ok := t.(*types.Named)
	if !ok {
		return nil
	}
	if _, ok := named.Underlying().(*types.Struct); !ok {
		return nil
	}
	return named
}

// isRuleSet reports whether the composite literal has the type
// validator.RuleSet from github.com/raoptimus/validator.go.
func isRuleSet(pass *analysis.Pass, lit *ast.CompositeLit) bool {
	tv, ok := pass.TypesInfo.Types[lit]
	if !ok {
		return false
	}
	named, ok := types.Unalias(tv.Type).(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	pkg := obj.Pkg()
	return obj.Name() == "RuleSet" && pkg != nil && isValidatorLib(pkg.Path())
}

func isValidatorLib(path string) bool {
	const validatorLibPath = "github.com/raoptimus/validator.go"
	return path == validatorLibPath || strings.HasPrefix(path, validatorLibPath+"/")
}

// checkRuleSet checks every "Field": {rules...} entry of a RuleSet literal
// against the fields of target and recurses into nested rule sets.
func checkRuleSet(pass *analysis.Pass, cfg Settings, lit *ast.CompositeLit,
	target *types.Named, visited map[*ast.CompositeLit]bool,
) {
	visited[lit] = true
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		field, ok := stringConst(kv.Key)
		if !ok {
			continue
		}
		rules, ok := kv.Value.(*ast.CompositeLit)
		if !ok {
			continue
		}
		fieldType := fieldType(pass, target, field)
		for _, rule := range rules.Elts {
			checkRule(pass, cfg, rule, target, field, fieldType, visited)
		}
	}
}

// checkRule checks a single rule expression of a RuleSet entry.
func checkRule(pass *analysis.Pass, cfg Settings, rule ast.Expr,
	target *types.Named, field string, fieldType types.Type,
	visited map[*ast.CompositeLit]bool,
) {
	base := baseCall(rule)
	if base == nil {
		return
	}
	switch calleeName(pass, base) {
	case "NewRequired":
		if fieldType == nil {
			return // field not resolvable — skip (FP-safe)
		}
		if _, isPtr := types.Unalias(fieldType).(*types.Pointer); isPtr {
			return // proto3 optional enum: nil pointer is genuinely empty
		}
		if !isProtoEnum(fieldType) {
			return
		}
		targetObj := target.Obj()
		if exclude.Match(cfg.Exclude, targetObj.Name(), field) {
			return
		}
		pass.Reportf(base.Pos(),
			"%s: NewRequired on proto3 enum field %q treats *_UNSPECIFIED=0 as empty. Fix: validator.NewInRange(pb.Status_ACTIVE, pb.Status_CLOSED)",
			ruleID, field)
	case "NewNested":
		nested := ruleSetArg(pass, base)
		if nested == nil {
			return
		}
		visited[nested] = true
		if fieldType == nil {
			return // field not resolvable — skip the nested set (FP-safe)
		}
		if nt := namedStruct(fieldType); nt != nil {
			checkRuleSet(pass, cfg, nested, nt, visited)
		}
	case "NewEach":
		// NewEach(NewNested(RuleSet{...})): nested type is the slice element.
		for _, arg := range base.Args {
			inner := baseCall(arg)
			if inner == nil || calleeName(pass, inner) != "NewNested" {
				continue
			}
			nested := ruleSetArg(pass, inner)
			if nested == nil {
				continue
			}
			visited[nested] = true
			if nt := namedStruct(sliceElem(fieldType)); nt != nil {
				checkRuleSet(pass, cfg, nested, nt, visited)
			}
		}
	}
}

// baseCall unwraps a chained rule expression
// (validator.NewRequired().When(...).SkipOnEmpty()) down to the base call.
func baseCall(expr ast.Expr) *ast.CallExpr {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return nil
	}
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		if prev, ok := sel.X.(*ast.CallExpr); ok {
			return baseCall(prev)
		}
	}
	return call
}

// calleeName returns the name of the called validator.go function
// ("NewRequired", "NewNested", ...) or "" for any other callee.
func calleeName(pass *analysis.Pass, call *ast.CallExpr) string {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	fn, ok := pass.TypesInfo.Uses[sel.Sel].(*types.Func)
	if !ok {
		return ""
	}
	fnPkg := fn.Pkg()
	if fnPkg == nil || !isValidatorLib(fnPkg.Path()) {
		return ""
	}
	return fn.Name()
}

// ruleSetArg returns the RuleSet composite literal passed as the first
// argument of a call (validator.NewNested(validator.RuleSet{...})), or nil.
func ruleSetArg(pass *analysis.Pass, call *ast.CallExpr) *ast.CompositeLit {
	if len(call.Args) == 0 {
		return nil
	}
	lit, ok := call.Args[0].(*ast.CompositeLit)
	if !ok || !isRuleSet(pass, lit) {
		return nil
	}
	return lit
}

// fieldType resolves the type of the named field of the target struct
// (embedded fields included), or nil if it is not a field.
func fieldType(pass *analysis.Pass, target *types.Named, name string) types.Type {
	obj, _, _ := types.LookupFieldOrMethod(target, true, pass.Pkg, name)
	v, ok := obj.(*types.Var)
	if !ok || !v.IsField() {
		return nil
	}
	return v.Type()
}

// isProtoEnum reports whether t is a generated proto3 enum: a named type
// with underlying int32 that has a String() string method and an
// EnumDescriptor/Descriptor method.
func isProtoEnum(t types.Type) bool {
	named, ok := types.Unalias(t).(*types.Named)
	if !ok {
		return false
	}
	basic, ok := named.Underlying().(*types.Basic)
	if !ok || basic.Kind() != types.Int32 {
		return false
	}
	hasStringer, hasDescriptor := false, false
	mset := types.NewMethodSet(named) // proto enums use value receivers
	for i := 0; i < mset.Len(); i++ {
		sel := mset.At(i)
		fn, ok := sel.Obj().(*types.Func)
		if !ok {
			continue
		}
		switch fn.Name() {
		case "String":
			hasStringer = isStringerSig(fn)
		case "EnumDescriptor", "Descriptor":
			hasDescriptor = true
		}
	}
	return hasStringer && hasDescriptor
}

// isStringerSig reports whether fn has the String() string signature.
func isStringerSig(fn *types.Func) bool {
	sig, ok := fn.Type().(*types.Signature)
	if !ok {
		return false
	}
	params := sig.Params()
	results := sig.Results()
	if params.Len() != 0 || results.Len() != 1 {
		return false
	}
	result := results.At(0)
	basic, ok := types.Unalias(result.Type()).(*types.Basic)
	return ok && basic.Kind() == types.String
}

// sliceElem returns the element type of a slice (pointer unwrapped), or nil.
func sliceElem(t types.Type) types.Type {
	if t == nil {
		return nil
	}
	unaliased := types.Unalias(t)
	slice, ok := unaliased.Underlying().(*types.Slice)
	if !ok {
		return nil
	}
	return slice.Elem()
}

// stringConst extracts the string value of a RuleSet key literal.
func stringConst(expr ast.Expr) (string, bool) {
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
