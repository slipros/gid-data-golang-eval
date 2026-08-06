// Package docref implements rule GID-262 (linter giddocref): a comment
// explains the code, not the paperwork the code was written from.
//
// The shape it closes: a comment whose payload is a pointer into the
// development documentation — a requirement id (`@ФТ-11`, `ФТ33`), a decision
// code of an ARD (`ARD Р-11`, `К-3`), a section (`ARD §12`), a backlog entry
// (`BACKLOG B-48`), a task number (`задача 29`, `ревью задачи 20`), a BDD tag
// (`@negative`) or a traceability marker (`VERIFY спеки`). The document is not
// in the repository and outlives no refactor, so the reader of the function
// gets a token they cannot resolve instead of the constraint itself. The
// constraint is what belongs in the source: "one call per page — a per-item
// resolve would be N requests" says everything `(ARD Р-11, задача 29)` was
// meant to say, and keeps saying it after the ARD is rewritten.
//
// Detection: a regexp scan over the text of every comment in the file
// (settings.patterns, plus settings.extra). One diagnostic per comment — on
// the leftmost match — so a comment naming three documents is one finding.
//
// The fix depends on the class of the marker. A task number or a backlog entry
// is dropped and replaced by the constraint it stood for; a **requirement id**
// (`@ФТ-11`, an acceptance-criteria phrase, a VERIFY/BDD marker) is *moved*:
// it is the only record of which test proves which requirement, so the
// diagnostic asks for it to go into the requirement map — a file of its own
// linked from the README, one line per requirement
// (`ФТ-15 → TestCreate_DuplicateTitle_AlreadyExists`). Deleting those ids
// comment by comment would dissolve a coverage map nothing else holds.
//
// Skipped: generated code (ast.IsGenerated) and directive comments
// (`//nolint:…`, `//go:build`, `//lint:…` — a tool directive is machine input,
// and the free-text justification of a //nolint is exactly the place where
// naming the decision that granted the exception is legitimate).
//
// A _test.go file is **not** judged (internal/srcfile.IsTest). The requirement
// map the rule asks production code to hand over is `ФТ-15 → the test that
// proves it`, and the test is the one place where that pairing is not a
// pointer into a foreign document but a fact about the file it sits in: the
// name of the test is right there. Judging tests would delete the very
// trace the map is built from — settings.include-tests turns them back on
// once the map has been extracted into a file of its own.
package docref

import (
	"go/ast"
	"go/token"
	"regexp"
	"strings"
	"unicode"

	"github.com/pkg/errors"
	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/srcfile"
)

const ruleID = "GID-262"

// The two halves of the fix, picked by the class of the marker that matched.
//
// A requirement id is not noise the way a task number is: it is the only
// record of which test proves which requirement, and deleting it comment by
// comment dissolves the coverage map. So the fix for it is a move, not a
// deletion — the map goes into a file of its own, linked from the README,
// where it can be read as a whole and rebuilt after the comments are cleaned.
const (
	fixGeneric = `state the constraint itself — "one call per page: a per-item resolve would be N requests" — ` +
		`and leave the document reference in the document`
	fixRequirement = `state the constraint itself in the comment, and move the requirement id into the requirement map — ` +
		`a file of its own linked from the README, one line per requirement: "ФТ-15 → TestCreate_DuplicateTitle_AlreadyExists". ` +
		`The map is then readable as a whole and survives the cleanup, instead of living scattered across comments`
)

// defaultPatterns — the built-in marker list, calibrated on the services of
// the UDMP-3762 sandbox (resource-registry, advertising-api,
// ad-cabinet-connector, lk-api).
//
// Latin abbreviations are anchored with \b (ASCII word boundary), so
// TRACE_ENABLED, SkipVerify and ADR inside an identifier stay clean; the
// Cyrillic markers carry their own left delimiter [^\p{L}\d], which is what
// keeps UTF-8 out of the single-letter document-code pattern.
var defaultPatterns = []pattern{
	// Names of the development documents themselves.
	{expr: `\b(?:ARD|PRD|BDD|ADR|BACKLOG)\b`, fix: fixGeneric},
	// The same in Russian: спека, постановка, бэклог (any inflection).
	{expr: `(?:^|[^\p{L}])(?:спек|постановк|бэклог|беклог)\p{L}*`, fix: fixGeneric},
	// Requirement id: ФТ-33, ФТ33, @ФТ-11, НФТ-2, ФР-4.
	{expr: `(?i)@?Н?Ф[ТР]\s?-?\s?\d+`, fix: fixRequirement},
	// Document decision/contract code: Р-11, К-3, B-48, A-2.
	{expr: `(?:^|[^\p{L}\d])[A-ZА-Я]-\d+\b`, fix: fixGeneric},
	// Section of a document: §12, § 8.
	{expr: `§\s*\d+`, fix: fixGeneric},
	// Task of the decomposition: задача 29, задачи 13, ревью задачи 20.
	{expr: `(?i)(?:^|[^\p{L}])задач\p{L}*\s*(?:№\s*)?\d+`, fix: fixGeneric},
	// Acceptance criteria of a task.
	{expr: `(?i)критери\p{L}*\s+при[её]мк\p{L}*`, fix: fixRequirement},
	// Traceability marker of a BDD suite.
	{expr: `\bVERIFY\b`, fix: fixRequirement},
	// BDD scenario tags.
	{expr: `@(?:positive|negative|boundary|smoke)\b`, fix: fixRequirement},
	// A commit named as the source of a decision.
	{expr: `(?i)(?:коммит|commit)\p{L}*\s+[0-9a-f]{7,40}\b`, fix: fixGeneric},
}

