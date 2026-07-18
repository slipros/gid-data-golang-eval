// Boundary: this package lives at svc/job/server/http/handler — the
// server/http segments are nested under the job layer, not anchored right
// after the module root, so GID-162 does not apply here even though the
// shapes below look exactly like the flagged ones in
// svc/server/http/handler. A substring-style Contains(path,"server","http")
// check would false-positive; the layer-anchored check (HasLayer) must stay
// clean.
package handler

import "net/http"

type Snapshot struct{}

func (h *Snapshot) handleError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func (h *Snapshot) Get(w http.ResponseWriter, r *http.Request) error {
	return nil
}
