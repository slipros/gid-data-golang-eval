// Package mappedfields implements rules GID-266 and GID-275: in a BFF, a
// gRPC client call carries a MappedFields option, and a mapper does not carry
// a redundant identity field mapping.
//
// A BFF validates the request it received from the frontend and then forwards
// it to the service that owns the data. That service validates it again, in its
// own vocabulary, and answers InvalidArgument with a ValidationError whose
// fields are named after ITS contract (genproto udmpapis/type/error). Handed
// over untouched, that error reaches the browser naming fields the frontend
// never sent — the form has nothing to highlight, and the message points at a
// contract nobody outside the backend knows.
//
// The library closes this with a per-call option:
//
//	resp, err := c.client.CreateSegment(ctx, req,
//	    gdgrpcerror.WithMappedFields(gdmapper.MappedFields{
//	        gdmapper.NewMappedFieldStringEqualWithWholePart("segment_id", "segmentId"),
//	    }))
//
// MappedFieldsUnaryClientInterceptor picks the option out of the call options,
// and on a ValidationError renames the fields before the error travels further.
// Without the option the interceptor forwards the error as it is — and nothing
// about the call looks broken until somebody submits a form and reads the
// answer.
//
// Two diagnostics:
//
//   - the call has no MappedFields option at all;
//   - the option is there but the mapping is empty (`WithMappedFields(nil)`,
//     `WithMappedFields(gdmapper.MappedFields{})`) — the interceptor returns
//     early on `len(o.MappedFields) == 0`, so an empty mapping is the same as no
//     option, spelled in a way that looks handled (the shape GID-265 reports for
//     nil router metrics).
//
// Scope — a BFF module only: one laid out as a service (modlayout.IsServiceModule)
// that owns no data layer (no /dal, no /repository). That is the module whose
// business logic IS calling other services over gRPC and shaping the answer for
// the frontend — lk-api, the same shape GID-160 was taught to leave alone. A
// service with a data layer reaches other services through a repository, and a
// library module has no frontend to answer to. Within the BFF the judged layers
// are /domain/service and /domain/usecase, mirroring GID-160.
//
// What is judged is the CALL, recognised by its signature: a variadic trailing
// parameter of type grpc.CallOption. That covers both a generated client and
// the consumer-side interface a BFF declares next to its user (GID-134), which
// mirrors it — the marker GID-176 uses for the same purpose. Two deliberate
// blind spots follow from it: a call spreading a prepared slice
// (`c.client.Foo(ctx, req, opts...)`) is left alone, because the option may
// well be inside it, and an interface that drops the call options from its
// method signature cannot be judged at all — there is nowhere to pass the
// option, so the rule has nothing to ask for. Trimming the options out of a
// client interface therefore hides the call from this rule; that is a review
// matter, not a linter one.
//
// A call that sends no request data is not judged: an RPC taking nothing beside
// the context (a Ping), a request passed as nil, or an empty literal
// (`client.Statuses(ctx, nil)`, `client.AvailableDatesRange(ctx,
// &rpc.AvailableDatesRangeRequest{})` — both live in lk-api). The callee has no
// field to reject, so no ValidationError can name one. Judging those would
// demand the impossible: the mapping would have to be empty, which is exactly
// what the second diagnostic reports. A request built in a variable is another
// matter — what it holds is not followed, and the call stays judged.
//
// The option is recognised by the name of its type (a named type whose name
// contains "MappedFields" — the library's MappedFieldsInterceptorCallOption),
// so a value held in a variable counts as well as a fresh WithMappedFields
// call, and a project wrapping the constructor in a helper of its own is
// recognised by the helper's name carrying the same marker.
//
// GID-275 also rejects a gdmapper.NewMappedFieldStringEqual*WithWholePart
// call when both field paths are string literals and normalize to the same
// lowerCamel path. It applies wherever the mapper is used. Underscores are
// retained during normalization: a mapping such as page_size -> pageSize is
// meaningful and is not an identity mapping.
//
// A _test.go file is not judged: a double of a client interface repeats its
// signatures, and a test calling it passes the arguments its assertion needs,
// not the options production code owes the frontend.
//
// Exceptions: //nolint:gidmappedfields, or settings.exclude for an RPC whose
// errors need no mapping ("Method" | "Client.Method" — the CALLED method, e.g.
// "Ping" or "HealthClient.Check").
package mappedfields

