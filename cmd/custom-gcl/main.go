// Command custom-gcl is a full golangci-lint v2.9.0 with the gid* linters
// of this repository built in.
//
// It is an alternative to building via `golangci-lint custom` (.custom-gcl.yml):
// the binary is installed directly and does not require cloning golangci-lint —
//
//	go install github.com/slipros/gid-data-golang-eval/cmd/custom-gcl@latest
//
// Usage is identical to regular golangci-lint (a .golangci.yml with the gid*
// linters enabled is all you need):
//
//	custom-gcl run ./...
//
// The golangci-lint version is pinned in go.mod (v2.9.0) — it must match
// the version in .custom-gcl.yml.
package main

import (
	"fmt"
	"os"

	"github.com/golangci/golangci-lint/v2/pkg/commands"
	"github.com/golangci/golangci-lint/v2/pkg/exitcodes"

	// Registers all gid* linters via the gidrules package init().
	// The binary's build entry point must import the root package —
	// the same contract as the generated `golangci-lint custom`.
	//nolint:gidupwardimport // the plugin composition root imports the root per the plugin system contract
	_ "github.com/slipros/gid-data-golang-eval"
)

func main() {
	info := commands.BuildInfo{
		Version:   "custom-gcl (gid-data-golang-eval)",
		Commit:    "(see go module version)",
		Date:      "(unknown)",
		GoVersion: "unknown",
	}
	if err := commands.Execute(info); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "The command is terminated due to an error: %v\n", err)
		os.Exit(exitcodes.Failure)
	}
}
