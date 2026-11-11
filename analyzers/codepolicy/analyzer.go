// Package codepolicy checks count-based layout and narrow repository/domain idioms.
package codepolicy

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"regexp"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"

	"github.com/slipros/gid-data-golang-eval/internal/astwalk"
	"github.com/slipros/gid-data-golang-eval/internal/librarypath"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
	"github.com/slipros/gid-data-golang-eval/internal/srcfile"
)

var positional = regexp.MustCompile(`\$[1-9]\d*\b`)
var sqlStart = regexp.MustCompile(`(?i)^\s*(SELECT|INSERT|UPDATE|DELETE|WITH)\b`)

// Analyzer groups the closely related code shape checks GID-281 through GID-297.
var Analyzer = &analysis.Analyzer{
	Name:     "gidcodepolicy",
	Doc:      "GID-281–297: explicit repository/domain idioms and count-based multiline layout. Fix: use named intermediates, batch converters and one item per line.",
	Requires: astwalk.Requires,
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	repository := pathseg.HasLayer(pass.Pkg.Path(), "dal", "repository")
	domain := pathseg.HasLayer(pass.Pkg.Path(), "domain", "service") || pathseg.HasLayer(pass.Pkg.Path(), "domain", "usecase")
	converter := pathseg.HasLayer(pass.Pkg.Path(), "domain", "service", "convert") || pathseg.HasLayer(pass.Pkg.Path(), "domain", "usecase", "convert")
	skip := func(file *ast.File) bool { return ast.IsGenerated(file) || srcfile.IsTest(pass, file) }
	var functions []*functionContext
	astwalk.Nodes(pass, []ast.Node{(*ast.FuncDecl)(nil), (*ast.FuncLit)(nil)}, skip, func(_ *ast.File, n ast.Node) {
		switch fn := n.(type) {
		case *ast.FuncDecl:
			if fn.Body != nil {
				functions = append(functions, &functionContext{fn.Body.Pos(), fn.Body.End(), returnsError(pass, fn.Type), fn.Body})
			}
		case *ast.FuncLit:
			functions = append(functions, &functionContext{fn.Body.Pos(), fn.Body.End(), returnsError(pass, fn.Type), fn.Body})
		}
	})
	astwalk.Nodes(pass, []ast.Node{(*ast.CompositeLit)(nil), (*ast.CallExpr)(nil), (*ast.FuncType)(nil), (*ast.TypeSpec)(nil), (*ast.GenDecl)(nil), (*ast.BlockStmt)(nil), (*ast.IfStmt)(nil), (*ast.RangeStmt)(nil), (*ast.BinaryExpr)(nil), (*ast.SwitchStmt)(nil)}, skip, func(_ *ast.File, n ast.Node) {
		switch n := n.(type) {
		case *ast.CompositeLit:
			checkLiteralLayout(pass, n)
		case *ast.CallExpr:
			checkCallLayout(pass, n)
			checkValidated(pass, n)
			if repository {
				checkNamedArgs(pass, n)
				checkSQLCall(pass, n)
			}
			if domain {
				checkMarshal(pass, n)
			}
		case *ast.FuncType:
			checkSignatureLayout(pass, n)
		case *ast.TypeSpec:
			if repository && n.Assign.IsValid() {
				report(pass, n, "282", "repository type alias hides the entity contract", "use entity.Datamart directly")
			}
		case *ast.GenDecl:
			if repository && n.Tok == token.CONST {
				checkSQL(pass, n)
			}
		case *ast.BlockStmt:
			checkSuccessReturn(pass, n)
			checkFailedResults(pass, n, functions)
			checkSets(pass, n)
			if repository {
				checkQueryAssignments(pass, n)
			}
		case *ast.IfStmt:
			if domain && inErrorFunction(functions, n.Pos()) {
				checkEnumIf(pass, n)
			}
		case *ast.RangeStmt:
			if domain && !converter {
				checkBatch(pass, n)
			}
		case *ast.BinaryExpr:
			checkStatusLiteral(pass, n)
		case *ast.SwitchStmt:
			if domain && inErrorFunction(functions, n.Pos()) {
				checkEnumSwitch(pass, n)
			}
		}
	})
	// Top-level declarations need visual separation too. Stock wsl only visits bodies.
	for _, file := range pass.Files {
		if skip(file) {
			continue
		}
		for i := 1; i < len(file.Decls); i++ {
			if sentinelDecl(pass, file.Decls[i-1]) && sentinelDecl(pass, file.Decls[i]) {
				report(pass, file.Decls[i], "296", "adjacent error variables use separate declarations", "group the error variables in one var block")
			}
			previous, prevOK := file.Decls[i-1].(*ast.FuncDecl)
			next, nextOK := file.Decls[i].(*ast.FuncDecl)
			if prevOK && nextOK && previous.Body != nil && line(pass, next.Pos()) == line(pass, previous.End())+1 {
				report(pass, next, "295", "adjacent functions have no blank line", "insert a blank line between function declarations")
			}
		}
	}
	return nil, nil
}