import (
	"go/ast"
	"go/token"
	"go/types"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"

	"github.com/slipros/gid-data-golang-eval/internal/astwalk"
	"github.com/slipros/gid-data-golang-eval/internal/exclude"
	"github.com/slipros/gid-data-golang-eval/internal/modlayout"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
	"github.com/slipros/gid-data-golang-eval/internal/srcfile"
)

const (
	ruleID         = "GID-266"
	identityRuleID = "GID-275"
	// optionMarker — the substring identifying the call option carrying the
	// mapping: gdgrpcerror.MappedFieldsInterceptorCallOption, built by
	// gdgrpcerror.WithMappedFields.
	optionMarker = "MappedFields"
	// fixExample — the canonical option, used in both diagnostics.
	fixExample = `gdgrpcerror.WithMappedFields(gdmapper.MappedFields{` +
		`gdmapper.NewMappedFieldStringEqualWithWholePart("segment_id", "segmentId")})`
)

// scopes — the layers judged inside a BFF module, the same pair GID-160 names.
var scopes = [][]string{
	{"domain", "service"},
	{"domain", "usecase"},
}

// Analyzer — variant without exclusions.
var Analyzer = NewAnalyzer(Settings{})

// Settings — linter settings from .golangci.yml.
type Settings struct {
	// Exclude — RPCs that need no mapping, as "Method" or "Client.Method".
	// The name matched is the one being CALLED, not the caller.
	Exclude []string `json:"exclude"`
}

// excluded reports whether the RPC being called is on the exclusion list.
func (s Settings) excluded(pass *analysis.Pass, call *ast.CallExpr) bool {
	if len(s.Exclude) == 0 {
		return false
	}

	sel, isSel := call.Fun.(*ast.SelectorExpr)
	if !isSel {
		return false
	}

	var recv string
	if selection := pass.TypesInfo.Selections[sel]; selection != nil {
		if named := namedType(selection.Recv()); named != nil {
			obj := named.Obj()
			recv = obj.Name()
		}
	}

	return exclude.Match(s.Exclude, recv, sel.Sel.Name)
}

// NewAnalyzer builds the GID-266/GID-275 analyzer from the linter settings (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "gidmappedfields",
		Doc: ruleID + ": a gRPC client call in a BFF carries a field mapping, so a validation error of the " +
			"callee reaches the frontend in the vocabulary of the request it sent. Fix: pass " + fixExample +
			" among the call options; " + identityRuleID + ": a literal mapper field path does not map to " +
			"the same lowerCamel path. Fix: remove the redundant NewMappedField call",
		Requires: astwalk.Requires,
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, s)
		},
	}
}

func run(pass *analysis.Pass, s Settings) (any, error) {
	judgeGRPCCalls := inScope(pass.Pkg.Path()) && isBFF(pass)

	skip := func(file *ast.File) bool {
		return ast.IsGenerated(file) || srcfile.IsTest(pass, file)
	}

	astwalk.NodesOf(pass, skip, func(_ *ast.File, call *ast.CallExpr) {
		if judgeGRPCCalls {
			check(pass, call, s)
		}
		checkIdentityMappedField(pass, call)
	})

	return nil, nil
}

