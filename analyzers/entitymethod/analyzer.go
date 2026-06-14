// Package entitymethod implements rule GID-114: exported struct methods
// in the root packages of the /dal/repository and /domain/service layers
// are named after the entity.
//
// Three checks:
//  1. the List prefix is forbidden — use the plural instead (Jobs, not ListJobs);
//  2. the ByID suffix is forbidden — Job(ctx, id) instead of JobByID
//     (only the exact ByID suffix; ByStageID and other By<Field>ID are allowed —
//     that is a query refinement, not fetching by primary key);
//  3. the method name must contain the entity name — the receiver type name
//     as a CamelCase substring (Job → Job, Jobs, CreateJob, JobsByStageID).
//
// Check 3 applies only to receivers with a meaningful entity name
// (len > 2); single-letter/utility names are not checked. Verb methods
// without an entity name (Close, Ping, Flush) will hit check 3 — they are
// rarely legitimate and are disabled via exclude/nolint.
//
// Scope — only the root packages of the layer (pathseg.EndsWith); the
// convert/build subpackages are not touched. New* constructors are functions,
// not methods, and do not fall under this rule.
//
// Exceptions:
//   - targeted: //nolint:gidentitymethod
//   - centralized: settings.exclude in .golangci.yml —
//     entries like "Close" (a method name) or "Job.Close" (a specific type).
package entitymethod

import (
	"go/ast"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/exclude"
	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

const ruleID = "GID-114"

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

// NewAnalyzer builds the GID-114 analyzer from the linter settings (.golangci.yml).
func NewAnalyzer(s Settings) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name: "gidentitymethod",
		Doc: ruleID + ": repo/service methods are named after the entity, " +
			"without a List prefix, without a ByID suffix, including the entity name. Fix: rename accordingly",
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
			recv := recvTypeName(fn)
			name := fn.Name.Name
			if exclude.Match(s.Exclude, recv, name) {
				continue
			}
			checkName(pass, fn, recv, name)
		}
	}
	return nil, nil
}

func checkName(pass *analysis.Pass, fn *ast.FuncDecl, recv, name string) {
	// Check 1: the List prefix is forbidden.
	if hasWordPrefix(name, "List") {
		pass.Reportf(fn.Name.Pos(),
			"%s: drop the List prefix. Fix: use the plural Jobs instead of ListJobs",
			ruleID)
		return
	}
	// Check 2: the exact ByID suffix is forbidden (ByStageID and others are allowed).
	if hasExactByIDSuffix(name) {
		pass.Reportf(fn.Name.Pos(),
			"%s: drop the ByID suffix. Fix: use Job(ctx, id) instead of JobByID",
			ruleID)
		return
	}
	// Check 3: the method name must contain the entity name (the receiver name)
	// as a CamelCase substring. Only for meaningful entity names:
	// names of length <= 2 (T, ID, etc.) are treated as utility names and skipped.
	const minEntityLen = 2
	if len(recv) <= minEntityLen {
		return
	}
	if !containsEntity(name, recv) {
		pass.Reportf(fn.Name.Pos(),
			"%s: method name %q must contain the entity name %q "+
				"(Job, Jobs, CreateJob, JobsByStageID; exceptions: nolint or settings.exclude)",
			ruleID, name, recv)
	}
}

func inScope(pkgPath string) bool {
	for _, scope := range scopes {
		if pathseg.EndsWith(pkgPath, scope...) {
			return true
		}
	}
	return false
}

// hasWordPrefix: the name starts with the word at a CamelCase boundary
// (List, ListJobs — yes; Listen — no, since the next rune is not uppercase).
func hasWordPrefix(name, word string) bool {
	if name == word {
		return true
	}
	if len(name) <= len(word) || name[:len(word)] != word {
		return false
	}
	r, _ := utf8.DecodeRuneInString(name[len(word):])
	return unicode.IsUpper(r) || unicode.IsDigit(r)
}

// hasExactByIDSuffix: the name ends exactly with "ByID" at a word boundary.
// JobByID — yes; JobsByStageID — no (the part before ID is not "By"); a standalone
// ByID — no (it is not an entity name with a suffix, but also not valid — check 3 catches it).
func hasExactByIDSuffix(name string) bool {
	const suffix = "ByID"
	if !strings.HasSuffix(name, suffix) {
		return false
	}
	return len(name) > len(suffix)
}

// containsEntity: the method name contains entity as a CamelCase substring.
// A word boundary is the start of the name, or a preceding lowercase rune
// before the uppercase first rune of entity (CreateJob: ...e|Job).
func containsEntity(name, entity string) bool {
	for idx := strings.Index(name, entity); idx >= 0; idx = nextIndex(name, entity, idx) {
		if isWordBoundary(name, idx) {
			return true
		}
	}
	return false
}

func nextIndex(name, entity string, prev int) int {
	rest := strings.Index(name[prev+1:], entity)
	if rest < 0 {
		return -1
	}
	return prev + 1 + rest
}

// isWordBoundary: the position idx starts a CamelCase word.
// True if idx == 0 or the preceding rune is not uppercase
// (the camelCase boundary: lowerUpper). This cuts off matches inside a word.
func isWordBoundary(name string, idx int) bool {
	if idx == 0 {
		return true
	}
	prev, _ := utf8.DecodeLastRuneInString(name[:idx])
	return !unicode.IsUpper(prev)
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