func report(pass *analysis.Pass, n ast.Node, id, problem, fix string) {
	pass.Reportf(n.Pos(), "GID-%s: %s. Fix: %s.", id, problem, fix)
}

func line(pass *analysis.Pass, pos token.Pos) int { return pass.Fset.Position(pos).Line }

func checkLiteralLayout(pass *analysis.Pass, lit *ast.CompositeLit) {
	t := pass.TypesInfo.TypeOf(lit)
	if t == nil {
		return
	}
	threshold := 0
	switch t.Underlying().(type) {
	case *types.Struct:
		threshold = 5
	case *types.Map:
		threshold = 4
	}
	if threshold == 0 || len(lit.Elts) < threshold {
		return
	}
	checkItems(pass, lit, lit.Lbrace, lit.Rbrace, lit.Elts)
}

func checkItems(pass *analysis.Pass, n ast.Node, open, closing token.Pos, items []ast.Expr) {
	previous := line(pass, open)
	for _, item := range items {
		if line(pass, item.Pos()) <= previous {
			report(pass, n, "281", "large construct packs multiple items on a line", "put each item on its own line with opening and closing delimiters on separate lines")
			return
		}
		previous = line(pass, item.End())
	}
	if line(pass, closing) <= previous {
		report(pass, n, "281", "large construct closes on an item line", "put the closing delimiter on its own line")
	}
}

func checkCallLayout(pass *analysis.Pass, call *ast.CallExpr) {
	if len(call.Args) >= 8 {
		checkItems(pass, call, call.Lparen, call.Rparen, call.Args)
	}
}

func checkSignatureLayout(pass *analysis.Pass, fn *ast.FuncType) {
	if fn.Params == nil {
		return
	}
	count := 0
	for _, f := range fn.Params.List {
		count += max(1, len(f.Names))
	}
	if count < 8 {
		return
	}
	previous := line(pass, fn.Params.Opening)
	for _, f := range fn.Params.List {
		if len(f.Names) > 1 || line(pass, f.Pos()) <= previous {
			report(pass, fn, "281", "large signature packs parameters on a line", "put each parameter on its own line")
			return
		}
		previous = line(pass, f.End())
	}
	if line(pass, fn.Params.Closing) <= previous {
		report(pass, fn, "281", "large signature closes on a parameter line", "put the closing parenthesis on its own line")
	}
}

func checkSQL(pass *analysis.Pass, decl *ast.GenDecl) {
	for _, spec := range decl.Specs {
		value, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		for _, expr := range value.Values {
			tv := pass.TypesInfo.Types[expr]
			if tv.Value == nil || tv.Value.Kind() != constant.String {
				continue
			}
			query := constant.StringVal(tv.Value)
			if !sqlStart.MatchString(query) {
				continue
			}
			// Quoted SQL text, comments and dollar strings cannot be bind placeholders.
			executable := sqlCode(query)
			if positional.MatchString(executable) {
				report(pass, expr, "284", "repository SQL uses positional placeholders", "use @id with gdpostgres.NamedArgs")
			}
			if formatted := strings.TrimSpace(query); !strings.Contains(formatted, "\n") && len(formatted) > 80 {
				report(pass, expr, "283", "long SQL constant is a single line", "use a raw multiline SQL literal with clauses on separate lines")
			}
		}
	}
}