// check judges one call: a gRPC client call must carry a non-empty MappedFields
// option among its call options.
func check(pass *analysis.Pass, call *ast.CallExpr, s Settings) {
	sig, isFunc := pass.TypesInfo.TypeOf(call.Fun).(*types.Signature)
	if !isFunc || !takesCallOptions(sig) || s.excluded(pass, call) {
		return
	}

	if nothingToMap(pass, call, sig) {
		return
	}

	opts, judged := callOptions(call, sig)
	if !judged {
		// The caller spread a prepared slice: the option may be inside it.
		return
	}

	option := mappedFieldsOption(pass, opts)
	if option == nil {
		pass.Reportf(call.Pos(),
			"%s: the gRPC call %s carries no MappedFields option, so a validation error of the callee reaches "+
				"the frontend naming the fields of the foreign contract. Fix: %s "+
				"(exceptions: nolint or settings.exclude)",
			ruleID, calleeText(call), fixExample)

		return
	}

	if emptyMapping(pass, option) {
		pass.Reportf(option.Pos(),
			"%s: the MappedFields option of the gRPC call %s is empty, so the interceptor forwards the "+
				"validation error untouched — the same outcome as passing no option. Fix: %s",
			ruleID, calleeText(call), fixExample)
	}
}

// checkIdentityMappedField reports a redundant mapper constructor. The
// constructor is checked independently of the gRPC call that may consume it,
// so package-level mapping declarations are covered as well.
func checkIdentityMappedField(pass *analysis.Pass, call *ast.CallExpr) {
	fn, from, to, normalized, ok := identityMappedField(pass, call)
	if !ok {
		return
	}

	pass.Reportf(call.Pos(),
		"%s: mapped field path %q -> %q is an identity mapping after lowerCamel normalization (%q), "+
			"so the constructor is redundant. Fix: remove gdmapper.%s(%s, %s).",
		identityRuleID, from, to, normalized, fn.Name(), strconv.Quote(from), strconv.Quote(to))
}

// identityMappedField extracts the two field paths of a mapper constructor
// when both are string literals and normalize to the same path.
func identityMappedField(
	pass *analysis.Pass,
	call *ast.CallExpr,
) (fn *types.Func, from, to, normalized string, ok bool) {
	fn = typeutil.StaticCallee(pass.TypesInfo, call)
	if !isMapperFieldEquality(fn) || len(call.Args) < 2 {
		return nil, "", "", "", false
	}

	from, ok = stringLiteral(call.Args[0])
	if !ok {
		return nil, "", "", "", false
	}

	to, ok = stringLiteral(call.Args[1])
	if !ok {
		return nil, "", "", "", false
	}

	normalizedFrom := normalizeFieldPath(from)
	normalizedTo := normalizeFieldPath(to)
	if normalizedFrom != normalizedTo {
		return nil, "", "", "", false
	}

	return fn, from, to, normalizedFrom, true
}

// isMapperFieldEquality recognises the mapper constructors whose first two
// arguments are complete field paths. Prefix, suffix, contains and regexp
// constructors use a pattern as one argument, so equal-looking literals do
// not prove that their mapping is redundant.
func isMapperFieldEquality(fn *types.Func) bool {
	if fn == nil {
		return false
	}

	pkg := fn.Pkg()
	if pkg == nil || pkg.Name() != "mapper" {
		return false
	}

	name := fn.Name()

	return strings.HasPrefix(name, "NewMappedFieldStringEqual") && strings.HasSuffix(name, "WithWholePart")
}

// stringLiteral returns a decoded string literal. Values computed in a
// variable or expression are deliberately left alone because their paths are
// not statically known.
func stringLiteral(expr ast.Expr) (string, bool) {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return "", false
	}

	value, err := strconv.Unquote(lit.Value)
	if err != nil {
		return "", false
	}

	return value, true
}

// normalizeFieldPath canonicalizes the spelling of each path segment to
// lowerCamel without converting underscores. Keeping snake_case distinct from
// camelCase is important: page_size -> pageSize is a real cross-contract map.
func normalizeFieldPath(path string) string {
	parts := strings.Split(path, ".")
	for i, part := range parts {
		parts[i] = normalizeLowerCamelPart(part)
	}

	return strings.Join(parts, ".")
}

func normalizeLowerCamelPart(part string) string {
	words := strings.Split(part, "_")
	for i, word := range words {
		words[i] = normalizeLowerCamelWord(word)
	}

	return strings.Join(words, "_")
}

