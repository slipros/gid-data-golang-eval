// Eval for GID-163 (data-response.go instead of plain handlers).
package handler

import "net/http"

type Snapshot struct{}

// --- Positive: a plain golang handler ---

func (h *Snapshot) Get(w http.ResponseWriter, r *http.Request) { // want `GID-163: "Get" is a plain golang handler, which is forbidden\. Fix: use github\.com/raoptimus/data-response\.go/v2 \(exceptions: nolint or settings\.exclude\)`
	w.WriteHeader(http.StatusOK)
}

// Edge case: a package-level plain handler.
func List(w http.ResponseWriter, r *http.Request) { // want `GID-163: "List" is a plain golang handler, which is forbidden`
	w.WriteHeader(http.StatusOK)
}

// --- Negative: a different signature — not a plain handler ---

func (h *Snapshot) Convert(r *http.Request) string { return r.URL.Path }

// Not applicable: a ResponseWriter without a Request.
func write(w http.ResponseWriter, body []byte) {
	_, _ = w.Write(body) //nolint // errcheck is outside this eval
}