// sqlCode blanks quoted strings, quoted identifiers, comments and dollar strings.
func sqlCode(sql string) string {
	var out strings.Builder
	for i := 0; i < len(sql); {
		if sql[i] == '\'' || sql[i] == '"' {
			quote := sql[i]
			escaped := quote == '\'' && i > 0 && (sql[i-1] == 'E' || sql[i-1] == 'e')
			i++
			for i < len(sql) {
				if escaped && sql[i] == '\\' && i+1 < len(sql) {
					i += 2
					continue
				}
				if sql[i] == quote {
					i++
					if i < len(sql) && sql[i] == quote {
						i++
						continue
					}
					break
				}
				i++
			}
			out.WriteByte(' ')
			continue
		}
		if strings.HasPrefix(sql[i:], "--") {
			for i < len(sql) && sql[i] != '\n' {
				i++
			}
			out.WriteByte(' ')
			continue
		}
		if strings.HasPrefix(sql[i:], "/*") {
			depth := 1
			i += 2
			for i < len(sql) && depth > 0 {
				switch {
				case strings.HasPrefix(sql[i:], "/*"):
					depth++
					i += 2
				case strings.HasPrefix(sql[i:], "*/"):
					depth--
					i += 2
				default:
					i++
				}
			}
			out.WriteByte(' ')
			continue
		}
		if sql[i] == '$' {
			j := i + 1
			for j < len(sql) && (sql[j] == '_' || sql[j] >= 'a' && sql[j] <= 'z' || sql[j] >= 'A' && sql[j] <= 'Z' || j > i+1 && sql[j] >= '0' && sql[j] <= '9') {
				j++
			}
			if j < len(sql) && sql[j] == '$' {
				delimiter := sql[i : j+1]
				if end := strings.Index(sql[j+1:], delimiter); end >= 0 {
					i = j + 1 + end + len(delimiter)
					out.WriteByte(' ')
					continue
				}
			}
		}
		out.WriteByte(sql[i])
		i++
	}
	return out.String()
}

func namedType(t types.Type) *types.Named {
	if t == nil {
		return nil
	}
	if p, ok := t.(*types.Pointer); ok {
		t = p.Elem()
	}
	named, ok := types.Unalias(t).(*types.Named)
	if !ok {
		return nil
	}
	return named
}

func checkNamedArgs(pass *analysis.Pass, call *ast.CallExpr) {
	for _, arg := range call.Args {
		lit, ok := ast.Unparen(arg).(*ast.CompositeLit)
		if !ok {
			continue
		}
		// The corporate package re-exports pgx.NamedArgs through a type alias.
		// Resolve the source type name before unaliasing its runtime type.
		var obj *types.TypeName
		switch typ := ast.Unparen(lit.Type).(type) {
		case *ast.SelectorExpr:
			obj, ok = pass.TypesInfo.ObjectOf(typ.Sel).(*types.TypeName)
		case *ast.Ident:
			obj, ok = pass.TypesInfo.ObjectOf(typ).(*types.TypeName)
		}
		if !ok || obj == nil || obj.Pkg() == nil || obj.Name() != "NamedArgs" {
			continue
		}
		objectPackage := obj.Pkg()
		pkg := librarypath.Canonical(pass, objectPackage.Path())
		if pathseg.SameLibrary(pkg, "gitlab.gid.team/gid-data/tech/golang/libs/postgres.git") ||
			pathseg.SameLibrary(pkg, "gid.team/gid-data/tech/golang/libs/postgres.git") ||
			pathseg.SameLibrary(pkg, "github.com/jackc/pgx") {
			report(pass, arg, "285", "NamedArgs literal is inline in a call", "assign args := gdpostgres.NamedArgs{...} before the call")
		}
	}
}

func checkValidated(pass *analysis.Pass, call *ast.CallExpr) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return
	}
	selection := pass.TypesInfo.Selections[sel]
	if selection == nil || selection.Kind() != types.MethodVal {
		return
	}
	receiver := ast.Unparen(sel.X)
	if unary, ok := receiver.(*ast.UnaryExpr); ok && unary.Op == token.AND {
		receiver = ast.Unparen(unary.X)
	}
	if _, ok := receiver.(*ast.CompositeLit); ok {
		report(pass, call, "287", "method is called on a temporary composite literal", "assign the model to a variable, then call value."+sel.Sel.Name+"()")
	}
}

