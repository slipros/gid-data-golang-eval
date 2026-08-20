// Command custom-gcl is a full golangci-lint v2.13.0 with the gid* linters
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
// `--gid-rules-only` selects the second built-in config: the gid rules alone,
// without the ~40 stock linters and the formatter. It is for a repository that
// ALREADY runs a golangci-lint of its own — there the stock set would run
// twice. Measured on lk-api (1681 files): 41.7 s for the full config against
// 1.24 s for this one, for byte-identical gid diagnostics. Check first what the
// repository config actually reports: `run.tests: false` and
// `issues.max-issues-per-linter` routinely make that run cover far less than it
// looks (see gid-golangci-rules.yml).
//
// The golangci-lint version is pinned in go.mod (v2.13.0) — it must match
// the version in .custom-gcl.yml.
package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"

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

	// rulesOnlyFlag selects the second built-in config: the gid rules alone,
	// without the stock linters and the formatter. It is for a repository that
	// already runs a golangci-lint of its own, where those would otherwise run
	// twice. golangci-lint knows nothing about the flag, so it is consumed here
	// and never reaches the command line it parses.
	const rulesOnlyFlag = "--gid-rules-only"

	args, rulesOnly := takeFlag(os.Args, rulesOnlyFlag)
	os.Args = args

	config := gidrules.DefaultConfig()
	if rulesOnly {
		config = gidrules.RulesOnlyConfig()
	}

	if len(os.Args) > 1 && os.Args[1] == printConfigCommand {
		_, err := os.Stdout.Write(config)

		return errors.Wrap(err, "print the built-in config")
	}

	useDefaultConfig(config)

	return commands.Execute(buildInfo())
}

// takeFlag removes flag from args, reporting whether it was there. Both the
// bare form and --flag=true are accepted; --flag=false is a way to ask for the
// default config explicitly.
func takeFlag(args []string, flag string) (rest []string, found bool) {
	rest = make([]string, 0, len(args))

	for _, arg := range args {
		switch arg {
		case flag, flag + "=true":
			found = true
		case flag + "=false":
			found = false
		default:
			rest = append(rest, arg)
		}
	}

	return rest, found
}

// buildInfo fills the version golangci-lint reports in `custom-gcl version` —
// and, more importantly, uses as the binary salt of its cache (initHashSalt).
// Cached data is keyed by that salt, so a constant version would keep serving
// facts computed by an older build of the rules.
//
// A binary built from a committed tree is identified by its module version:
// the tag it was installed by, or the pseudo-version of the commit it was built
// from — either way a new commit is a new salt. A build with uncommitted
// changes reports develVersion instead, because the version Go derives for it
// (`v1.2.3+dirty`) is the same for every edit of the working tree; golangci-lint
// answers develVersion by hashing the executable, which is the only key that
// changes on every rebuild of unreleased rules.
func buildInfo() commands.BuildInfo {
	// develVersion — the version golangci-lint treats as "not a release": for
	// it, and only for it, the cache salt falls back to hashing the executable.
	const develVersion = "(devel)"

	// unknownValue — the placeholder for build metadata a binary built outside
	// a repository has no source for.
	const unknownValue = "(unknown)"

	info := commands.BuildInfo{
		Version:   develVersion,
		Commit:    unknownValue,
		Date:      unknownValue,
		GoVersion: runtime.Version(),
	}

	raw, ok := debug.ReadBuildInfo()
	if !ok {
		return info
	}

	if raw.GoVersion != "" {
		info.GoVersion = raw.GoVersion
	}

	var modified bool

	for i := range raw.Settings {
		setting := &raw.Settings[i]

		switch setting.Key {
		case "vcs.revision":
			info.Commit = setting.Value
		case "vcs.time":
			info.Date = setting.Value
		case "vcs.modified":
			modified = setting.Value == "true"
		}
	}

	if modified {
		info.Commit += " (modified)"

		return info
	}

	if raw.Main.Version != "" && raw.Main.Version != develVersion {
		info.Version = raw.Main.Version
	}

	return info
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
