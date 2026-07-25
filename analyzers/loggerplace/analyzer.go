// Package loggerplace implements rule GID-251 (slug logger-not-in-options,
// linter gidloggerplace): the logger is not a field of an options struct — it
// is a separate constructor parameter.
//
// An options struct describes the entity's configuration: values a caller may
// tune, each with a meaningful default. A logger is not configuration — it is
// a dependency, and hiding it in Options has two consequences seen in
// practice: the constructor signature no longer says the entity logs at all
// (GID-153, which places the logger after opts, has nothing to place), and the
// "nil means slog.Default()" fallback that inevitably follows grabs the global
// logger deep inside a layer (GID-214).
//
// Scope: a struct type whose name ends with the "Options" suffix (the naming
// GID-126 enforces), in any layer. A field is a logger when its type belongs
// to a known logger stack — logrus or slog, see internal/lgr — so the rule is
// not pinned to one library. The suffix is configurable through
// settings.suffixes for a service that names its config types differently.
//
// Escape hatch: //nolint:gidloggerplace on the field.
package loggerplace

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/lgr"
)

const ruleID = "GID-251"

// defaultSuffixes — type-name suffixes that mark an options struct.
var defaultSuffixes = []string{"Options", "Config", "Settings"}

// Analyzer — GID-251 with the default suffixes.
var Analyzer = NewAnalyzer(Settings{})

// Settings — linter settings from .golangci.yml.
type Settings struct {
	// Suffixes — type-name suffixes that mark an options struct. Replaces the
	// default list ("Options", "Config", "Settings").
	Suffixes []string `json:"suffixes"`
}

// NewAnalyzer builds the GID-251 analyzer from the linter settings.
func NewAnalyzer(s Settings) *analysis.Analyzer {
	suffixes := s.Suffixes
	if len(suffixes) == 0 {
		suffixes = defaultSuffixes
	}
	return &analysis.Analyzer{
		Name: "gidloggerplace",
		Doc: ruleID + ": the logger is not a field of an options struct — a logger is a dependency, not " +
			"configuration. Fix: drop the field and take the logger as a separate constructor parameter, " +
			"after opts (GID-153)",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, suffixes)
		},
	}
}

func run(pass *analysis.Pass, suffixes []string) (any, error) {
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, spec := range gen.Specs {
				checkTypeSpec(pass, spec, suffixes)
			}
		}
	}
	return nil, nil
}

func checkTypeSpec(pass *analysis.Pass, spec ast.Spec, suffixes []string) {
	ts, ok := spec.(*ast.TypeSpec)
	if !ok || !hasSuffix(ts.Name.Name, suffixes) {
		return
	}
	st, ok := ts.Type.(*ast.StructType)
	if !ok {
		return
	}
	for _, field := range st.Fields.List {
		if !lgr.IsType(pass.TypesInfo.TypeOf(field.Type)) {
			continue
		}
		pass.Reportf(field.Pos(),
			"%s: options struct %q holds a logger — a logger is a dependency, not configuration. "+
				"Fix: drop the field and take the logger as a separate constructor parameter, after opts "+
				"(GID-153); the entity names itself on it in the constructor (GID-154)",
			ruleID, ts.Name.Name)
	}
}

func hasSuffix(name string, suffixes []string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(name, suffix) {
			return true
		}
	}
	return false
}
