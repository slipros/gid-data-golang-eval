// Package handler is in scope (the package path ends with the "handler" segment).
package handler

import "svc/extlib"

// LocalOptions is declared in this package — embedding it is a violation.
type LocalOptions struct {
	Verbose bool
}

type Handler struct {
	LocalOptions    // want `GID-152: embedding LocalOptions is forbidden`
	opts LocalOptions

	// extlib.Options comes from another package: even embedded or exported it
	// is never flagged — it cannot be fixed here.
	extlib.Options
	Ext extlib.Options
}