func normalizeLowerCamelWord(word string) string {
	runes := []rune(word)
	if len(runes) == 0 || !unicode.IsUpper(runes[0]) {
		return word
	}

	upperEnd := 1
	for upperEnd < len(runes) && unicode.IsUpper(runes[upperEnd]) {
		upperEnd++
	}

	if upperEnd < len(runes) && upperEnd > 1 && unicode.IsLower(runes[upperEnd]) {
		upperEnd--
	}

	for i := 0; i < upperEnd; i++ {
		runes[i] = unicode.ToLower(runes[i])
	}

	return string(runes)
}

// isBFF reports whether the module under analysis is a BFF: laid out as a
// service (a composition root of its own) and owning no data layer. A module
// with a /dal reaches other services through a repository, and a library module
// answers to no frontend.
func isBFF(pass *analysis.Pass) bool {
	return modlayout.IsServiceModule(pass) && !modlayout.HasDataLayer(pass)
}

func inScope(pkgPath string) bool {
	for _, scope := range scopes {
		if pathseg.HasLayer(pkgPath, scope...) {
			return true
		}
	}

	return false
}

// nothingToMap reports whether the call sends no request data at all: the RPC
// takes nothing beside the context, or was handed nil / an empty literal. The
// callee has no field to complain about, so it cannot answer with a
// ValidationError naming one, and there is nothing to rename. Asking for the
// option here would demand the impossible: the mapping would have to be empty,
// and an empty mapping is what the second diagnostic reports. Measured on
// lk-api: three calls of 68, among them Statuses(ctx, nil) and
// AvailableDatesRange(ctx, &rpc.AvailableDatesRangeRequest{}).
func nothingToMap(pass *analysis.Pass, call *ast.CallExpr, sig *types.Signature) bool {
	params := sig.Params()

	// The request is the last parameter before the variadic options.
	request := params.Len() - 1
	if request == 0 {
		return true // options only: no request at all
	}

	param := params.At(request - 1)
	if isContext(param.Type()) {
		return true // the RPC takes nothing beside the context, as a Ping does
	}

	if len(call.Args) < request {
		return false // fewer arguments than parameters: not our call to judge
	}

	return emptyRequest(pass, call.Args[request-1])
}

// emptyRequest reports whether the request argument carries no field: nil, or a
// literal with nothing in it — `&rpc.StatusesRequest{}` as well as the bare
// value. A request built in a variable is not followed: what it holds is a
// guess, so the call stays judged.
func emptyRequest(pass *analysis.Pass, arg ast.Expr) bool {
	if unary, isUnary := arg.(*ast.UnaryExpr); isUnary && unary.Op == token.AND {
		arg = unary.X
	}

	return isEmptyValue(pass, arg)
}

// isContext reports whether t is context.Context.
func isContext(t types.Type) bool {
	const (
		contextPkg  = "context"
		contextType = "Context"
	)

	named := namedType(t)
	if named == nil {
		return false
	}

	obj := named.Obj()
	pkg := obj.Pkg()

	return obj.Name() == contextType && pkg != nil && pkg.Path() == contextPkg
}

// takesCallOptions reports whether the signature ends in a variadic
// grpc.CallOption — the marker of a gRPC client call, be it a generated client
// or the consumer-side interface mirroring it (GID-134).
func takesCallOptions(sig *types.Signature) bool {
	// The type of the trailing variadic parameter that marks a gRPC client call.
	const (
		grpcPkg    = "google.golang.org/grpc"
		callOption = "CallOption"
	)

	if !sig.Variadic() {
		return false
	}

	params := sig.Params()

	last := params.At(params.Len() - 1)

	slice, isSlice := types.Unalias(last.Type()).(*types.Slice)
	if !isSlice {
		return false
	}

	named := namedType(slice.Elem())
	if named == nil {
		return false
	}

	obj := named.Obj()
	pkg := obj.Pkg()

	return obj.Name() == callOption && pkg != nil && pkg.Path() == grpcPkg
}