func checkMarshal(pass *analysis.Pass, call *ast.CallExpr) {
	fn := typeutil.StaticCallee(pass.TypesInfo, call)
	if fn == nil || fn.Pkg() == nil {
		return
	}
	pkg := fn.Pkg()
	if pkg.Path() != "encoding/json" || fn.Name() != "Marshal" || len(call.Args) != 1 {
		return
	}
	t := pass.TypesInfo.TypeOf(call.Args[0])
	if ptr, ok := t.(*types.Pointer); ok {
		t = ptr.Elem()
	}
	if _, ok := types.Unalias(t).(*types.Struct); ok {
		report(pass, call, "292", "domain logic marshals an anonymous struct", "declare a named model with json tags in domain/model and marshal that model")
	}
}

func object(pass *analysis.Pass, expr ast.Expr) types.Object {
	id, ok := ast.Unparen(expr).(*ast.Ident)
	if !ok {
		return nil
	}
	return pass.TypesInfo.ObjectOf(id)
}

func guardedError(pass *analysis.Pass, condition ast.Expr) types.Object {
	binary, ok := ast.Unparen(condition).(*ast.BinaryExpr)
	if !ok || binary.Op != token.NEQ {
		return nil
	}
	x, y := binary.X, binary.Y
	if id, ok := x.(*ast.Ident); ok && id.Name == "nil" {
		x, y = y, x
	}
	id, ok := y.(*ast.Ident)
	if !ok || id.Name != "nil" {
		return nil
	}
	obj := object(pass, x)
	if obj == nil {
		return nil
	}
	errorObject := types.Universe.Lookup("error")
	errorType := errorObject.Type()
	if !types.AssignableTo(obj.Type(), errorType) {
		return nil
	}
	return obj
}

func endsReturn(block *ast.BlockStmt) bool {
	if len(block.List) == 0 {
		return false
	}
	_, ok := block.List[len(block.List)-1].(*ast.ReturnStmt)
	return ok
}

func checkSuccessReturn(pass *analysis.Pass, block *ast.BlockStmt) {
	for i := 1; i < len(block.List); i++ {
		guard, ok := block.List[i-1].(*ast.IfStmt)
		if !ok || guard.Else != nil || !endsReturn(guard.Body) {
			continue
		}
		errObj := guardedError(pass, guard.Cond)
		ret, ok := block.List[i].(*ast.ReturnStmt)
		if !ok || errObj == nil || len(ret.Results) < 2 {
			continue
		}
		if object(pass, ret.Results[len(ret.Results)-1]) == errObj {
			report(pass, ret.Results[len(ret.Results)-1], "288", "success return repeats the error already checked by a terminating guard", "return created, nil")
		}
	}
}

func conversionCall(pass *analysis.Pass, expr ast.Expr) bool {
	call, ok := ast.Unparen(expr).(*ast.CallExpr)
	if !ok {
		return false
	}
	fn := typeutil.StaticCallee(pass.TypesInfo, call)
	if fn == nil || fn.Pkg() == nil {
		return false
	}
	pkg := fn.Pkg()
	return pathseg.EndsWith(pkg.Path(), "convert")
}

func checkFailedResults(pass *analysis.Pass, block *ast.BlockStmt, functions []*functionContext) {
	for i, stmt := range block.List {
		guard, ok := stmt.(*ast.IfStmt)
		if !ok {
			continue
		}
		errObj := guardedError(pass, guard.Cond)
		if errObj == nil {
			continue
		}
		assignment, ok := guard.Init.(*ast.AssignStmt)
		if !ok && i > 0 {
			assignment, ok = block.List[i-1].(*ast.AssignStmt)
		}
		if !ok || len(assignment.Rhs) != 1 || len(assignment.Lhs) < 2 || !conversionCall(pass, assignment.Rhs[0]) {
			continue
		}
		if object(pass, assignment.Lhs[len(assignment.Lhs)-1]) != errObj {
			continue
		}
		results := map[types.Object]bool{}
		for _, lhs := range assignment.Lhs[:len(assignment.Lhs)-1] {
			if obj := object(pass, lhs); obj != nil {
				results[obj] = true
			}
		}
		ast.Inspect(guard.Body, func(n ast.Node) bool {
			if _, ok := n.(*ast.FuncLit); ok {
				return false
			}
			switch n := n.(type) {
			case *ast.SelectorExpr:
				if results[object(pass, n.X)] {
					report(pass, n, "289", "failed conversion result fields are used inside its error branch", "return the conversion error before using the result")
				}
			case *ast.AssignStmt:
				// A replacement is a fresh value, no longer the failed result.
				for _, lhs := range n.Lhs {
					if obj := object(pass, lhs); obj != nil {
						delete(results, obj)
					}
				}
				if endsReturn(guard.Body) || !hasFollowingSuccessReturn(pass, functions, guard, errObj) {
					return true
				}
				for _, rhs := range n.Rhs {
					if object(pass, rhs) == errObj {
						for _, lhs := range n.Lhs {
							if _, ok := lhs.(*ast.SelectorExpr); ok {
								report(pass, n, "294", "conversion error is saved in a field and execution continues", "propagate the conversion error to the caller")
							}
						}
					}
				}
			}
			return true
		})
	}
}

