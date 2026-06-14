// Package createupdate implements rule GID-112: methods that create an
// entity or update state (Create*/Update*) in repo and service
// return only error. If data is needed after creation, the calling
// code fetches it with a separate query.
//
// Exceptions (sometimes it is convenient to get the entity right away):
//   - targeted: //nolint:gidcreateupdate
//   - centralized: settings.exclude in .golangci.yml —
//     entries like "CreateSession" (a method name) or "Job.CreateJob"
//     (a specific type).
package createupdate

import (
	"go/ast"
	"go/types"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/exclude"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-112"

var verbs = []string{"Create", "Update"}

var scopes = [][]string{
	{"dal", "repository"},
	{"domain", "service"},
}

// Analyzer — the variant with default settings (no exclusions).
var Analyzer = NewAnalyzer(Settings{})

// Settings — linter settings from .golangci.yml.
type Settings struct {
	// Exclude — excluded methods: "Method" or "Type.Method".
	Exclude []string `json:"exclude"`
}

// NewAnalyzer builds the GID-112 analyzer from the linter settings (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "gidcreateupdate",
		Doc:  ruleID + ": Create*/Update* methods in repo and service must return only error. Fix: drop the extra return and fetch the entity with a separate query",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, s)
		},
	}
}

func run(pass *analysis.Pass, s Settings) (any, error) {
	if !inScope(pass.Pkg.Path()) {
		return nil, nil
	}
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || !fn.Name.IsExported() {
				continue
			}
			if !hasVerbPrefix(fn.Name.Name) {
				continue
			}
			if exclude.Match(s.Exclude, recvTypeName(fn), fn.Name.Name) {
				continue
			}
			checkResults(pass, fn)
		}
	}
	return nil, nil
}

func checkResults(pass *analysis.Pass, fn *ast.FuncDecl) {
	obj, ok := pass.TypesInfo.Defs[fn.Name].(*types.Func)
	if !ok {
		return
	}
	sig, ok := obj.Type().(*types.Signature)
	if !ok {
		return
	}
	results := sig.Results()
	if results.Len() == 0 {
		return
	}
	for v := range results.Variables() {
		if isError(v.Type()) {
			continue
		}
		pass.Reportf(fn.Name.Pos(),
			"%s: method %q creates/updates state and must return only error. "+
				"Fix: fetch the entity with a separate query (exceptions: nolint or settings.exclude)",
			ruleID, fn.Name.Name)
		return
	}
}

func inScope(pkgPath string) bool {
	for _, scope := range scopes {
		if pathseg.Contains(pkgPath, scope...) {
			return true
		}
	}
	return false
}

// hasVerbPrefix: the name starts with the word Create/Update
// (CreateJob, Update — yes; CreatedAt — no).
func hasVerbPrefix(name string) bool {
	for _, verb := range verbs {
		if name == verb {
			return true
		}
		if len(name) > len(verb) && name[:len(verb)] == verb {
			r, _ := utf8.DecodeRuneInString(name[len(verb):])
			if unicode.IsUpper(r) || unicode.IsDigit(r) {
				return true
			}
		}
	}
	return false
}

func recvTypeName(fn *ast.FuncDecl) string {
	t := fn.Recv.List[0].Type
	if star, ok := t.(*ast.StarExpr); ok {
		t = star.X
	}
	if ident, ok := t.(*ast.Ident); ok {
		return ident.Name
	}
	return ""
}

func isError(t types.Type) bool {
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	return obj.Pkg() == nil && obj.Name() == "error"
}
