// Non-applicability: flag names listed in settings.exclude are not flagged
// by GID-239 (GID-238 naming is unaffected by the exclusion list).
package app

import (
	cli "github.com/urfave/cli/v3"
)

// "legacy-mode" is in settings.exclude — no GID-239 diagnostic even though
// neither Required nor Value is set.
var legacyModeFlag = &cli.BoolFlag{
	Name: "legacy-mode",
}

// Not excluded — still flagged.
var missingFlag = &cli.StringFlag{ // want `GID-239: flag "db-host" has neither Required nor a default Value\. Fix: add Required: true \(a flag consumed by wiring must not silently zero-value\) or set an explicit default Value`
	Name: "db-host",
}
