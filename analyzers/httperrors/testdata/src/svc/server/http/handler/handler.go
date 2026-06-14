// Eval for GID-162 (an http handler handles its errors itself).
package handler

import "net/http"

type Snapshot struct{}

// --- Positive: an error-handling super-method ---

func (h *Snapshot) handleError(w http.ResponseWriter, err error) { // want `GID-162: "handleError" is a forbidden error-handling super-method\. Fix: handle errors inside each http handler`
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

// Boundary case: a package-level super-function.
func writeError(w http.ResponseWriter, status int, err error) { // want `GID-162: "writeError" is a forbidden error-handling super-method`
	http.Error(w, err.Error(), status)
}

// --- Positive: the handler returns an error outward ---

func (h *Snapshot) Get(w http.ResponseWriter, r *http.Request) error { // want `GID-162: http handler "Get" must not return error\. Fix: handle the error in place`
	return nil
}

// --- Negative: the handler handles the error inside ---

func (h *Snapshot) List(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// --- Inapplicable: functions without ResponseWriter ---

func convert(err error) string { return err.Error() }
