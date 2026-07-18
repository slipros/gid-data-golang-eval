// Eval boundary GID-163: a server/http segment nested under another layer
// (client/x) is NOT the http layer itself — the layer is anchored to the
// module root (pathseg.HasLayer), so a plain golang handler here must NOT be
// flagged, unlike svc/server/http/handler itself.
package handler

import "net/http"

// Would be flagged if the layer segment were matched anywhere in the path
// (pathseg.Contains) instead of being anchored to the module root.
func Get(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
