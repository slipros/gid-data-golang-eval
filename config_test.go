package gidrules

import (
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/golangci/plugin-module-register/register"
	"gopkg.in/yaml.v3"
)

// notShipped — registered linters deliberately kept out of the built-in
// configs. A rule parked here needs a reason: the default is that a registered
// rule ships, because a rule nobody runs is a rule that does not exist.
var notShipped = map[string]string{}

// enabledLinters reads linters.enable out of a built-in config.
func enabledLinters(t *testing.T, config []byte, name string) []string {
	t.Helper()

	var parsed struct {
		Linters struct {
			Enable []string `yaml:"enable"`
		} `yaml:"linters"`
	}

	if err := yaml.Unmarshal(config, &parsed); err != nil {
		t.Fatalf("%s is not valid YAML: %v", name, err)
	}

	return parsed.Linters.Enable
}

// gidOnly keeps the gid* rules of a linter list.
func gidOnly(names []string) []string {
	out := make([]string, 0, len(names))

	for _, name := range names {
		if strings.HasPrefix(name, "gid") {
			out = append(out, name)
		}
	}

	return out
}

// TestDefaultConfigEnablesRegisteredLinters — the config is now embedded in the
// binary, so it can no longer be edited independently of the linters compiled
// into it. A gid* linter listed in `linters.enable` but never registered aborts
// every run with `plugin "gidxxx" not found` — this test is that gate.
func TestDefaultConfigEnablesRegisteredLinters(t *testing.T) {
	var config struct {
		Linters struct {
			Enable []string `yaml:"enable"`
		} `yaml:"linters"`
	}

	if err := yaml.Unmarshal(DefaultConfig(), &config); err != nil {
		t.Fatalf("the built-in config is not valid YAML: %v", err)
	}

	enabled := config.Linters.Enable
	if len(enabled) == 0 {
		t.Fatal("the built-in config enables no linters")
	}

	for _, name := range enabled {
		if !strings.HasPrefix(name, "gid") {
			continue // a stock golangci-lint linter, not ours to register
		}

		if _, err := register.GetPlugin(name); err != nil {
			t.Errorf("the built-in config enables %q, but no such plugin is registered in plugin.go — "+
				"every run of the binary would abort. Fix: register the linter or drop it from gid-golangci.yml",
				name)
		}
	}
}

// registeredPlugins reads the gid* linter names out of plugin.go — the
// register package keeps its map private, so the registration source itself is
// the list. A name only reaches a user of the binary through a config, so this
// is the one place that knows what the binary *could* run.
func registeredPlugins(t *testing.T) []string {
	t.Helper()

	src, err := os.ReadFile("plugin.go")
	if err != nil {
		t.Fatalf("cannot read plugin.go: %v", err)
	}

	matches := regexp.MustCompile(`register\.Plugin\("(gid[a-z]+)"`).FindAllStringSubmatch(string(src), -1)
	if len(matches) == 0 {
		t.Fatal("no register.Plugin(\"gid…\") calls found in plugin.go — the pattern went stale")
	}

	out := make([]string, 0, len(matches))
	for _, m := range matches {
		out = append(out, m[1])
	}

	return out
}

// TestBuiltinConfigsShipEveryRegisteredLinter closes the direction the other
// two gates leave open. TestDefaultConfigEnablesRegisteredLinters checks that
// everything *enabled* is registered, and TestRulesOnlyConfigMatchesDefault
// compares the two built-in configs *to each other* — so a linter registered in
// plugin.go and enabled only in the repository's own .golangci.yml passes both:
// it lints this repository via `make lint-fast` and silently never runs for
// anybody using the binary.
//
// That is not hypothetical. GID-246 (gidapproot) was registered and enabled in
// .golangci.yml on the day it was written and never added to either built-in
// config; it produced zero diagnostics on advertising-api, which holds eleven
// adapter structs, and the gap went unnoticed until 2026-08-13 — together with
// nine other rules in the same state.
func TestBuiltinConfigsShipEveryRegisteredLinter(t *testing.T) {
	configs := map[string][]string{
		"gid-golangci.yml":       enabledLinters(t, DefaultConfig(), "gid-golangci.yml"),
		"gid-golangci-rules.yml": enabledLinters(t, RulesOnlyConfig(), "gid-golangci-rules.yml"),
	}

	for _, name := range registeredPlugins(t) {
		if reason, parked := notShipped[name]; parked {
			t.Logf("%s is deliberately not shipped: %s", name, reason)

			continue
		}

		for config, enabled := range configs {
			if !slices.Contains(enabled, name) {
				t.Errorf("plugin.go registers %q, but %s does not enable it — the rule ships in the binary "+
					"and never runs. Fix: add it to %s (RULES.md, step 5 of the process), or park it in "+
					"notShipped with a reason", name, config, config)
			}
		}
	}
}

// TestRulesOnlyConfigMatchesDefault gates the second built-in config
// (--gid-rules-only) against the first: it is the same ruleset with the stock
// linters and the formatter taken out, so a rule added to gid-golangci.yml and
// forgotten here would silently not run for every agent using that flag.
func TestRulesOnlyConfigMatchesDefault(t *testing.T) {
	// keptStock — the stock linters the rules-only config carries anyway,
	// because a service .golangci.yml almost never enables them: depguard is
	// the uuid-fork ban (GID-137), musttag and perfsprint are GID-208.
	keptStock := []string{"depguard", "musttag", "perfsprint"}

	full := enabledLinters(t, DefaultConfig(), "gid-golangci.yml")
	rules := enabledLinters(t, RulesOnlyConfig(), "gid-golangci-rules.yml")

	if len(rules) == 0 {
		t.Fatal("the rules-only config enables no linters")
	}

	wantGID := gidOnly(full)
	gotGID := gidOnly(rules)

	for _, name := range wantGID {
		if !slices.Contains(gotGID, name) {
			t.Errorf("gid-golangci.yml enables %q, gid-golangci-rules.yml does not — an agent running with "+
				"--gid-rules-only would not get that rule. Fix: add it to gid-golangci-rules.yml", name)
		}
	}

	for _, name := range gotGID {
		if !slices.Contains(wantGID, name) {
			t.Errorf("gid-golangci-rules.yml enables %q, which gid-golangci.yml does not — the two configs "+
				"disagree about the ruleset. Fix: enable it in gid-golangci.yml too, or drop it here", name)
		}
	}

	for _, name := range rules {
		if strings.HasPrefix(name, "gid") || slices.Contains(keptStock, name) {
			continue
		}

		t.Errorf("gid-golangci-rules.yml enables the stock linter %q — the whole point of this config is that "+
			"the repository's own golangci-lint owns the stock set. Fix: drop it, or add it to keptStock with "+
			"a reason", name)
	}

	for _, name := range keptStock {
		if !slices.Contains(rules, name) {
			t.Errorf("gid-golangci-rules.yml does not enable %q, which no service config enables either — "+
				"the rule it implements would run nowhere. Fix: enable it", name)
		}
	}
}
