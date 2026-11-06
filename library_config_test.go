package gidrules

import (
	"path/filepath"
	"testing"

	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis/analysistest"
)

const (
	corporatePostgres = "gitlab.gid.team/gid-data/tech/golang/libs/postgres.git"
	forkPostgres      = "fork.example/postgres"
	librariesSetting  = "libraries"
)

func TestLibraryOverrides(t *testing.T) {
	cases := []struct {
		name    string
		linter  string
		fixture string
		from    string
		to      string
	}{
		{"banned symbol", "gidbansymbol", "banned", corporatePostgres, forkPostgres},
		{"helper diagnostic", "gidallptr", "iterate", "gitlab.gid.team/gid-data/tech/golang/libs/helper.git", "fork.example/helper"},
		{"named SQL arguments", "gidcodepolicy", "namedargs/dal/repository", corporatePostgres, forkPostgres},
		{"validator", "gidvalidator", "validated/validate", "github.com/raoptimus/validator.go/v2", "fork.example/validator"},
	}

	//nolint:gidallptr // the standalone plugin does not depend on the corporate helper library
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			factory, err := register.GetPlugin(tc.linter)
			if err != nil {
				t.Fatal(err)
			}

			plugin, err := factory(map[string]any{librariesSetting: map[string]string{tc.from: tc.to}})
			if err != nil {
				t.Fatal(err)
			}

			analyzers, err := plugin.BuildAnalyzers()
			if err != nil {
				t.Fatal(err)
			}

			analysistest.Run(t, filepath.Join(analysistest.TestData(), "libraries"), analyzers[0], tc.fixture)
		})
	}
}

func TestLibraryOverrideSettingsValidation(t *testing.T) {
	factory, err := register.GetPlugin("gidbansymbol")
	if err != nil {
		t.Fatal(err)
	}

	for _, settings := range []map[string]any{
		{librariesSetting: "invalid"},
		{librariesSetting: map[string]string{"": forkPostgres}},
		{"unknown-setting": true},
	} {
		if _, err := factory(settings); err == nil {
			t.Errorf("accepted invalid settings %v", settings)
		}
	}
}
