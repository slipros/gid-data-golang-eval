// Eval for settings.exclude GID-163.
package handler

import "net/http"

// Excluded as "Health" — a plain handler is acceptable for a health check.
func Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// Not excluded — reported.
func Ready(w http.ResponseWriter, r *http.Request) { // want `GID-163: "Ready" is a plain golang handler, which is forbidden`
	w.WriteHeader(http.StatusOK)
}
