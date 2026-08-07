package gidrules

import (
	"strings"
	"testing"

	"github.com/golangci/plugin-module-register/register"
	"gopkg.in/yaml.v3"
)

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
