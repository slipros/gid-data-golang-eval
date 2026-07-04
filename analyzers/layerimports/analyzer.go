// Package layerimports implements the dependency-direction rules between
// Clean Architecture layers (full matrix — RULES.md, GID-132/170/172,
// GID-224…229).
//
// GID-132 (layer-imports):
//   - /dal/** does not import /domain/** — repository works only with entity;
//   - /domain/model does not import /dal/** — model is pure;
//   - /domain/usecase does not import /dal/** — usecase works only
//     with model and talks to DAL through services;
//   - /domain/service does not import /dal/repository — the dependency on
//     the repository is described by an interface next to the consumer.
//     Importing /dal/entity is allowed for a service: it converts model <-> entity.
//
// GID-170 (no-event-import):
//   - /domain/** does not import /event/**;
//   - /dal/** does not import /event/** — the event layer (kafka producer/consumer,
//     DTO) depends on domain/model and converts model <-> DTO, not the other way.
//
// GID-172 (client-no-entity):
//   - /client/** does not import /dal/** — the client has its own types and knows
//     nothing about entity/repository.
//
// GID-224 (transport-imports):
//   - transport (/server, /schedule, /validate, /event) sees only /domain/model
//     (and /validate) from the service layers — concrete service/usecase are
//     injected through interfaces at the consumer.
//
// GID-225 (root-and-leaves):
//   - /internal/app (composition root) and the transport leaves (/server,
//     /schedule, /validate) are imported by nobody.
//
// GID-226 (metric-standalone):
//   - /metric does not import service layers; domain/dal receive metrics
//     through an interface — /metric is not available to them (wiring in app).
//
// GID-227 (model-pure):
//   - /domain/model does not import any service layer — it is the pure
//     vocabulary; the subpackages /domain/model/* are a full-fledged model layer.
//
// GID-228 (no-direct-client):
//   - /domain/usecase does not import /client/** — a client is used by a
//     repository (client models are converted to entity in dal/repository/convert)
//     or directly by a service (the service converts model <-> client models;
//     its API always takes and returns model); /domain/model is shielded by GID-227.
//
// GID-229 (client-isolated):
//   - /client/** does not import service layers (including all of /domain) —
//     the client has its own types, conversion lives at the consumer.
//
// Bans apply only within a single module. The module boundary is resolved in
// priority order:
//  1. the /internal/ segment (canonical layout) — the module root is the
//     prefix before it;
//  2. otherwise a /pkg/<module>/ segment (module.md: the application-module
//     layout — pkg/<module> repeats the same layered structure as internal/,
//     scoped to one module) — the module root is <prefix>/pkg/<module>, and
//     the full layer matrix applies inside it exactly as inside internal/;
//  3. otherwise the first path segment (testdata, non-standard layout).
//
// Because the module root differs between /internal/ and /pkg/<module>, an
// import from pkg/<module>/** of repo/internal/** (shared entities) is,
// by this rule, a cross-module import — sameModule is false and the matrix
// does not ban it. This is intentional: module.md treats such access to
// common internal/ entities as legal.
//
// Third-party libraries with client/event/metric segments in their path are
// not affected.
//
// Per-project relaxation — settings.disable (list of GID-IDs); custom rules —
// settings.rules (id, scope, banned, reason). Pointwise — //nolint:gidlayerimports.
package layerimports

import (
	"go/ast"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"

	"github.com/slipros/gid-data-golang-eval/internal/pathseg"
)