func checkBatch(pass *analysis.Pass, rng *ast.RangeStmt) {
	// Only loops whose observable shape is conversion followed by append are judged.
	converted := map[types.Object]bool{}
	for _, stmt := range rng.Body.List {
		assignment, ok := stmt.(*ast.AssignStmt)
		if !ok || len(assignment.Rhs) != 1 {
			continue
		}
		if conversionCall(pass, assignment.Rhs[0]) {
			for _, lhs := range assignment.Lhs {
				if obj := object(pass, lhs); obj != nil {
					converted[obj] = true
				}
			}
			continue
		}
		call, ok := assignment.Rhs[0].(*ast.CallExpr)
		if !ok {
			continue
		}
		fn, ok := call.Fun.(*ast.Ident)
		if !ok || fn.Name != "append" {
			continue
		}
		if builtin, ok := pass.TypesInfo.ObjectOf(fn).(*types.Builtin); !ok || builtin.Name() != "append" {
			continue
		}
		for _, arg := range call.Args[1:] {
			if conversionCall(pass, arg) || converted[object(pass, arg)] {
				report(pass, rng, "290", "domain loop converts individual values while collecting a slice", "add a batch convert function and call it once with the input slice")
				return
			}
		}
	}
}

func enumOperand(pass *analysis.Pass, expr ast.Expr) bool {
	named := namedType(pass.TypesInfo.TypeOf(expr))
	if named == nil {
		return false
	}
	obj := named.Obj()
	pkg := obj.Pkg()
	if pkg == nil || !pathseg.EndsWith(pkg.Path(), "enum") {
		return false
	}
	_, basic := named.Underlying().(*types.Basic)
	return basic
}

func checkEnumIf(pass *analysis.Pass, guard *ast.IfStmt) {
	if guard.Else == nil {
		return
	}
	binary, ok := ast.Unparen(guard.Cond).(*ast.BinaryExpr)
	if !ok || binary.Op != token.EQL && binary.Op != token.NEQ {
		return
	}
	if enumOperand(pass, binary.X) && enumOperand(pass, binary.Y) {
		report(pass, guard, "291", "enum dispatch uses if/else and silently treats unknown values as a known variant", "use an exhaustive switch with default returning gderr.NewUnhandledValueError(value)")
	}
}

func checkStatusLiteral(pass *analysis.Pass, binary *ast.BinaryExpr) {
	if binary.Op != token.EQL && binary.Op != token.NEQ {
		return
	}
	for _, expr := range []ast.Expr{binary.X, binary.Y} {
		lit, ok := expr.(*ast.BasicLit)
		if !ok || lit.Kind != token.STRING {
			continue
		}
		tv := pass.TypesInfo.Types[lit]
		if tv.Value != nil && constant.StringVal(tv.Value) != "" {
			report(pass, lit, "293", "comparison embeds a nonempty string literal", "declare a named constant such as const cabinetNotVerified = \"cabinet_not_verified\"")
		}
	}
}

