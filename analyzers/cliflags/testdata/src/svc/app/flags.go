// Eval GID-238/GID-239: urfave/cli/v3 flag literal hygiene.
package app

import (
	cli "github.com/urfave/cli/v3"
)

// --- GID-238 positive class: bad Name casing ---

var camelCaseFlag = &cli.StringFlag{
	Name:     "myFlag", // want `GID-238: cli flag name "myFlag" must be kebab-case`
	Required: true,
}

var snakeCaseFlag = &cli.StringFlag{
	Name:     "my_flag", // want `GID-238: cli flag name "my_flag" must be kebab-case`
	Required: true,
}

// --- GID-238 positive class: bad env var casing, both v2 and v3 styles ---

var badEnvVarsSliceFlag = &cli.StringFlag{
	Name:     "db-url",
	Required: true,
	EnvVars:  []string{"db-url"}, // want `GID-238: env var "db-url" must be UPPER_SNAKE_CASE`
}

var badEnvVarsCamelFlag = &cli.StringFlag{
	Name:     "db-url",
	Required: true,
	EnvVars:  []string{"dbUrl"}, // want `GID-238: env var "dbUrl" must be UPPER_SNAKE_CASE`
}

var badSourcesFlag = &cli.StringFlag{
	Name:     "db-url",
	Required: true,
	Sources:  cli.EnvVars("db-url"), // want `GID-238: env var "db-url" must be UPPER_SNAKE_CASE`
}

// --- GID-238 negative class: correctly-cased flags ---

var goodKebabFlag = &cli.StringFlag{
	Name:     "db-url",
	Required: true,
	EnvVars:  []string{"DB_URL"},
}

var goodSourcesFlag = &cli.StringFlag{
	Name:     "db-url",
	Required: true,
	Sources:  cli.EnvVars("DB_URL"),
}

// --- GID-238 boundary class: single-word names need no separator ---

var singleWordFlag = &cli.IntFlag{
	Name:     "port",
	Required: true,
	EnvVars:  []string{"PORT"},
}

// --- GID-238/GID-239 non-applicability class: not a cli flag type ---

// Config looks like a flag (has a Name field) but is not declared in the
// cli package — neither rule applies to it.
type Config struct {
	Name string
}

var notAFlag = Config{
	Name: "my_config",
}

// --- GID-239 positive class: neither Required nor Value ---

var missingBoth = &cli.StringFlag{ // want `GID-239: flag "db-host" has neither Required nor a default Value — a flag consumed by wiring must be required or carry a default`
	Name: "db-host",
}

// --- GID-239 negative class: satisfied by Required or by a default Value ---

var hasRequired = &cli.StringFlag{
	Name:     "db-host",
	Required: true,
}

var hasValue = &cli.StringFlag{
	Name:  "db-host",
	Value: "localhost",
}

// --- GID-239 boundary class: an explicit zero Value still counts as a default ---

var zeroValue = &cli.IntFlag{
	Name:  "retry-count",
	Value: 0,
}

// --- GID-239 boundary class: Name is not a string literal — the diagnostic
// still fires (Required/Value are still absent) but falls back to "<flag>" ---

func dynamicName(name string) *cli.StringFlag {
	return &cli.StringFlag{ // want `GID-239: flag "<flag>" has neither Required nor a default Value — a flag consumed by wiring must be required or carry a default`
		Name: name,
	}
}