// Rule order matters: the first matching rule is reported for an import
// (specific rules before general ones), so duplicate diagnostics are not produced.
var layerRules = []layerRule{
	{
		id:     "GID-132",
		scope:  []string{"dal"},
		banned: [][]string{{"domain"}},
		reason: "the dal layer works only with entity, domain types are not available to it",
	},
	{
		id:     "GID-132",
		scope:  []string{"domain", "model"},
		banned: [][]string{{"dal"}},
		reason: "model does not depend on the dal layer",
	},
	{
		id:     "GID-132",
		scope:  []string{"domain", "usecase"},
		banned: [][]string{{"dal"}},
		reason: "usecase works only with model and talks to DAL through services",
	},
	{
		id:     "GID-132",
		scope:  []string{"domain", "service"},
		banned: [][]string{{"dal", "repository"}},
		reason: "a service depends on the repository through an interface next to the consumer",
	},
	{
		id:     "GID-170",
		scope:  []string{"domain"},
		banned: [][]string{{"event"}},
		reason: "domain does not depend on the event layer; event converts model <-> DTO, not the other way",
	},
	{
		id:     "GID-170",
		scope:  []string{"dal"},
		banned: [][]string{{"event"}},
		reason: "dal does not depend on the event layer; event converts model <-> DTO, not the other way",
	},
	{
		id:     "GID-172",
		scope:  []string{"client"},
		banned: [][]string{{"dal"}},
		reason: "the client has its own types and knows nothing about entity/repository from the dal layer",
	},
	// --- layer isolation matrix (2026-06-07) ---
	{
		id:    "GID-227",
		scope: []string{"domain", "model"},
		banned: [][]string{
			{"domain", "service"}, {"domain", "usecase"},
			{"client"}, {"metric"}, {"server"}, {"schedule"}, {"validate"},
		},
		reason: "domain/model is the pure vocabulary of the service; layers do not flow into it",
	},
	{
		id:    "GID-226",
		scope: []string{"metric"},
		banned: [][]string{
			{"dal"}, {"domain"}, {"client"}, {"event"},
			{"server"}, {"schedule"}, {"validate"},
		},
		reason: "the metric package is a standalone Prometheus aggregator; service layers are not available to it",
	},
	{
		id:    "GID-229",
		scope: []string{"client"},
		banned: [][]string{
			{"domain"}, {"event"}, {"metric"},
			{"server"}, {"schedule"}, {"validate"},
		},
		reason: "the client has its own types; model <-> client DTO conversion lives at the consumer",
	},
	{
		id:    "GID-224",
		scope: []string{"server"},
		banned: [][]string{
			{"dal"}, {"domain", "service"}, {"domain", "usecase"},
			{"client"}, {"metric"}, {"event"}, {"schedule"}, {"app"},
		},
		reason: "transport works only with domain/model; services and dependencies are injected as interfaces at the consumer",
	},
	{
		id:    "GID-224",
		scope: []string{"schedule"},
		banned: [][]string{
			{"dal"}, {"domain", "service"}, {"domain", "usecase"},
			{"client"}, {"metric"}, {"event"}, {"server"}, {"app"},
		},
		reason: "transport works only with domain/model; services and dependencies are injected as interfaces at the consumer",
	},
	{
		id:    "GID-224",
		scope: []string{"validate"},
		banned: [][]string{
			{"dal"}, {"domain", "service"}, {"domain", "usecase"},
			{"client"}, {"metric"}, {"event"}, {"server"}, {"schedule"}, {"app"},
		},
		reason: "validators work only with domain/model and request types",
	},
	{
		id:    "GID-224",
		scope: []string{"event"},
		banned: [][]string{
			{"dal"}, {"domain", "service"}, {"domain", "usecase"},
			{"client"}, {"metric"}, {"server"}, {"schedule"}, {"app"},
		},
		reason: "transport works only with domain/model; services and dependencies are injected as interfaces at the consumer",
	},
	{
		id:     "GID-226",
		scope:  []string{"domain"},
		banned: [][]string{{"metric"}},
		reason: "domain receives metrics through an interface; the metric package is wired in app",
	},
	{
		id:     "GID-226",
		scope:  []string{"dal"},
		banned: [][]string{{"metric"}},
		reason: "dal receives metrics through an interface; the metric package is wired in app",
	},
	{
		id:     "GID-228",
		scope:  []string{"domain", "usecase"},
		banned: [][]string{{"client"}},
		reason: "usecase orchestrates services; a client is used by a repository or directly by a service",
	},
	{
		id:     "GID-225",
		scope:  []string{"domain"},
		banned: [][]string{{"app"}, {"server"}, {"schedule"}, {"validate"}},
		reason: "the composition root and transport are leaves; nobody imports them",
	},
	{
		id:     "GID-225",
		scope:  []string{"dal"},
		banned: [][]string{{"app"}, {"server"}, {"schedule"}, {"validate"}},
		reason: "the composition root and transport are leaves; nobody imports them",
	},
	{
		id:     "GID-225",
		scope:  []string{"client"},
		banned: [][]string{{"app"}},
		reason: "the composition root and transport are leaves; nobody imports them",
	},
	{
		id:     "GID-225",
		scope:  []string{"metric"},
		banned: [][]string{{"app"}},
		reason: "the composition root and transport are leaves; nobody imports them",
	},
}

// Analyzer — GID-132/170/172/224…229 rules with default settings.
var Analyzer = NewAnalyzer(Settings{})

// Settings — linter settings from .golangci.yml.
type Settings struct {
	// Disable — GID-IDs of built-in rules that the project deliberately turns off
	// (for example, GID-224 during a transition period).
	Disable []string `json:"disable"`
	// Rules — additional project rules on top of the built-in matrix.
	Rules []RuleSetting `json:"rules"`
}

// RuleSetting — an additional import-direction rule: packages in the
// Scope layer are forbidden to import Banned. Layers are given as
// slash-paths of segments ("domain/service", "dal").
type RuleSetting struct {
	ID     string   `json:"id"`
	Scope  string   `json:"scope"`
	Banned []string `json:"banned"`
	Reason string   `json:"reason"`
}

// NewAnalyzer builds the analyzer for import direction between layers.
func NewAnalyzer(s Settings) *analysis.Analyzer {
	rules := effectiveRules(s)
	return &analysis.Analyzer{
		Name: "gidlayerimports",
		Doc: "GID-132/GID-170/GID-172/GID-224…229: dependency direction " +
			"between layers (isolation matrix: dal/domain/server/schedule/" +
			"validate/event/client/metric/app)",
		Run: func(pass *analysis.Pass) (any, error) {
			return run(pass, rules)
		},
	}
}