// callOptions returns the arguments that landed in the variadic call-option
// parameter. The second result is false when the call cannot be judged at all —
// the caller spread a slice, so the individual options are not in this call.
func callOptions(call *ast.CallExpr, sig *types.Signature) ([]ast.Expr, bool) {
	if call.Ellipsis.IsValid() {
		return nil, false
	}

	params := sig.Params()

	fixed := params.Len() - 1
	if len(call.Args) <= fixed {
		return nil, true // no options passed at all
	}

	return call.Args[fixed:], true
}

// mappedFieldsOption returns the option carrying the field mapping, or nil.
func mappedFieldsOption(pass *analysis.Pass, opts []ast.Expr) ast.Expr {
	for _, opt := range opts {
		if isMappedFieldsOption(pass, opt) {
			return opt
		}
	}

	return nil
}

// isMappedFieldsOption recognises the option by the name of its type — the
// library's MappedFieldsInterceptorCallOption, whether built on the spot or
// held in a variable. A project helper returning a bare grpc.CallOption hides
// the type, so the name of the callee is checked for the same marker.
func isMappedFieldsOption(pass *analysis.Pass, opt ast.Expr) bool {
	if named := namedType(pass.TypesInfo.TypeOf(opt)); named != nil && hasMarker(named) {
		return true
	}

	call, isCall := opt.(*ast.CallExpr)
	if !isCall {
		return false
	}

	fn := typeutil.StaticCallee(pass.TypesInfo, call)

	return fn != nil && strings.Contains(fn.Name(), optionMarker)
}

// emptyMapping reports whether the option was handed an empty mapping: the
// interceptor returns early on len(MappedFields) == 0, so the option is there
// and does nothing.
func emptyMapping(pass *analysis.Pass, option ast.Expr) bool {
	switch node := option.(type) {
	case *ast.CallExpr:
		if node.Ellipsis.IsValid() {
			return false // a spread slice: its contents are not in this call
		}
		if len(node.Args) == 0 {
			return true
		}

		return len(node.Args) == 1 && isEmptyValue(pass, node.Args[0])
	case *ast.CompositeLit:
		return emptyMappingField(pass, node)
	default:
		return false
	}
}

// emptyMappingField judges the option written as a literal
// (gdgrpcerror.MappedFieldsInterceptorCallOption{...}): the mapping field is either
// missing or empty.
func emptyMappingField(pass *analysis.Pass, lit *ast.CompositeLit) bool {
	for _, elt := range lit.Elts {
		kv, isKV := elt.(*ast.KeyValueExpr)
		if !isKV {
			continue
		}

		key, isIdent := kv.Key.(*ast.Ident)
		if !isIdent || key.Name != optionMarker {
			continue
		}

		return isEmptyValue(pass, kv.Value)
	}

	return true // the field is not set at all
}

// isEmptyValue reports whether the expression is nil or an empty literal.
func isEmptyValue(pass *analysis.Pass, expr ast.Expr) bool {
	switch node := expr.(type) {
	case *ast.Ident:
		_, isNil := pass.TypesInfo.Uses[node].(*types.Nil)

		return isNil
	case *ast.CompositeLit:
		return len(node.Elts) == 0
	default:
		return false
	}
}

// calleeText renders the call for the diagnostic: the method name as written.
func calleeText(call *ast.CallExpr) string {
	switch fun := call.Fun.(type) {
	case *ast.SelectorExpr:
		return fun.Sel.Name
	case *ast.Ident:
		return fun.Name
	default:
		return "of the gRPC client"
	}
}

// hasMarker reports whether the name of the type carries the MappedFields
// marker — the fingerprint of the call option holding the mapping.
func hasMarker(named *types.Named) bool {
	obj := named.Obj()

	return strings.Contains(obj.Name(), optionMarker)
}

// namedType returns the named type behind t — through a pointer — or nil.
func namedType(t types.Type) *types.Named {
	if t == nil {
		return nil
	}

	if ptr, isPtr := types.Unalias(t).(*types.Pointer); isPtr {
		t = ptr.Elem()
	}

	named, isNamed := types.Unalias(t).(*types.Named)
	if !isNamed {
		return nil
	}

	return named
}