func checkSets(pass *analysis.Pass, block *ast.BlockStmt) {
	for _, stmt := range block.List {
		assignment, ok := stmt.(*ast.AssignStmt)
		if !ok || assignment.Tok != token.DEFINE || len(assignment.Lhs) != 1 || len(assignment.Rhs) != 1 {
			continue
		}
		obj := object(pass, assignment.Lhs[0])
		if obj == nil {
			continue
		}
		if named := namedType(obj.Type()); named != nil {
			namedObject := named.Obj()
			namedPackage := namedObject.Pkg()
			if namedPackage != nil && namedObject.Name() == "Set" {
				pkg := librarypath.Canonical(pass, namedPackage.Path())
				if pathseg.SameLibrary(pkg, "gitlab.gid.team/gid-data/tech/golang/libs/helper.git") || pathseg.SameLibrary(pkg, "gid.team/gid-data/tech/golang/libs/helper.git") {
					continue
				}
			}
		}
		objectType := obj.Type()
		mapType, ok := objectType.Underlying().(*types.Map)
		if !ok {
			continue
		}
		elementType := mapType.Elem()
		empty, ok := elementType.Underlying().(*types.Struct)
		if !ok || empty.NumFields() != 0 {
			continue
		}
		lookup, insertion, other := false, false, false
		// Classify every reference by its parent. Only membership lookup + insertion is a set.
		var stack []ast.Node
		ast.Inspect(block, func(n ast.Node) bool {
			if n == nil {
				stack = stack[:len(stack)-1]
				return true
			}
			if id, ok := n.(*ast.Ident); ok && pass.TypesInfo.Uses[id] == obj {
				switch setReferenceRole(stack, id) {
				case 1:
					lookup = true
				case 2:
					insertion = true
				default:
					other = true
				}
			}
			stack = append(stack, n)
			return true
		})
		if lookup && insertion && !other {
			report(pass, assignment, "286", "manual map implements a membership set", "declare seen := gdhelper.NewSet[string]() for uniqueness")
		}
	}
}

func sentinelDecl(pass *analysis.Pass, decl ast.Decl) bool {
	gd, ok := decl.(*ast.GenDecl)
	if !ok || gd.Tok != token.VAR || len(gd.Specs) == 0 {
		return false
	}
	errorObject := types.Universe.Lookup("error")
	errorType := errorObject.Type()
	for _, spec := range gd.Specs {
		value, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}
		for _, id := range value.Names {
			obj := pass.TypesInfo.ObjectOf(id)
			if obj == nil || id.Name == "_" || !types.AssignableTo(obj.Type(), errorType) {
				return false
			}
		}
	}
	return true
}

// checkEnumSwitch judges only typed enum switches in the orchestration layers.
func checkEnumSwitch(pass *analysis.Pass, stmt *ast.SwitchStmt) {
	if stmt.Tag == nil || !enumOperand(pass, stmt.Tag) {
		return
	}
	handledDefault := false
	for _, node := range stmt.Body.List {
		clause, ok := node.(*ast.CaseClause)
		if !ok {
			continue
		}
		if clause.List != nil {
			continue
		}
		unhandledVars := map[types.Object]bool{}
		for _, statement := range clause.Body {
			if assignment, ok := statement.(*ast.AssignStmt); ok {
				for i, lhs := range assignment.Lhs {
					obj := object(pass, lhs)
					if obj == nil {
						continue
					}
					unhandledVars[obj] = i < len(assignment.Rhs) && containsUnhandled(pass, assignment.Rhs[i])
				}
			}
			ret, ok := statement.(*ast.ReturnStmt)
			if !ok {
				continue
			}
			for _, result := range ret.Results {
				if unhandledVars[object(pass, result)] || containsUnhandled(pass, result) {
					handledDefault = true
				}
			}
		}
	}
	if !handledDefault {
		report(pass, stmt, "291", "enum switch has no default returning an unhandled-value error", "add default returning gderr.NewUnhandledValueError(value)")
	}
}

func checkQueryAssignments(pass *analysis.Pass, block *ast.BlockStmt) {
	for _, statement := range block.List {
		assignment, ok := statement.(*ast.AssignStmt)
		if !ok {
			continue
		}
		for _, rhs := range assignment.Rhs {
			tv := pass.TypesInfo.Types[rhs]
			if tv.Value == nil || tv.Value.Kind() != constant.String {
				continue
			}
			query := constant.StringVal(tv.Value)
			if sqlStart.MatchString(query) && positional.MatchString(sqlCode(query)) {
				report(pass, rhs, "284", "repository SQL uses positional placeholders", "use @id with gdpostgres.NamedArgs")
			}
		}
	}
}

type functionContext struct {
	start, end  token.Pos
	errorResult bool
	body        *ast.BlockStmt
}