// directivePrefix matches a tool directive comment (//nolint:…, //go:build,
// //lint:ignore): machine input, not prose.
var directivePrefix = regexp.MustCompile(`^//[a-z0-9]+:\S`)

// Analyzer — the variant with default settings.
var Analyzer = NewAnalyzer(Settings{})

// pattern — one marker of the development documentation and the fix its class
// calls for.
type pattern struct {
	expr string
	fix  string
}

// marker — a compiled pattern.
type marker struct {
	re  *regexp.Regexp
	fix string
}

// Settings — linter settings from .golangci.yml.
type Settings struct {
	// Patterns — the marker list, replacing the built-in one entirely.
	// Empty → defaultPatterns.
	Patterns []string `json:"patterns"`
	// Extra — additional markers, always appended to whatever Patterns
	// resolved to (a project's tracker key: `\bUDMP-\d+\b`).
	Extra []string `json:"extra"`
	// IncludeTests judges _test.go files too. Off by default: the requirement
	// map lives in the tests until it is extracted into a file of its own —
	// turn this on once it has been, and the ids leave the tests as well.
	IncludeTests bool `json:"include-tests"`
}

// NewAnalyzer builds the GID-262 analyzer with the given settings.
//
// A pattern given in settings gets the generic fix: only the built-in list
// knows which of its markers is a requirement id, i.e. which finding has a
// coverage map to preserve.
func NewAnalyzer(cfg Settings) *analysis.Analyzer {
	patterns := defaultPatterns
	if len(cfg.Patterns) > 0 {
		patterns = withGenericFix(cfg.Patterns)
	}
	patterns = append(append([]pattern{}, patterns...), withGenericFix(cfg.Extra)...)
	markers, compileErr := compile(patterns)

	return &analysis.Analyzer{
		Name: "giddocref",
		Doc: ruleID + ": a comment references development documentation (ARD/PRD/backlog id, requirement id, task number) " +
			"instead of explaining the code. Fix: state the constraint itself; a requirement id moves into the requirement map " +
			"(a file of its own linked from the README, \"ФТ-15 → TestCreate_DuplicateTitle_AlreadyExists\"), not into the void.",
		Run: func(pass *analysis.Pass) (any, error) {
			if compileErr != nil {
				return nil, compileErr
			}

			return run(pass, markers, cfg.IncludeTests)
		},
	}
}

func withGenericFix(exprs []string) []pattern {
	patterns := make([]pattern, 0, len(exprs))
	for _, expr := range exprs {
		patterns = append(patterns, pattern{expr: expr, fix: fixGeneric})
	}

	return patterns
}

func compile(patterns []pattern) ([]marker, error) {
	markers := make([]marker, 0, len(patterns))
	//nolint:gidallptr // the plugin does not depend on the internal gdhelper library
	for _, p := range patterns {
		re, err := regexp.Compile(p.expr)
		if err != nil {
			return nil, errors.Wrapf(err, "compile marker pattern %q", p.expr)
		}
		markers = append(markers, marker{re: re, fix: p.fix})
	}

	return markers, nil
}

func run(pass *analysis.Pass, markers []marker, includeTests bool) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		if !includeTests && srcfile.IsTest(pass, file) {
			continue
		}
		for _, group := range file.Comments {
			for _, comment := range group.List {
				if directivePrefix.MatchString(comment.Text) {
					continue
				}
				if found, ok := firstMarker(comment.Text, markers); ok {
					report(pass, comment.Pos()+token.Pos(found.offset), found)
				}
			}
		}
	}

	return nil, nil
}

// reference — the document reference a comment was reported for.
type reference struct {
	// offset — byte offset of the reference inside the comment text.
	offset int
	// text — the reference as written (`@ФТ-11`, `BACKLOG`).
	text string
	// fix — the hint of the marker class that matched.
	fix string
}

// firstMarker returns the leftmost reference of the comment: one diagnostic
// per comment, pointed at the first document reference it carries.
func firstMarker(text string, markers []marker) (found reference, ok bool) {
	found.offset = -1
	//nolint:gidallptr // the plugin does not depend on the internal gdhelper library
	for _, m := range markers {
		loc := m.re.FindStringIndex(text)
		if loc == nil {
			continue
		}
		start, marker := trimDelimiter(text[loc[0]:loc[1]], loc[0])
		if found.offset < 0 || start < found.offset {
			found = reference{offset: start, text: marker, fix: m.fix}
		}
	}

	return found, found.offset >= 0
}

// trimDelimiter drops the left delimiter a pattern had to capture in place of
// a Unicode word boundary (`(ARD`, ` задача 29`), so both the position and the
// quoted marker of the diagnostic point at the reference itself.
func trimDelimiter(match string, offset int) (start int, marker string) {
	trimmed := strings.TrimLeftFunc(match, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '@' && r != '§'
	})

	return offset + len(match) - len(trimmed), trimmed
}

func report(pass *analysis.Pass, pos token.Pos, found reference) {
	pass.Reportf(pos,
		"%s: comment references development documentation (%q) instead of explaining the code. Fix: %s",
		ruleID, found.text, found.fix)
}
