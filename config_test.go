package gidrules

import (
	"slices"
	"strings"
	"testing"

	"github.com/golangci/plugin-module-register/register"
	"gopkg.in/yaml.v3"
)

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
