// Package entitymethod implements rule GID-114: exported struct methods
// in the root packages of the /dal/repository and /domain/service layers
// are named after the entity.
//
// Three checks:
//  1. the List prefix is forbidden — use the plural instead (Jobs, not ListJobs);
//  2. the ByID suffix is forbidden — Job(ctx, id) instead of JobByID
//     (only the exact ByID suffix; ByStageID and other By<Field>ID are allowed —
//     that is a query refinement, not fetching by primary key);
//  3. the method name must name the entity — the receiver type name or its
//     leading CamelCase part, as a whole word (Job → Job, Jobs, CreateJob,
//     JobsByStageID).
//
// Check 3 matches the receiver name or any of its leading word parts longer
// than two characters: the role suffix of the type is not part of the entity
// name (AdCabinetResolver → AdCabinet.AdCabinet), and a specialised entity is
// legitimately called by its more general name (SegmentBuildOutbox →
// EnqueueSegmentBuild). A trailing plural "s" counts as the same word
// (SegmentCleanup → CleanupSegments). Receivers with no meaningful entity name
// (len <= 2) are not checked. Verb methods without an entity name (Close, Ping,
// Flush) will hit check 3 — they are rarely legitimate and are disabled via
// exclude/nolint.
//
// Check 3 does not judge an unexported receiver: such a type is an
// implementation detail — an adapter or a nop stub whose method names are
// dictated by the interface it implements, not by its own entity.
//
// Scope — only the root packages of the layer (pathseg.EndsWith); the
// convert/build subpackages are not touched. New* constructors are functions,
// not methods, and do not fall under this rule. A _test.go file is not judged
// at all: tests live in the same package (GID-250), and a test double
// implementing the interface under test copies the method names from that
// interface — renaming them is impossible.
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
	"github.com/slipros/gid-data-golang-eval/internal/srcfile"
)

const ruleID = "GID-114"

// minEntityLen — the length up to which a name (or a leading part of one) is a
// utility name (T, ID, Ad) rather than an entity name.
const minEntityLen = 2

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
		if ast.IsGenerated(file) || srcfile.IsTest(pass, file) {
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
	// Check 3: the method name must name the entity — the receiver name or its
	// leading word part. Only for meaningful entity names: names of length <= 2
	// (T, ID, etc.) are treated as utility names and skipped, and an unexported
	// receiver is an implementation detail whose method names come from the
	// interface it implements.
	if len(recv) <= minEntityLen || !ast.IsExported(recv) {
		return
	}
	if !namesEntity(name, recv) {
		pass.Reportf(fn.Name.Pos(),
			"%s: method name %q does not name the entity %q. Fix: use the entity name or its "+
				"leading part — Job, Jobs, CreateJob, JobsByStageID "+
				"(exceptions: nolint or settings.exclude)",
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

// namesEntity: the method name contains the entity name or any of its leading
// word parts. SegmentBuildOutbox is matched by SegmentBuildOutbox,
// SegmentBuild and Segment — a role suffix (Resolver, Service) is not part of
// the entity name, and a specialised entity may be called by its more general
// name. Parts of two characters or less (Ad, Id) are too weak to identify an
// entity and are not matched on their own.
func namesEntity(name, entity string) bool {
	for _, part := range leadingParts(entity) {
		if containsWord(name, part) {
			return true
		}
	}
	return false
}

// leadingParts returns the entity name and every leading CamelCase part of it
// longer than minEntityLen
// (SegmentBuildOutbox → SegmentBuildOutbox, Segment, SegmentBuild).
func leadingParts(entity string) []string {
	parts := []string{entity}
	for idx, r := range entity {
		if idx > minEntityLen && unicode.IsUpper(r) {
			parts = append(parts, entity[:idx])
		}
	}
	return parts
}

// containsWord: the method name contains word as a whole CamelCase word.
// The left boundary is the start of the name or a preceding non-uppercase rune
// (CreateJob: ...e|Job). The right boundary is the end of the name, the start
// of the next word (an uppercase rune or a digit), or a plural "s" followed by
// one of those (CleanupSegments, JobsByStageID) — so Ad does not match
// Advance, while Segment does match Segments.
func containsWord(name, word string) bool {
	for idx := strings.Index(name, word); idx >= 0; idx = nextIndex(name, word, idx) {
		if isWordBoundary(name, idx) && endsWord(name[idx+len(word):]) {
			return true
		}
	}
	return false
}

// endsWord reports whether rest — what follows the matched word — ends that
// word: the name is over, a new CamelCase word starts, or a plural "s" comes
// first.
func endsWord(rest string) bool {
	if rest == "" {
		return true
	}
	r, size := utf8.DecodeRuneInString(rest)
	if unicode.IsUpper(r) || unicode.IsDigit(r) {
		return true
	}
	if r != 's' {
		return false
	}
	rest = rest[size:]
	if rest == "" {
		return true
	}
	next, _ := utf8.DecodeRuneInString(rest)
	return unicode.IsUpper(next) || unicode.IsDigit(next)
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
