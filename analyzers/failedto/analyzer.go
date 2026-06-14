// Package failedto implements rule GID-184 (linter gidfailedto):
// an error message describes the operation, not the fact of failure
// (Uber: error wrapping). A string literal message in calls to
// errors.Wrap/Wrapf/WithMessage/WithMessagef/Errorf/New of the
// github.com/pkg/errors package must not start (case-insensitively) with
// a forbidden prefix (failed to, failed, unable to, error, couldn't,
// could not, can't, cannot, etc.).
//
// Instead of "failed to select user" → "select user": when the chain is
// unwound, the message reads as a description of operations.
//
// pkg/errors is detected by the import path github.com/pkg/errors via
// TypesInfo. A non-literal message (a variable, concatenation with a
// variable) is not checked. Generated code (ast.IsGenerated)
// is skipped.
package failedto

import (
	"go/ast"
	"go/constant"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"
)

const ruleID = "GID-184"

// errFuncs — pkg/errors functions whose message argument is checked.
// The value is the index of the message argument in the call.
var errFuncs = map[string]int{
	"Wrap":         1, // Wrap(err, message)
	"Wrapf":        1, // Wrapf(err, format, ...)
	"WithMessage":  1, // WithMessage(err, message)
	"WithMessagef": 1, // WithMessagef(err, format, ...)
	"Errorf":       0, // Errorf(format, ...)
	"New":          0, // New(message)
}

// defaultPrefixes — forbidden message prefixes (case-insensitive).
var defaultPrefixes = []string{
	"failed to",
	"failed",
	"unable to",
	"error",
	"couldn't",
	"could not",
	"can't",
	"cannot",
}

// Analyzer — variant with default settings.
var Analyzer = NewAnalyzer(Settings{})

// Settings — linter settings from .golangci.yml.
type Settings struct {
	// Prefixes — forbidden message prefixes. When set, it replaces
	// the default list entirely.
	Prefixes []string `json:"prefixes"`
}

// NewAnalyzer builds the GID-184 analyzer with the given settings.
func NewAnalyzer(s Settings) *analysis.Analyzer {
	prefixes := defaultPrefixes
	if len(s.Prefixes) > 0 {
		prefixes = s.Prefixes
	}
	lower := make([]string, len(prefixes))
	for i, p := range prefixes {
		lower[i] = strings.ToLower(p)
	}
	return &analysis.Analyzer{
		Name: "gidfailedto",
		Doc:  ruleID + ": an error message describes the operation, not the fact of failure. Fix: drop prefixes like 'failed to'",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, lower)
		},
	}
}

func run(pass *analysis.Pass, prefixes []string) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			name := pkgErrorsCallName(pass, call)
			if name == "" {
				return true
			}
			idx, ok := errFuncs[name]
			if !ok || idx >= len(call.Args) {
				return true
			}
			msg, ok := stringLiteral(pass, call.Args[idx])
			if !ok {
				return true
			}
			if prefix, hit := matchPrefix(msg, prefixes); hit {
				pass.Reportf(call.Args[idx].Pos(),
					"%s: error message starts with %q. Fix: describe the operation, e.g. \"failed to select user\" → \"select user\"",
					ruleID, prefix)
			}
			return true
		})
	}
	return nil, nil
}

// matchPrefix reports whether the message starts (case-insensitively) with a
// forbidden prefix as a word (the prefix is followed by end of string or a
// non-letter/digit, so that "failure" does not match "failed"... or "fail" —
// there is no "fail" prefix here, but the word boundary guards against substrings).
func matchPrefix(msg string, prefixes []string) (string, bool) {
	low := strings.ToLower(strings.TrimSpace(msg))
	for _, p := range prefixes {
		if !strings.HasPrefix(low, p) {
			continue
		}
		rest := low[len(p):]
		if rest == "" {
			return p, true
		}
		r := rune(rest[0])
		if !isWordRune(r) {
			return p, true
		}
	}
	return "", false
}

func isWordRune(r rune) bool {
	return r == '_' ||
		(r >= 'a' && r <= 'z') ||
		(r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9')
}

// stringLiteral returns the value of a string literal (including an
// untyped string constant). A variable or concatenation with a
// variable → (_, false): such messages are not checked.
func stringLiteral(pass *analysis.Pass, expr ast.Expr) (string, bool) {
	tv, ok := pass.TypesInfo.Types[expr]
	if !ok || tv.Value == nil || tv.Value.Kind() != constant.String {
		return "", false
	}
	// Literals and constants only — variables have no value here.
	return constant.StringVal(tv.Value), true
}

// pkgErrorsCallName returns the name of the github.com/pkg/errors function
// if call invokes it; otherwise "".
func pkgErrorsCallName(pass *analysis.Pass, call *ast.CallExpr) string {
	const pkgErrorsPath = "github.com/pkg/errors"
	fn := typeutil.Callee(pass.TypesInfo, call)
	f, ok := fn.(*types.Func)
	if !ok || f.Pkg() == nil {
		return ""
	}
	pkg := f.Pkg()
	if pkg.Path() != pkgErrorsPath {
		return ""
	}
	return f.Name()
}
