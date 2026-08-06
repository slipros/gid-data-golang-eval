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
// Skipped: generated code (ast.IsGenerated) and directive comments
// (`//nolint:…`, `//go:build`, `//lint:…` — a tool directive is machine input,
// and the free-text justification of a //nolint is exactly the place where
// naming the decision that granted the exception is legitimate).
//
// A _test.go file is judged: a test doc comment states what the test proves
// as easily as production code does, and nothing forces a double to carry a
// requirement id.
package docref

import (
	"go/ast"
	"go/token"
	"regexp"
	"strings"
	"unicode"

	"github.com/pkg/errors"
	"golang.org/x/tools/go/analysis"
)

const ruleID = "GID-262"

// defaultPatterns — the built-in marker list, calibrated on the services of
// the UDMP-3762 sandbox (resource-registry, advertising-api,
// ad-cabinet-connector, lk-api).
//
// Latin abbreviations are anchored with \b (ASCII word boundary), so
// TRACE_ENABLED, SkipVerify and ADR inside an identifier stay clean; the
// Cyrillic markers carry their own left delimiter [^\p{L}\d], which is what
// keeps UTF-8 out of the single-letter document-code pattern.
var defaultPatterns = []string{
	// Names of the development documents themselves.
	`\b(?:ARD|PRD|BDD|ADR|BACKLOG)\b`,
	// The same in Russian: спека, постановка, бэклог (any inflection).
	`(?:^|[^\p{L}])(?:спек|постановк|бэклог|беклог)\p{L}*`,
	// Requirement id: ФТ-33, ФТ33, @ФТ-11, НФТ-2, ФР-4.
	`(?i)@?Н?Ф[ТР]\s?-?\s?\d+`,
	// Document decision/contract code: Р-11, К-3, B-48, A-2.
	`(?:^|[^\p{L}\d])[A-ZА-Я]-\d+\b`,
	// Section of a document: §12, § 8.
	`§\s*\d+`,
	// Task of the decomposition: задача 29, задачи 13, ревью задачи 20.
	`(?i)(?:^|[^\p{L}])задач\p{L}*\s*(?:№\s*)?\d+`,
	// Acceptance criteria of a task.
	`(?i)критери\p{L}*\s+при[её]мк\p{L}*`,
	// Traceability marker of a BDD suite.
	`\bVERIFY\b`,
	// BDD scenario tags.
	`@(?:positive|negative|boundary|smoke)\b`,
	// A commit named as the source of a decision.
	`(?i)(?:коммит|commit)\p{L}*\s+[0-9a-f]{7,40}\b`,
}

// directivePrefix matches a tool directive comment (//nolint:…, //go:build,
// //lint:ignore): machine input, not prose.
var directivePrefix = regexp.MustCompile(`^//[a-z0-9]+:\S`)

// Analyzer — the variant with default settings.
var Analyzer = NewAnalyzer(Settings{})

// Settings — linter settings from .golangci.yml.
type Settings struct {
	// Patterns — the marker list, replacing the built-in one entirely.
	// Empty → defaultPatterns.
	Patterns []string `json:"patterns"`
	// Extra — additional markers, always appended to whatever Patterns
	// resolved to (a project's tracker key: `\bUDMP-\d+\b`).
	Extra []string `json:"extra"`
}

// NewAnalyzer builds the GID-262 analyzer with the given settings.
func NewAnalyzer(cfg Settings) *analysis.Analyzer {
	patterns := cfg.Patterns
	if len(patterns) == 0 {
		patterns = defaultPatterns
	}
	patterns = append(append([]string{}, patterns...), cfg.Extra...)
	markers, compileErr := compile(patterns)

	return &analysis.Analyzer{
		Name: "giddocref",
		Doc: ruleID + ": a comment references development documentation (ARD/PRD/backlog id, requirement id, task number) " +
			"instead of explaining the code. Fix: state the constraint itself — \"one call per page: a per-item resolve would be N requests\".",
		Run: func(pass *analysis.Pass) (any, error) {
			if compileErr != nil {
				return nil, compileErr
			}

			return run(pass, markers)
		},
	}
}

func compile(patterns []string) ([]*regexp.Regexp, error) {
	markers := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		re, err := regexp.Compile(p)
		if err != nil {
			return nil, errors.Wrapf(err, "compile marker pattern %q", p)
		}
		markers = append(markers, re)
	}

	return markers, nil
}

func run(pass *analysis.Pass, markers []*regexp.Regexp) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, group := range file.Comments {
			for _, comment := range group.List {
				if directivePrefix.MatchString(comment.Text) {
					continue
				}
				if offset, marker, ok := firstMarker(comment.Text, markers); ok {
					report(pass, comment.Pos()+token.Pos(offset), marker)
				}
			}
		}
	}

	return nil, nil
}

// firstMarker returns the leftmost marker of the comment: one diagnostic per
// comment, pointed at the first document reference it carries.
func firstMarker(text string, markers []*regexp.Regexp) (offset int, marker string, ok bool) {
	offset, marker = -1, ""
	for _, re := range markers {
		loc := re.FindStringIndex(text)
		if loc == nil {
			continue
		}
		start, found := trimDelimiter(text[loc[0]:loc[1]], loc[0])
		if offset < 0 || start < offset {
			offset, marker = start, found
		}
	}

	return offset, marker, offset >= 0
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

func report(pass *analysis.Pass, pos token.Pos, marker string) {
	pass.Reportf(pos,
		"%s: comment references development documentation (%q) instead of explaining the code. "+
			"Fix: state the constraint itself — \"one call per page: a per-item resolve would be N requests\" — "+
			"and leave the requirement id in the tracker",
		ruleID, marker)
}
