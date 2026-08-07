// Command custom-gcl is a full golangci-lint v2.12.2 with the gid* linters
// of this repository built in.
//
// It is an alternative to building via `golangci-lint custom` (.custom-gcl.yml):
// the binary is installed directly and does not require cloning golangci-lint —
//
//	go install github.com/slipros/gid-data-golang-eval/cmd/custom-gcl@latest
//
// Usage is identical to regular golangci-lint:
//
//	custom-gcl run ./...
//
// A repository with its own .golangci.yml is linted by that config, exactly as
// with regular golangci-lint. A repository without one is linted by the gid
// ruleset embedded in the binary (gid-golangci.yml) — nothing to copy next to
// the binary, nothing to keep in sync with it. `--no-config` opts out, and
// `custom-gcl gid-config` prints the built-in config to stdout, so it can be
// used as a starting point for a repository config.
//
// The golangci-lint version is pinned in go.mod (v2.12.2) — it must match
// the version in .custom-gcl.yml.
package main

import (
	"fmt"
	"os"

	"github.com/golangci/golangci-lint/v2/pkg/commands"
	"github.com/golangci/golangci-lint/v2/pkg/exitcodes"
	"github.com/pkg/errors"

	// The binary's build entry point must import the root package — the same
	// contract as the generated `golangci-lint custom`: its init() registers
	// all gid* linters, and it carries the embedded default config.
	//nolint:gidupwardimport // the plugin composition root imports the root per the plugin system contract
	gidrules "github.com/slipros/gid-data-golang-eval"
	"github.com/slipros/gid-data-golang-eval/internal/defaultconfig"
)

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "The command is terminated due to an error: %v\n", err)
		os.Exit(exitcodes.Failure)
	}
}

func run() error {
	// printConfigCommand — the subcommand dumping the built-in config, so a
	// service can start its own .golangci.yml from the exact ruleset the
	// binary runs: custom-gcl gid-config > .golangci.yml
	const printConfigCommand = "gid-config"

	config := gidrules.DefaultConfig()

	if len(os.Args) > 1 && os.Args[1] == printConfigCommand {
		_, err := os.Stdout.Write(config)

		return errors.Wrap(err, "print the built-in config")
	}

	useDefaultConfig(config)

	info := commands.BuildInfo{
		Version:   "custom-gcl (gid-data-golang-eval)",
		Commit:    "(see go module version)",
		Date:      "(unknown)",
		GoVersion: "unknown",
	}

	return commands.Execute(info)
}

// useDefaultConfig rewrites the command line so that the run is made with the
// config built into the binary. A failure to write that config is not fatal:
// it is reported, and the run goes on with the command line as given — plain
// golangci-lint behaviour, without the gid rules.
func useDefaultConfig(config []byte) {
	args, path, err := defaultconfig.Inject(os.Args, config)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "custom-gcl: cannot use the built-in config: %v\n", err)

		return
	}

	if path != "" {
		_, _ = fmt.Fprintf(os.Stderr, "custom-gcl: using the built-in gid config (%s)%s\n", path, ignoredLocalConfig())
	}

	os.Args = args
}

// ignoredLocalConfig — the tail of the notice naming the repository config the
// run is not using. Regular golangci-lint would have read that file, so its
// silent absence from the run is exactly what needs saying out loud.
func ignoredLocalConfig() string {
	local := defaultconfig.LocalConfig()
	if local == "" {
		return ""
	}

	return fmt.Sprintf("; %s is ignored — pass --config %s to use it", local, local)
}
