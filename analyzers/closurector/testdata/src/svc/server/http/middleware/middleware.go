// Eval for GID-273, non-applicability: the transport layer is outside the
// rule. A middleware factory is named after the ecosystem it plugs into
// (chi/middleware.RequestID), and a module publishes Router() by the
// composition-root convention — renaming those would diverge from the library
// convention, not fix a defect.
package middleware

// RequestID — the chi middleware shape: a closure builder outside /domain/**.
func RequestID(header string) func(string) string {
	return func(id string) string {
		if id == "" {
			return header
		}

		return id
	}
}
