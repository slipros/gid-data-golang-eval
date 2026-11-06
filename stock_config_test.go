package gidrules

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/bombsimon/wsl/v5"
	"github.com/leonklingele/grouper/pkg/analyzer"
	"golang.org/x/tools/go/analysis/analysistest"
	"gopkg.in/yaml.v3"
)

func TestStockReviewRules(t *testing.T) {
	var config struct {
		Linters struct {
			Settings struct {
				WSL struct {
					Default           string   `yaml:"default"`
					Enable            []string `yaml:"enable"`
					AllowFirstInBlock bool     `yaml:"allow-first-in-block"`
				} `yaml:"wsl_v5"`
				Grouper struct {
					SingleConst bool `yaml:"const-require-single-const"`
				} `yaml:"grouper"`
			} `yaml:"settings"`
		} `yaml:"linters"`
	}

	if err := yaml.Unmarshal(DefaultConfig(), &config); err != nil {
		t.Fatal(err)
	}

	enabled := enabledLinters(t, DefaultConfig(), "default")
	for _, name := range []string{"wsl_v5", "grouper"} {
		if !slices.Contains(enabled, name) {
			t.Errorf("the shipped config must enable %s", name)
		}
	}

	t.Run("spacing", func(t *testing.T) {
		settings := config.Linters.Settings.WSL
		cfg, err := wsl.NewWithChecks(settings.Default, settings.Enable, nil)
		if err != nil {
			t.Fatal(err)
		}

		cfg.AllowFirstInBlock = settings.AllowFirstInBlock
		analysistest.Run(t, filepath.Join(analysistest.TestData(), "stock"), wsl.NewAnalyzer(cfg), "spacing")
	})

	t.Run("grouping", func(t *testing.T) {
		if !config.Linters.Settings.Grouper.SingleConst {
			t.Fatal("the shipped config must group package constants")
		}

		a := analyzer.New()
		if err := a.Flags.Set("const-require-single-const", "true"); err != nil {
			t.Fatal(err)
		}

		analysistest.Run(t, filepath.Join(analysistest.TestData(), "stock"), a, "grouping")
	})
}