func returnsError(pass *analysis.Pass, fn *ast.FuncType) bool {
	if fn.Results == nil || len(fn.Results.List) == 0 {
		return false
	}
	last := fn.Results.List[len(fn.Results.List)-1]
	errorObject := types.Universe.Lookup("error")
	return types.Identical(pass.TypesInfo.TypeOf(last.Type), errorObject.Type())
}

func inErrorFunction(functions []*functionContext, pos token.Pos) bool {
	var innermost functionContext
	for _, fn := range functions {
		if fn.start <= pos && pos < fn.end && fn.start >= innermost.start {
			innermost = *fn
		}
	}
	return innermost.errorResult
}

func containsUnhandled(pass *analysis.Pass, expr ast.Expr) bool {
	found := false
	ast.Inspect(expr, func(n ast.Node) bool {
		if _, ok := n.(*ast.FuncLit); ok {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		fn := typeutil.StaticCallee(pass.TypesInfo, call)
		if fn != nil && fn.Pkg() != nil && fn.Pkg() != pass.Pkg && (fn.Name() == "NewUnhandledValueError" || fn.Name() == "Unhandled") {
			found = true
		}
		return true
	})
	return found
}

func hasFollowingSuccessReturn(pass *analysis.Pass, functions []*functionContext, guard *ast.IfStmt, errObj types.Object) bool {
	var enclosing functionContext
	for _, fn := range functions {
		if fn.start <= guard.Pos() && guard.Pos() < fn.end && fn.start >= enclosing.start {
			enclosing = *fn
		}
	}
	if !enclosing.errorResult {
		return false
	}
	positive := false
	for _, statement := range enclosing.body.List {
		ret, ok := statement.(*ast.ReturnStmt)
		if !ok || ret.Pos() < guard.End() || len(ret.Results) == 0 {
			continue
		}
		last, ok := ret.Results[len(ret.Results)-1].(*ast.Ident)
		if ok && pass.TypesInfo.ObjectOf(last) == types.Universe.Lookup("nil") {
			positive = true
		}
	}
	if !positive {
		return false
	}
	propagated := false
	ast.Inspect(enclosing.body, func(n ast.Node) bool {
		if _, ok := n.(*ast.FuncLit); ok {
			return false
		}
		ret, ok := n.(*ast.ReturnStmt)
		if !ok || ret.Pos() < guard.End() {
			return true
		}
		for _, expr := range ret.Results {
			if object(pass, expr) == errObj {
				propagated = true
			}
		}
		return true
	})
	return !propagated
}

// setReferenceRole classifies a membership lookup or insertion, zero for other uses.
func setReferenceRole(stack []ast.Node, id *ast.Ident) int {
	if len(stack) < 2 {
		return 0
	}
	index, ok := stack[len(stack)-1].(*ast.IndexExpr)
	if !ok || index.X != id {
		return 0
	}
	owner, ok := stack[len(stack)-2].(*ast.AssignStmt)
	if !ok || len(owner.Rhs) != 1 {
		return 0
	}
	switch {
	case len(owner.Lhs) == 2 && owner.Rhs[0] == index:
		return 1
	case len(owner.Lhs) == 1 && owner.Lhs[0] == index:
		lit, ok := owner.Rhs[0].(*ast.CompositeLit)
		if ok && len(lit.Elts) == 0 {
			return 2
		}
	}
	return 0
}

func checkSQLCall(pass *analysis.Pass, call *ast.CallExpr) {
	// A type conversion is part of the expression, not a query consumer.
	if pass.TypesInfo.Types[call.Fun].IsType() {
		return
	}
	for _, arg := range call.Args {
		switch ast.Unparen(arg).(type) {
		case *ast.Ident, *ast.SelectorExpr:
			// Named constants already satisfy the extraction requirement.
			continue
		}
		tv := pass.TypesInfo.Types[arg]
		if tv.Value == nil || tv.Value.Kind() != constant.String {
			continue
		}
		query := constant.StringVal(tv.Value)
		if !sqlStart.MatchString(query) {
			continue
		}
		report(pass, arg, "297", "constant SQL is passed directly to a call", "declare const sqlQuery = `SELECT 1` before the call and pass sqlQuery")
		if positional.MatchString(sqlCode(query)) {
			report(pass, arg, "284", "repository SQL uses positional placeholders", "use @id with gdpostgres.NamedArgs")
		}
	}
}