// layerRule: packages in scope are forbidden to import banned. id is the ID
// of the rule under which a violation is reported.
type layerRule struct {
	id     string
	scope  []string
	banned [][]string
	reason string
}

// effectiveRules — the built-in matrix minus settings.disable plus
// settings.rules (custom rules are checked after the built-in ones).
func effectiveRules(s Settings) []layerRule {
	disabled := make(map[string]struct{}, len(s.Disable))
	for _, id := range s.Disable {
		disabled[id] = struct{}{}
	}
	rules := make([]layerRule, 0, len(layerRules)+len(s.Rules))
	//nolint:gidallptr // the plugin does not depend on the internal gdhelper library
	for _, rule := range layerRules {
		if _, ok := disabled[rule.id]; ok {
			continue
		}
		rules = append(rules, rule)
	}
	//nolint:gidallptr // the plugin does not depend on the internal gdhelper library
	for _, rs := range s.Rules {
		rule := layerRule{
			id:     rs.ID,
			scope:  segments(rs.Scope),
			reason: rs.Reason,
		}
		for _, b := range rs.Banned {
			if seg := segments(b); len(seg) > 0 {
				rule.banned = append(rule.banned, seg)
			}
		}
		if _, ok := disabled[rule.id]; ok || len(rule.scope) == 0 || len(rule.banned) == 0 {
			continue
		}
		rules = append(rules, rule)
	}
	return rules
}

// segments splits a slash-path of a layer ("domain/service") into segments.
func segments(path string) []string {
	var out []string
	for seg := range strings.SplitSeq(path, "/") {
		if seg != "" {
			out = append(out, seg)
		}
	}
	return out
}

func run(pass *analysis.Pass, rules []layerRule) (any, error) {
	pkgPath := pass.Pkg.Path()
	var scoped []layerRule
	//nolint:gidallptr // the plugin does not depend on the internal gdhelper library
	for _, rule := range rules {
		if pathseg.Contains(pkgPath, rule.scope...) {
			scoped = append(scoped, rule)
		}
	}
	if len(scoped) == 0 {
		return nil, nil
	}
	for _, file := range pass.Files {
		if ast.IsGenerated(file) {
			continue
		}
		checkImports(pass, scoped, file)
	}
	return nil, nil
}

func checkImports(pass *analysis.Pass, rules []layerRule, file *ast.File) {
	for _, imp := range file.Imports {
		path, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			continue
		}
		if !sameModule(pass.Pkg.Path(), path) {
			continue
		}
		reportFirstMatch(pass, rules, imp, path)
	}
}

// reportFirstMatch reports the first matching rule: specific rules come
// before general ones, and a single import does not get duplicate diagnostics.
func reportFirstMatch(pass *analysis.Pass, rules []layerRule, imp *ast.ImportSpec, path string) {
	//nolint:gidallptr // the plugin does not depend on the internal gdhelper library
	for _, rule := range rules {
		for _, banned := range rule.banned {
			if !pathseg.Contains(path, banned...) {
				continue
			}
			pass.Reportf(imp.Pos(),
				"%s: package %q must not import %q. Fix: %s",
				rule.id, pass.Pkg.Path(), path, rule.reason)
			return
		}
	}
}

// sameModule tells whether an import belongs to the same module as the
// importing package: layer bans do not affect third-party libraries
// with client/event/metric segments in their path. The module boundary is
// resolved in priority order — see the package doc comment: /internal/,
// then /pkg/<module>/, then (testdata, non-standard layout) the first path
// segment.
func sameModule(pkgPath, importPath string) bool {
	const internalSeg = "/internal/"
	if module, _, ok := strings.Cut(pkgPath, internalSeg); ok {
		return strings.HasPrefix(importPath, module+internalSeg)
	}
	if module, ok := pkgModuleRoot(pkgPath); ok {
		return importPath == module || strings.HasPrefix(importPath, module+"/")
	}
	return firstSegment(pkgPath) == firstSegment(importPath)
}

// pkgModuleRoot returns the "<prefix>/pkg/<module>" root for a package path
// under the pkg/<module> application-module layout, or ok=false if pkgPath
// has no /pkg/ segment (or nothing follows it).
func pkgModuleRoot(pkgPath string) (string, bool) {
	// The module.md application-module layout marker: pkg/<module>/ repeats
	// the same layered structure (dal/, domain/, server/) as internal/.
	const pkgSeg = "/pkg/"
	prefix, rest, ok := strings.Cut(pkgPath, pkgSeg)
	if !ok || rest == "" {
		return "", false
	}
	module, _, _ := strings.Cut(rest, "/")
	if module == "" {
		return "", false
	}
	return prefix + pkgSeg + module, true
}

func firstSegment(path string) string {
	seg, _, _ := strings.Cut(path, "/")
	return seg
}
